package service // 业务逻辑层：流水线第 5 步排版合成

import (
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"

	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/storage"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

const (
	composePanelWidth  = 960 // 合成单格宽度（16:9）
	composePanelHeight = 540 // 合成单格高度（16:9）
	composePanelGap    = 12  // 格与格之间的间距
	composeTitleHeight = 80  // 顶部标题区高度
	composePadding     = 16  // 画布内边距
)

// ComposeService 步骤 5：imaging 竖向拼接 + gg 绘制标题
type ComposeService struct {
	store *storage.Local
}

// NewComposeService 创建排版合成服务
func NewComposeService(store *storage.Local) *ComposeService {
	return &ComposeService{store: store}
}

// Compose 将各分镜图按 16:9 比例竖直拼接成长图
func (s *ComposeService) Compose(ctx context.Context, state *model.ComicState) error {
	_ = ctx
	if state.Storyboard == nil || len(state.PanelImages) == 0 {
		return fmt.Errorf("panels not ready")
	}

	panelCount := len(state.Storyboard.Panels)
	canvasW := composePanelWidth + composePadding*2
	canvasH := composeTitleHeight + composePadding*2 +
		panelCount*composePanelHeight +
		(panelCount-1)*composePanelGap

	dc := gg.NewContext(canvasW, canvasH)
	dc.SetColor(color.White)
	dc.Clear()

	if state.StoryIdeation != nil && state.StoryIdeation.Title != "" {
		_ = dc.LoadFontFace("C:/Windows/Fonts/msyhbd.ttc", 28)
		dc.SetColor(color.Black)
		dc.DrawStringAnchored(state.StoryIdeation.Title, float64(canvasW)/2, composeTitleHeight/2+8, 0.5, 0.5)
	}

	contentTop := composeTitleHeight + composePadding
	for i, panel := range state.Storyboard.Panels {
		imgPath := s.store.PanelPath(state.TaskID, panel.PanelNo)
		src, err := loadImage(imgPath)
		if err != nil {
			return fmt.Errorf("load panel %d: %w", panel.PanelNo, err)
		}
		resized := imaging.Fit(src, composePanelWidth, composePanelHeight, imaging.Lanczos)
		x := composePadding + (composePanelWidth-resized.Bounds().Dx())/2
		y := contentTop + i*(composePanelHeight+composePanelGap) +
			(composePanelHeight-resized.Bounds().Dy())/2
		dc.DrawImage(resized, x, y)

		if i < panelCount-1 {
			sepY := contentTop + (i+1)*composePanelHeight + i*composePanelGap + composePanelGap/2
			dc.SetColor(color.RGBA{220, 220, 230, 255})
			dc.SetLineWidth(1)
			dc.DrawLine(float64(composePadding), float64(sepY), float64(canvasW-composePadding), float64(sepY))
			dc.Stroke()
		}

		if !hasAIGeneratedPanels(state.PanelImages) {
			drawBubble(dc, panel, float64(composePadding)+20,
				float64(y+resized.Bounds().Dy())-60)
		}
	}

	out := s.store.ComposedPath(state.TaskID)
	if err := dc.SavePNG(out); err != nil {
		return err
	}

	previewURL := s.store.PublicURL(state.TaskID, "composed.png")
	state.ComposedLayout = &model.ComposedLayoutResult{
		Format:     "long_image",
		PreviewURL: previewURL,
		AssetURLs:  collectPanelURLs(state.PanelImages),
		CoverImage: previewURL,
	}
	state.Phase = model.ComicPhaseLayoutCompose
	return nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func drawBubble(dc *gg.Context, panel model.StoryboardPanel, x, y float64) {
	text := ""
	if len(panel.Dialogue) > 0 {
		text = panel.Dialogue[0]
	} else if panel.Narration != "" {
		text = panel.Narration
	}
	if text == "" {
		return
	}
	_ = dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 18)
	w, h := dc.MeasureString(text)
	pad := 12.0
	bw, bh := w+pad*2, h+pad*2
	dc.SetColor(color.RGBA{255, 255, 255, 230})
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)
	dc.Fill()
	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)
	dc.Stroke()
	dc.DrawStringAnchored(text, x+bw/2, y+bh/2, 0.5, 0.5)
}

func hasAIGeneratedPanels(images []model.PanelImageResult) bool {
	for _, img := range images {
		if img.Method == "AI_GENERATE" {
			return true
		}
	}
	return false
}

func collectPanelURLs(images []model.PanelImageResult) []string {
	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}
	return urls
}
