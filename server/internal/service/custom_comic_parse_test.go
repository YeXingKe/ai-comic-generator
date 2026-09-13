package service

import (
	"strings"
	"testing"
)

func TestParseStructuredBeats(t *testing.T) {
	prompt := `
【本集故事】测试
【分镜】
第1张：深夜书桌，小艾趴桌半闭眼，猫不出场。
第2张：橘老板跳上键盘抱胸鄙视，两人同框。
【对话参考】："嗨"
第3张：小艾抬头辩解。
【后期对白】
1. a
`
	beats, ok := parseStructuredBeats(prompt, 3)
	if !ok {
		t.Fatal("expected structured beats")
	}
	if !strings.Contains(beats[0], "小艾") || !strings.Contains(beats[0], "趴桌") {
		t.Fatalf("beat1=%q", beats[0])
	}
	if !strings.Contains(beats[1], "橘老板") || !strings.Contains(beats[1], "键盘") {
		t.Fatalf("beat2=%q", beats[1])
	}
	if !strings.Contains(beats[2], "辩解") {
		t.Fatalf("beat3=%q", beats[2])
	}
}

func TestParseStructuredBeats_incomplete(t *testing.T) {
	prompt := "第1张：只有一格\n第3张：缺第二格"
	if _, ok := parseStructuredBeats(prompt, 3); ok {
		t.Fatal("expected fail when missing panel 2")
	}
}
