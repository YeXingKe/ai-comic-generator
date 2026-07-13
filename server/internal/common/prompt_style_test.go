package common

import "testing"

func TestGetComicStylePromptAnimal(t *testing.T) {
	got := GetComicStylePrompt(ComicStyleAnimal)
	if got == "" {
		t.Fatal("animal style prompt empty")
	}
	for _, want := range []string{"拟人", "anthropomorphic animal", "顶栏字幕", "flat 2D"} {
		if !containsSubstring(got, want) {
			t.Fatalf("missing %q in animal style prompt", want)
		}
	}
}

func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
