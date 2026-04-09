package sync

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"gorm.io/gorm"
)

// loadSiteURL 读取 general.site_url；无记录时返回空字符串。
func loadSiteURL(ctx context.Context, db *gorm.DB) (string, error) {
	if db == nil {
		return "", nil
	}
	var row models.Setting
	err := db.WithContext(ctx).Where(`"key" = ? AND "group" = ?`, "site_url", models.SettingGroupGeneral).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(row.Value), nil
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
