package common

import "testing"

func TestFormatPanelDialogue(t *testing.T) {
	got := FormatPanelDialogue([]string{"小明：今天天气真好", "角色B: 我们去吃饭吧"})
	want := "今天天气真好 我们去吃饭吧"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if FormatPanelDialogue(nil) != "" {
		t.Fatal("empty dialogue should return empty")
	}
}

func TestStripSpeakerPrefix(t *testing.T) {
	if got := stripSpeakerPrefix("旁白文字"); got != "旁白文字" {
		t.Fatalf("got %q", got)
	}
}
