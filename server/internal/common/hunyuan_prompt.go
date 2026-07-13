package common

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var hunyuanStripPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bspeech\s*bubble[s]?\b`),
	regexp.MustCompile(`(?i)\b(caption|dialogue|subtitle|text\s*box)[es]?\b`),
	regexp.MustCompile(`(?i)\b(comic\s*)?balloon[s]?\b`),
	regexp.MustCompile(`(?i)\b(with|contains?|include[s]?|show[s]?|display[s]?)\s+(text|words|letters|chinese|english)\b`),
}

// SanitizeHunyuanImagePrompt 去掉易触发乱码气泡/文字的片段
func SanitizeHunyuanImagePrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return prompt
	}
	for _, re := range hunyuanStripPatterns {
		prompt = re.ReplaceAllString(prompt, " ")
	}
	prompt = strings.Join(strings.Fields(prompt), " ")
	return strings.TrimSpace(prompt)
}

// HunyuanMaxPromptRunes 腾讯混元 TextToImageLite Prompt 上限（留余量避免边界报错）
const HunyuanMaxPromptRunes = 250

// TruncateHunyuanPrompt 按 rune 截断生图 Prompt，避免 TextLengthExceed
func TruncateHunyuanPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return prompt
	}
	if utf8.RuneCountInString(prompt) <= HunyuanMaxPromptRunes {
		return prompt
	}
	runes := []rune(prompt)
	return string(runes[:HunyuanMaxPromptRunes])
}

// TruncateRunes 按 rune 截断字符串
func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}
