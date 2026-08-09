package service // 业务逻辑层：流水线第 4 步画面生成

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/ai-comic-generator/server/internal/client/cos"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/storage"
	"github.com/fogleman/gg"
	"github.com/tmc/langchaingo/llms"
)

// ImageGenerator 生图后端抽象（hunyuan / openai-image 等均实现此接口）
type ImageGenerator interface {
	Enabled() bool
	Generate(ctx context.Context, prompt, destPath string) error
}

// SizedImageGenerator 支持请求级尺寸/分辨率覆盖（自定义创作画幅）
type SizedImageGenerator interface {
	ImageGenerator
	GenerateWithSize(ctx context.Context, prompt, destPath, size string) error
}

// ImageService 步骤 4：生图（支持混元 / OpenAI 兼容后端；未启用时生成占位图）
type ImageService struct {
	generators    map[string]ImageGenerator // 生图后端注册表：image_backend 值 -> 对应 generator
	cos           *cos.Client
	store         *storage.Local
	cfg           *config.Config
	llm           llms.Model
	promptBuilder *common.PromptBuilder
}

// NewImageService 创建画面生成服务，generators 由 app.go 按 image_backend 名称注册全部可用后端
func NewImageService(cfg *config.Config, store *storage.Local, generators map[string]ImageGenerator, cosClient *cos.Client, llm llms.Model, promptBuilder *common.PromptBuilder) *ImageService {
	return &ImageService{cfg: cfg, store: store, generators: generators, cos: cosClient, llm: llm, promptBuilder: promptBuilder}
}

// resolveGenerator 按任务选定的 image_backend 取对应生成器；未配置或未启用则返回 nil（走占位图兜底）
func (s *ImageService) resolveGenerator(backend string) ImageGenerator {
	if backend == "" {
		backend = common.ImageBackendHunyuan
	}
	gen, ok := s.generators[backend]
	if !ok || gen == nil || !gen.Enabled() {
		return nil
	}
	return gen
}

// GeneratePanels 为每个分镜格生成图片并写入 state.PanelImages
func (s *ImageService) GeneratePanels(ctx context.Context, state *model.ComicState) error {
	if state.Storyboard == nil || len(state.Storyboard.Panels) == 0 {
		return fmt.Errorf("storyboard empty")
	}
	if err := s.store.EnsureTaskDir(state.TaskID); err != nil {
		return err
	}

	charRef, _ := buildCharacterRef(state.Characters)
	results := make([]model.PanelImageResult, 0, len(state.Storyboard.Panels))
	generator := s.resolveGenerator(state.ImageBackend)

	for _, panel := range state.Storyboard.Panels {
		dest := s.store.PanelPath(state.TaskID, panel.PanelNo)
		dialogue := common.FormatPanelDialogue(panel.Dialogue)
		narration := strings.TrimSpace(panel.Narration)
		hyPrompt := s.buildPanelPrompt(ctx, state.Style, panel.Scene, charRef, panel.ImagePrompt, dialogue, narration, state.CaptionTextMode)

		var genErr error
		if generator != nil {
			genErr = generator.Generate(ctx, hyPrompt, dest)
			if genErr == nil {
				if err := OverlayCaption(dest, dialogue, narration, state.CaptionTextMode); err != nil {
					log.Printf("overlay panel caption failed taskId=%s panel=%d: %v", state.TaskID, panel.PanelNo, err)
				}
			}
		} else {
			log.Printf("image generator disabled, use placeholder panel: taskId=%s panel=%d", state.TaskID, panel.PanelNo)
			genErr = renderPlaceholderPanel(dest, panel.PanelNo, panel.Scene, dialogue, narration)
		}
		if genErr != nil {
			return fmt.Errorf("panel %d: %w", panel.PanelNo, genErr)
		}

		url := s.store.PublicURL(state.TaskID, fmt.Sprintf("panel_%d.png", panel.PanelNo))
		if s.cos.Enabled() {
			cosKey := fmt.Sprintf("comics/%s/panel_%d.png", state.TaskID, panel.PanelNo)
			cosURL, err := s.cos.UploadFile(ctx, cosKey, dest)
			if err != nil {
				log.Printf("cos upload panel failed taskId=%s panel=%d: %v", state.TaskID, panel.PanelNo, err)
			} else {
				url = cosURL
			}
		}
		results = append(results, model.PanelImageResult{
			PanelNo:     panel.PanelNo,
			URL:         url,
			Method:      panelImageMethod(generator != nil),
			ImagePrompt: hyPrompt,
		})
	}

	state.PanelImages = results
	state.Phase = model.ComicPhaseImageGeneration
	return nil
}

// buildPanelPrompt 优先经 LLM 增强（画面不含文字，台词由程序叠加），失败则直接拼装英文 Prompt
func (s *ImageService) buildPanelPrompt(ctx context.Context, style, scene, charRef, imagePrompt, dialogue, narration, captionTextMode string) string {
	base := imagePrompt
	if base == "" {
		base = scene
	}
	meta := s.promptBuilder.BuildPanelImageEnhance(style, scene, charRef, base, dialogue, narration, captionTextMode)
	if s.llm != nil {
		content, err := llms.GenerateFromSinglePrompt(ctx, s.llm, meta)
		if err != nil {
			log.Printf("panel prompt llm enhance failed, use direct prompt: %v", err)
		} else if trimmed := strings.TrimSpace(content); trimmed != "" {
			return common.TruncateHunyuanPrompt(common.SanitizeHunyuanImagePrompt(trimmed))
		}
	}
	return common.TruncateHunyuanPrompt(common.BuildDirectPanelImagePrompt(style, scene, charRef, base, dialogue, narration, captionTextMode))
}

func panelImageMethod(hunyuanOn bool) string {
	if hunyuanOn {
		return "AI_GENERATE"
	}
	return "PLACEHOLDER"
}

func buildCharacterRef(chars []model.ComicCharacter) (string, error) {
	if len(chars) == 0 {
		return "", nil
	}
	parts := make([]string, 0, len(chars))
	for _, c := range chars {
		parts = append(parts, fmt.Sprintf("%s: %s", c.Name, c.Appearance))
	}
	return strings.Join(parts, "; "), nil
}

func renderPlaceholderPanel(path string, _ int, scene, dialogue, narration string) error {
	const w, h = 960, 540 // 16:9 电影比例
	dc := gg.NewContext(w, h)
	dc.SetColor(color.RGBA{240, 240, 250, 255})
	dc.Clear()
	dc.SetColor(color.RGBA{120, 100, 220, 255})
	dc.DrawRectangle(20, 20, float64(w-40), float64(h-40))
	dc.SetLineWidth(4)
	dc.Stroke()
	dc.SetColor(color.Black)
	if err := dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 22); err != nil {
		_ = dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 22)
	}
	captionText := dialogue
	if captionText == "" {
		captionText = narration
	}
	if captionText != "" {
		_ = dc.LoadFontFace("C:/Windows/Fonts/msyhbd.ttc", 28)
		lines := wrapCaptionLines(common.TruncateRunes(captionText, panelCaptionMaxRunes), panelCaptionLineWidth)
		y := 36.0
		for i, line := range lines {
			dc.DrawStringAnchored(line, float64(w/2), y+float64(i)*34, 0.5, 0.5)
		}
	}
	wrap := wordWrap(scene, 18)
	y := 120.0
	for _, line := range wrap {
		dc.DrawStringAnchored(line, float64(w/2), y, 0.5, 0)
		y += 28
	}
	return dc.SavePNG(path)
}

func wordWrap(text string, width int) []string {
	runes := []rune(text)
	if len(runes) <= width {
		return []string{text}
	}
	var lines []string
	for i := 0; i < len(runes); i += width {
		end := i + width
		if end > len(runes) {
			end = len(runes)
		}
		lines = append(lines, string(runes[i:end]))
	}
	return lines
}
