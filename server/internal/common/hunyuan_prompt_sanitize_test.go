package common

import (
	"strings"
	"testing"
)

func TestSanitizeHunyuanImagePrompt(t *testing.T) {
	in := "cartoon style, speech bubble above head, with text on screen"
	got := SanitizeHunyuanImagePrompt(in)
	for _, bad := range []string{"speech bubble", "with text"} {
		if strings.Contains(got, bad) {
			t.Fatalf("still contains %q in %q", bad, got)
		}
	}
}

func TestSanitizeHunyuanImagePrompt_keepsChineseScene(t *testing.T) {
	in := "办公室内，角色对话"
	got := SanitizeHunyuanImagePrompt(in)
	if got != in {
		t.Fatalf("unexpected change: %q", got)
	}
}
