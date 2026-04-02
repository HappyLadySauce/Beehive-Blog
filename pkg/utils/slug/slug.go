package slug

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var nonSlugRunes = regexp.MustCompile(`[^a-z0-9]+`)

// FromTitle 将标题转为可用于 URL 的 slug 片段（仅 a-z0-9 与连字符）。
func FromTitle(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII {
			b.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	s := nonSlugRunes.ReplaceAllString(b.String(), "-")
	s = strings.Trim(s, "-")
	return s
}

// Normalize 校验并规范化用户提供的 slug：仅允许 a-z0-9 与连字符。
func Normalize(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "", false
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return "", false
	}
	return s, true
}

// Fallback 返回 post-{suffix} 形式占位 slug。
func Fallback(suffix int64) string {
	return fmt.Sprintf("post-%d", suffix)
}
