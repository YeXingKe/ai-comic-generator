package common

import (
	"strings"
	"unicode/utf8"
)

// FormatPanelDialogue 提取分镜台词中的中文对白（去掉「角色名：」前缀）
func FormatPanelDialogue(dialogues []string) string {
	if len(dialogues) == 0 {
		return ""
	}
	parts := make([]string, 0, len(dialogues))
	for _, line := range dialogues {
		if t := stripSpeakerPrefix(line); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func stripSpeakerPrefix(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	for i, r := range line {
		if r != '：' && r != ':' {
			continue
		}
		prefix := strings.TrimSpace(line[:i])
		if prefix != "" && len([]rune(prefix)) <= 16 {
			_, size := utf8.DecodeRuneInString(line[i:])
			return strings.TrimSpace(line[i+size:])
		}
		break
	}
	return line
}
