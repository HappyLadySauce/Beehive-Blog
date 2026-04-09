// Package markdownfrontmatter 提供与 cmd/app/routes/archives/import.go 中 splitFrontMatter 一致的
// Markdown 前导 YAML 分割逻辑，供导入、Hexo 同步、导出等共用。
package markdownfrontmatter

import "strings"

// SplitFrontMatter 从原始 Markdown 中分离第一段 YAML Front Matter 与正文。
// 规则与 archives.importOneMarkdown 使用的 splitFrontMatter 保持一致。
func SplitFrontMatter(raw string) (yamlPart string, body string, ok bool) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "\uFEFF")
	if !strings.HasPrefix(s, "---") {
		return "", "", false
	}
	s = strings.TrimPrefix(s, "---")
	s = strings.TrimLeft(s, "\r\n")
	idx := strings.Index(s, "\n---")
	if idx < 0 {
		return "", "", false
	}
	yamlPart = strings.TrimSpace(s[:idx])
	body = strings.TrimLeft(s[idx+4:], "\r\n") // after \n---
	if yamlPart == "" {
		return "", "", false
	}
	return yamlPart, body, true
}
