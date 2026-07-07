package service // 业务逻辑层：流水线第 4 步画面生成

import (
	"context"     // 控制混元 / LLM 请求超时
	"fmt"         // 格式化分镜错误信息
	"image/color" // 占位图背景与边框颜色
	"log"         // 记录混元未启用时的降级日志
	"strings"     // 拼接角色外貌、台词与 Prompt

	"github.com/ai-comic-generator/server/internal/client/hunyuan" // 腾讯混元生图客户端
	"github.com/ai-comic-generator/server/internal/common"         // 生图 Prompt 组装
	"github.com/ai-comic-generator/server/internal/config"         // 全局配置
	"github.com/ai-comic-generator/server/internal/model"          // 流水线状态与分镜图片结果
	"github.com/ai-comic-generator/server/internal/storage"        // 本地分镜图路径与 URL
	"github.com/fogleman/gg"                                       // 绘制占位分镜图
	"github.com/tmc/langchaingo/llms"                              // LLM 增强生图 Prompt
)

// ImageService 步骤 4：混元生图（未启用时生成占位图）
type ImageService struct {
	hunyuan *hunyuan.Client // 混元生图客户端
	store   *storage.Local  // 本地文件存储
	cfg     *config.Config  // 全局配置（预留扩展）
	llm     llms.Model      // 可选：用于 PanelImageEnhancePrompt 增强
}

// NewImageService 创建画面生成服务
func NewImageService(cfg *config.Config, store *storage.Local, hy *hunyuan.Client, llm llms.Model) *ImageService {
	return &ImageService{cfg: cfg, store: store, hunyuan: hy, llm: llm}
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

	for _, panel := range state.Storyboard.Panels {
		dest := s.store.PanelPath(state.TaskID, panel.PanelNo)
		dialogue := strings.Join(panel.Dialogue, " / ")
		hyPrompt := s.buildPanelPrompt(ctx, state.Style, panel.Scene, charRef, panel.ImagePrompt, dialogue, panel.Narration)

		var genErr error
		if s.hunyuan.Enabled() {
			genErr = s.hunyuan.Generate(ctx, hyPrompt, dest)
		} else {
			log.Printf("hunyuan disabled, use placeholder panel: taskId=%s panel=%d", state.TaskID, panel.PanelNo)
			genErr = renderPlaceholderPanel(dest, panel.PanelNo, panel.Scene, dialogue)
		}
		if genErr != nil {
			return fmt.Errorf("panel %d: %w", panel.PanelNo, genErr)
		}

		url := s.store.PublicURL(state.TaskID, fmt.Sprintf("panel_%d.png", panel.PanelNo))
		results = append(results, model.PanelImageResult{
			PanelNo:     panel.PanelNo,
			URL:         url,
			Method:      panelImageMethod(s.hunyuan.Enabled()),
			ImagePrompt: hyPrompt,
		})
	}

	state.PanelImages = results
	state.Phase = model.ComicPhaseImageGeneration
	return nil
}

// buildPanelPrompt 优先经 LLM 增强（含对白气泡），失败则直接拼装英文 Prompt
func (s *ImageService) buildPanelPrompt(ctx context.Context, style, scene, charRef, imagePrompt, dialogue, narration string) string {
	base := imagePrompt
	if base == "" {
		base = scene
	}
	meta := common.BuildPanelImageEnhancePrompt(style, scene, charRef, base, dialogue, narration)
	if s.llm != nil {
		content, err := llms.GenerateFromSinglePrompt(ctx, s.llm, meta)
		if err != nil {
			log.Printf("panel prompt llm enhance failed, use direct prompt: %v", err)
		} else if trimmed := strings.TrimSpace(content); trimmed != "" {
			return trimmed
		}
	}
	return common.BuildDirectPanelImagePrompt(style, scene, charRef, base, dialogue, narration)
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

func renderPlaceholderPanel(path string, panelNo int, scene, dialogue string) error {
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
	dc.DrawStringAnchored(fmt.Sprintf("Panel %d", panelNo), float64(w/2), 60, 0.5, 0.5)
	if dialogue != "" {
		drawPlaceholderBubble(dc, dialogue, float64(w/2)-80, 90)
	}
	wrap := wordWrap(scene, 18)
	y := 180.0
	for _, line := range wrap {
		dc.DrawStringAnchored(line, float64(w/2), y, 0.5, 0)
		y += 28
	}
	return dc.SavePNG(path)
}

func drawPlaceholderBubble(dc *gg.Context, text string, x, y float64) {
	_ = dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 16)
	w, h := dc.MeasureString(text)
	pad := 10.0
	bw, bh := w+pad*2, h+pad*2
	dc.SetColor(color.RGBA{255, 255, 255, 240})
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)
	dc.Fill()
	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)
	dc.Stroke()
	dc.DrawStringAnchored(text, x+bw/2, y+bh/2, 0.5, 0.5)
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
