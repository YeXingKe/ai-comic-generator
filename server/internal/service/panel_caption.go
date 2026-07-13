package service

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/color"
	_ "image/png"
	"os"
	"strings"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/fogleman/gg"
)

const (
	panelCaptionMaxRunes  = 24
	panelCaptionLineWidth = 12
)

// overlayPanelCaption 在分镜图顶部叠加居中中文台词/旁白（无气泡框，参考公众号条漫顶栏字幕）
func overlayPanelCaption(path, dialogue, narration string) error {
	text := strings.TrimSpace(dialogue)
	if text == "" {
		text = strings.TrimSpace(narration)
	}
	if text == "" {
		return nil
	}

	img, err := loadPanelImage(path)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dc := gg.NewContext(w, h)
	dc.DrawImage(img, 0, 0)

	lines := wrapCaptionLines(common.TruncateRunes(text, panelCaptionMaxRunes), panelCaptionLineWidth)
	if err := drawTopCaption(dc, lines, w, h); err != nil {
		return err
	}
	return dc.SavePNG(path)
}

func loadPanelImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func captionFontCandidates() []string {
	return []string{
		"C:/Windows/Fonts/msyhbd.ttc",
		"C:/Windows/Fonts/msyh.ttc",
		"C:/Windows/Fonts/simhei.ttf",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc",
	}
}

func loadCaptionFont(dc *gg.Context, size float64) error {
	for _, path := range captionFontCandidates() {
		if err := dc.LoadFontFace(path, size); err == nil {
			return nil
		}
	}
	return fmt.Errorf("load caption font: no Chinese font found")
}

func wrapCaptionLines(text string, width int) []string {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	var lines []string
	for i := 0; i < len(runes); i += width {
		end := i + width
		if end > len(runes) {
			end = len(runes)
		}
		lines = append(lines, string(runes[i:end]))
		if len(lines) >= 2 {
			break
		}
	}
	if len(runes) > width*2 {
		lines[1] = string([]rune(lines[1])[:width-1]) + "…"
	}
	return lines
}

func drawTopCaption(dc *gg.Context, lines []string, panelW, panelH int) error {
	if len(lines) == 0 {
		return nil
	}

	fontSize := float64(panelH) * 0.042
	if fontSize < 24 {
		fontSize = 24
	}
	if err := loadCaptionFont(dc, fontSize); err != nil {
		return err
	}

	lineH := fontSize * 1.22
	topPad := float64(panelH) * 0.04
	dc.SetColor(color.Black)
	centerX := float64(panelW) / 2
	textY := topPad + fontSize*0.88
	for i, line := range lines {
		dc.DrawStringAnchored(line, centerX, textY+float64(i)*lineH, 0.5, 0.5)
	}
	return nil
}
