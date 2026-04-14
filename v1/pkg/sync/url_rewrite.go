package sync

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"gorm.io/gorm"
)

// loadSiteURL 读取 general.site_url；无记录时返回空字符串。
// 使用 Find+Limit 而非 Take/First，避免 GORM 将「无行」记为 record not found 并打错误日志。
func loadSiteURL(ctx context.Context, db *gorm.DB) (string, error) {
	if db == nil {
		return "", nil
	}
	var rows []models.Setting
	if err := db.WithContext(ctx).Where(`"key" = ? AND "group" = ?`, "site_url", models.SettingGroupGeneral).Limit(1).Find(&rows).Error; err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return strings.TrimSpace(rows[0].Value), nil
}

// publicUploadsBase 计算 Hexo 导出时附件 URL 应使用的前缀。
// siteURL 非空：TrimRight(siteURL) + "/uploads"；否则与 storageBase 相同（不做语义替换）。
func publicUploadsBase(siteURL, storageBase string) string {
	storageBase = strings.TrimRight(strings.TrimSpace(storageBase), "/")
	if storageBase == "" {
		return ""
	}
	siteURL = strings.TrimSpace(siteURL)
	if siteURL == "" {
		return storageBase
	}
	return strings.TrimRight(siteURL, "/") + "/uploads"
}

// RewriteStorageURLs 将正文中出现的存储前缀替换为公开前缀（仅 Hexo 写出用，不改数据库）。
func RewriteStorageURLs(text, fromBase, toBase string) string {
	fromBase = strings.TrimRight(strings.TrimSpace(fromBase), "/")
	toBase = strings.TrimRight(strings.TrimSpace(toBase), "/")
	if text == "" || fromBase == "" || toBase == "" || fromBase == toBase {
		return text
	}
	out := text
	// 先替换带尾部斜杠，再替换无前缀斜杠，避免路径拼接异常
	out = strings.ReplaceAll(out, fromBase+"/", toBase+"/")
	out = strings.ReplaceAll(out, fromBase, toBase)
	return out
}
