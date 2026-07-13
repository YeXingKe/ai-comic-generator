package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverlayPanelCaptionPreview(t *testing.T) {
	base := filepath.Join("..", "..", "data", "comics", "d65f45ef-8b5a-40f7-81f7-13de80f45e0f")
	src := filepath.Join(base, "panel_2.png")
	if _, err := os.Stat(src); err != nil {
		t.Skip("sample panel not found")
	}
	out := filepath.Join(base, "caption_panel_2.png")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := overlayPanelCaption(out, "其实核心算法很简单。每天下午六点，准时下班！", ""); err != nil {
		t.Fatal(err)
	}
}
