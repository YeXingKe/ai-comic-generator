package common

import "strings"

// IsBlank 判断字符串是否为空或仅空白。
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Coalesce 若 s 为空串则返回 def（不做 Trim，适合配置缺省）。
func Coalesce(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// CoalesceTrim 若 s 空白则返回 def，否则返回 TrimSpace(s)。
func CoalesceTrim(s, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}

// FirstNonBlank 返回第一个非空白字符串（已 Trim）；都空则返回 ""。
func FirstNonBlank(vals ...string) string {
	for _, v := range vals {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

// RequireNonBlank 必填字符串校验，失败返回带自定义文案的 ErrParams。
func RequireNonBlank(s, message string) error {
	if IsBlank(s) {
		return ErrParams.WithMessage(message)
	}
	return nil
}

// IsNilOrEmpty 切片为 nil 或长度为 0。
func IsNilOrEmpty[S ~[]E, E any](s S) bool {
	return len(s) == 0
}
