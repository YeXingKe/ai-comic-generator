package common

import "testing"

func TestTruncateHunyuanPrompt(t *testing.T) {
	short := "cartoon comic panel, 16:9"
	if got := TruncateHunyuanPrompt(short); got != short {
		t.Fatalf("short prompt changed: %q", got)
	}

	long := ""
	for i := 0; i < 300; i++ {
		long += "a"
	}
	got := TruncateHunyuanPrompt(long)
	if utf8Count(got) != HunyuanMaxPromptRunes {
		t.Fatalf("want %d runes, got %d", HunyuanMaxPromptRunes, utf8Count(got))
	}
}

func utf8Count(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
