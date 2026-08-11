package common

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ai-comic-generator/server/internal/model"
)

const (
	visualAnchorMaxRunes = 48
	charAnchorBudget     = 110 // 在混元 250 上限内，优先留给角色锚点
)

// NormalizeCharacterAnchors 补全 visualAnchor（LLM 未返回时从 appearance 压缩）
func NormalizeCharacterAnchors(chars []model.ComicCharacter) {
	for i := range chars {
		anchor := strings.TrimSpace(chars[i].VisualAnchor)
		if anchor == "" {
			anchor = deriveVisualAnchor(chars[i].Appearance)
		}
		chars[i].VisualAnchor = TruncateRunes(strings.TrimSpace(anchor), visualAnchorMaxRunes)
	}
}

func deriveVisualAnchor(appearance string) string {
	appearance = strings.TrimSpace(appearance)
	if appearance == "" {
		return ""
	}
	// 取前若干字作为兜底锚点
	return TruncateRunes(appearance, visualAnchorMaxRunes)
}

// CharacterVisualAnchor 取角色可用的短锚点
func CharacterVisualAnchor(c model.ComicCharacter) string {
	if a := strings.TrimSpace(c.VisualAnchor); a != "" {
		return TruncateRunes(a, visualAnchorMaxRunes)
	}
	return deriveVisualAnchor(c.Appearance)
}

// BuildCharacterAnchorRef 构建强制注入生图的角色短锚点串（主角优先）
func BuildCharacterAnchorRef(chars []model.ComicCharacter) string {
	if len(chars) == 0 {
		return ""
	}
	ordered := make([]model.ComicCharacter, 0, len(chars))
	rest := make([]model.ComicCharacter, 0, len(chars))
	for _, c := range chars {
		if strings.EqualFold(c.Role, "protagonist") {
			ordered = append(ordered, c)
		} else {
			rest = append(rest, c)
		}
	}
	ordered = append(ordered, rest...)

	parts := make([]string, 0, len(ordered))
	used := 0
	for _, c := range ordered {
		anchor := CharacterVisualAnchor(c)
		if anchor == "" {
			continue
		}
		part := fmt.Sprintf("%s (%s)", strings.TrimSpace(c.Name), anchor)
		need := utf8.RuneCountInString(part) + 2
		if used > 0 && used+need > charAnchorBudget {
			break
		}
		parts = append(parts, part)
		used += need
	}
	if len(parts) == 0 {
		return ""
	}
	return "same characters always: " + strings.Join(parts, "; ")
}

// BuildLookbookPrompt 定妆照生图 Prompt（半身正脸，强调外貌锚点）
func BuildLookbookPrompt(style string, c model.ComicCharacter) string {
	anchor := CharacterVisualAnchor(c)
	if anchor == "" {
		anchor = TruncateRunes(c.Appearance, visualAnchorMaxRunes)
	}
	styleHint := style + " comic style"
	if style == ComicStyleAnimal {
		styleHint = "anthropomorphic animal comic, flat 2D illustration, bold outlines"
	}
	parts := []string{
		"character lookbook portrait",
		"half body",
		"front view",
		"neutral expression",
		"plain light background",
		styleHint,
		"clean line art",
		"consistent character design sheet",
		"no text",
		"no watermark",
		"character: " + strings.TrimSpace(c.Name),
		anchor,
	}
	return TruncateHunyuanPrompt(SanitizeHunyuanImagePrompt(strings.Join(parts, ", ")))
}

// ForceInjectCharacterAnchors 将角色锚点置于 Prompt 头部，截断时优先保留锚点
func ForceInjectCharacterAnchors(prompt, anchors string) string {
	prompt = strings.TrimSpace(SanitizeHunyuanImagePrompt(prompt))
	anchors = strings.TrimSpace(anchors)
	if anchors == "" {
		return TruncateHunyuanPrompt(prompt)
	}
	if prompt == "" {
		return TruncateHunyuanPrompt(anchors)
	}
	// 已含锚点关键词则不再重复前置（仍做长度保护）
	lower := strings.ToLower(prompt)
	if strings.Contains(lower, "same characters always") {
		return TruncateKeepingPrefix(anchors, trimAfterAnchors(prompt), HunyuanMaxPromptRunes)
	}
	return TruncateKeepingPrefix(anchors, prompt, HunyuanMaxPromptRunes)
}

func trimAfterAnchors(prompt string) string {
	idx := strings.Index(strings.ToLower(prompt), "same characters always")
	if idx < 0 {
		return prompt
	}
	// 去掉已有锚点段，避免重复
	rest := prompt
	if end := strings.Index(prompt[idx:], ","); end > 0 {
		rest = strings.TrimSpace(prompt[idx+end+1:])
	} else {
		rest = ""
	}
	return rest
}

// TruncateKeepingPrefix 保证 prefix 优先保留，剩余额度给 rest
func TruncateKeepingPrefix(prefix, rest string, max int) string {
	prefix = strings.TrimSpace(prefix)
	rest = strings.TrimSpace(rest)
	if prefix == "" {
		return TruncateRunes(rest, max)
	}
	prefixRunes := utf8.RuneCountInString(prefix)
	if prefixRunes >= max {
		return TruncateRunes(prefix, max)
	}
	if rest == "" {
		return prefix
	}
	budget := max - prefixRunes - 2 // ", "
	if budget <= 0 {
		return TruncateRunes(prefix, max)
	}
	return prefix + ", " + TruncateRunes(rest, budget)
}
