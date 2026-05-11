package attachment

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MapDBError maps GORM not-found to the attachment sentinel error.
// MapDBError 将 GORM not-found 映射为附件哨兵错误。
func MapDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// ObjectKeyFor creates a safe date-partitioned object key and sanitized filename.
// ObjectKeyFor 创建安全的按日期分区 object key 与清理后的文件名。
func ObjectKeyFor(purpose string, filename string) (objectKey string, safeName string, err error) {
	safeName = SafeFilename(filename)
	if safeName == "" {
		return "", "", fmt.Errorf("%w: filename is required", ErrInvalid)
	}
	token, err := randomHex(16)
	if err != nil {
		return "", "", fmt.Errorf("generate attachment object key: %w", err)
	}
	ext := strings.ToLower(path.Ext(safeName))
	day := time.Now().UTC().Format("2006/01/02")
	return path.Join(purpose, day, token+ext), safeName, nil
}

// SafeFilename strips path components and unsupported filename characters.
// SafeFilename 去除路径片段与不支持的文件名字符。
func SafeFilename(filename string) string {
	filename = strings.TrimSpace(strings.ReplaceAll(filename, "\\", "/"))
	filename = path.Base(filename)
	filename = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, filename)
	return strings.Trim(filename, ".- _")
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CleanOptional trims an optional string and converts blank values to nil.
// CleanOptional 修剪可选字符串，并将空白值转为 nil。
func CleanOptional(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// DerefString trims an optional string or returns an empty string for nil.
// DerefString 修剪可选字符串；nil 时返回空字符串。
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// UniqueInt64 preserves first-seen order while removing duplicates.
// UniqueInt64 保持首次出现顺序并去重。
func UniqueInt64(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
