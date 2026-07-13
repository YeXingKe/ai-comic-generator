package service

import (
	"testing"

	"github.com/fogleman/gg"
)

func TestCaptionFontLoading(t *testing.T) {
	dc := gg.NewContext(100, 100)
	for _, path := range captionFontCandidates() {
		if err := dc.LoadFontFace(path, 28); err == nil {
			t.Logf("loaded font: %s", path)
			return
		}
	}
	t.Fatal("no caption font could be loaded")
}
