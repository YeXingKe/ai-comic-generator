package llmjson

import (
	"encoding/json"
	"strings"
)

// Unmarshal 从 LLM 原始输出中提取 JSON 并反序列化（兼容 ```json 代码块）
func Unmarshal[T any](raw string, target *T) error {
	cleaned := ExtractJSON(raw)
	return json.Unmarshal([]byte(cleaned), target)
}

// ExtractJSON 去掉 markdown 包裹，截取 JSON 对象或数组
func ExtractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	startObj := strings.Index(s, "{")
	startArr := strings.Index(s, "[")
	start := -1
	end := -1
	if startObj >= 0 && (startArr < 0 || startObj < startArr) {
		start = startObj
		end = strings.LastIndex(s, "}")
	} else if startArr >= 0 {
		start = startArr
		end = strings.LastIndex(s, "]")
	}
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
