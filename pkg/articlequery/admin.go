// Package articlequery 提供管理员文章列表查询（含草稿等非发布状态）。
package articlequery

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"gorm.io/gorm"
)

// AdminArticleQuery 构建「未软删」文章查询，可选状态集合（空表示不限制状态）。
func AdminArticleQuery(db *gorm.DB, keyword, categorySlug, author, tagComma string, statuses []models.ArticleStatus) *gorm.DB {
	q := db.Model(&models.Article{}).
		Where("articles.deleted_at IS NULL")

	if len(statuses) > 0 {
		q = q.Where("articles.status IN ?", statuses)
	}

	if kw := strings.TrimSpace(keyword); kw != "" {
		pat := "%" + kw + "%"
		q = q.Where("(articles.title ILIKE ? OR articles.summary ILIKE ? OR articles.content ILIKE ?)", pat, pat, pat)
	}
	if cat := strings.TrimSpace(categorySlug); cat != "" {
		q = q.Joins("JOIN categories ON categories.id = articles.category_id").
			Where("categories.slug = ?", cat)
	}
	if auth := strings.TrimSpace(author); auth != "" {
		q = q.Joins("JOIN users AS authors ON authors.id = articles.author_id").
			Where("authors.username = ?", auth)
	}

	tagPart := strings.TrimSpace(tagComma)
	if tagPart != "" {
		parts := strings.Split(tagPart, ",")
		var slugs []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				slugs = append(slugs, p)
			}
		}
		if len(slugs) > 0 {
			tx := db.Session(&gorm.Session{NewDB: true})
			sub := tx.Table("article_tags").
				Select("article_tags.article_id").
				Joins("JOIN tags ON tags.id = article_tags.tag_id AND tags.slug IN ?", slugs).
				Group("article_tags.article_id").
				Having("COUNT(DISTINCT tags.slug) = ?", len(slugs))
			q = q.Where("articles.id IN (?)", sub)
		}
	}
	return q
}

// ListAdminPage 在已构造好的管理员查询链上分页列出文章。
func ListAdminPage(ctx context.Context, db *gorm.DB, q *gorm.DB, page, pageSize int, sort string) ([]models.Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderExpr := ListOrderExpr(sort)
	offset := (page - 1) * pageSize
	var rows []models.Article
	listQ := q.Session(&gorm.Session{}).
		Preload("Author").Preload("Category").Preload("Tags").
		Order(orderExpr).Offset(offset).Limit(pageSize)
	if err := listQ.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
