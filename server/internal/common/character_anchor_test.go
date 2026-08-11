package common

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ai-comic-generator/server/internal/model"
)

func TestBuildCharacterAnchorRef_PrioritizesProtagonist(t *testing.T) {
	chars := []model.ComicCharacter{
		{Name: "配角", Role: "supporting", VisualAnchor: "blue cat, scarf"},
		{Name: "小明", Role: "protagonist", VisualAnchor: "green frog, yellow hoodie"},
	}
	got := BuildCharacterAnchorRef(chars)
	if !strings.Contains(got, "same characters always") {
		t.Fatalf("missing prefix: %s", got)
	}
	idxProtagonist := strings.Index(got, "小明")
	idxSupport := strings.Index(got, "配角")
	if idxProtagonist < 0 || idxSupport < 0 || idxProtagonist > idxSupport {
		t.Fatalf("protagonist should come first: %s", got)
	}
}

func TestForceInjectCharacterAnchors_KeepsPrefix(t *testing.T) {
	anchors := "same characters always: XiaoMing (green frog, yellow hoodie)"
	rest := strings.Repeat("scene detail words ", 40)
	got := ForceInjectCharacterAnchors(rest, anchors)
	if !strings.HasPrefix(got, anchors) {
		t.Fatalf("anchors not prefixed: %s", got)
	}
	if utf8.RuneCountInString(got) > HunyuanMaxPromptRunes {
		t.Fatalf("exceeds hunyuan limit: %d", utf8.RuneCountInString(got))
	}
}

func TestNormalizeCharacterAnchors_Fallback(t *testing.T) {
	chars := []model.ComicCharacter{
		{Name: "A", Appearance: "一只戴圆框眼镜的绿色青蛙，穿着黄色卫衣"},
	}
	NormalizeCharacterAnchors(chars)
	if chars[0].VisualAnchor == "" {
		t.Fatal("expected fallback visualAnchor")
	}
}
