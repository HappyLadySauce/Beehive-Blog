// Package attachmentref 根据正文中的本站附件 URL 维护 article_attachments。
package attachmentref

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"gorm.io/gorm"
)

var urlInText = regexp.MustCompile(`https?://[^\s\])>'"<]+`)

// SyncForArticle 重建某篇文章与附件的关联（先删后插）。
func SyncForArticle(ctx context.Context, db *gorm.DB, articleID int64, content, summary, baseURL string) error {
	if articleID <= 0 || db == nil {
		return nil
	}
	text := content
	if summary != "" {
		text += "\n" + summary
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return nil
	}

	raw := urlInText.FindAllString(text, -1)
	seen := make(map[int64]struct{})
	var ids []int64
	for _, u := range raw {
		u = strings.TrimRight(strings.TrimSpace(u), "/.,;:!?")
		if u == "" || !strings.HasPrefix(u, base) {
			continue
		}
		norm, err := NormalizeURLString(u)
		if err != nil {
			continue
		}
		var aid int64
		err = db.WithContext(ctx).Model(&models.Attachment{}).Where("url = ?", norm).Limit(1).Select("id").Scan(&aid).Error
		if err != nil || aid <= 0 {
			aid = 0
			err = db.WithContext(ctx).Model(&models.Attachment{}).Where("url = ?", u).Limit(1).Select("id").Scan(&aid).Error
			if err != nil || aid <= 0 {
				continue
			}
		}
		if _, ok := seen[aid]; ok {
			continue
		}
		seen[aid] = struct{}{}
		ids = append(ids, aid)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleAttachment{}).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		rows := make([]models.ArticleAttachment, 0, len(ids))
		for _, aid := range ids {
			rows = append(rows, models.ArticleAttachment{
				ArticleID:    articleID,
				AttachmentID: aid,
			})
		}
		return tx.Create(&rows).Error
	})
}

// NormalizeURLString 去掉 fragment/query 并去掉末尾斜杠，供 URL 匹配与替换。
func NormalizeURLString(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	u.Fragment = ""
	u.RawQuery = ""
	out := u.String()
	return strings.TrimRight(out, "/"), nil
}
