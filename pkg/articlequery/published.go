// Package articlequery 提供公开文章列表的共用查询与映射，供 content、taxonomy 等复用。
package articlequery

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// PublishedArticleQuery 构建「已发布且未软删」文章查询，支持 keyword、categorySlug、author、tagComma（逗号多标签交集）。
func PublishedArticleQuery(db *gorm.DB, keyword, categorySlug, author, tagComma string) *gorm.DB {
	q := db.Model(&models.Article{}).
		Where("articles.deleted_at IS NULL").
		Where("articles.status = ?", models.ArticleStatusPublished)

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

// ListOrderExpr 与公开文章列表一致的排序表达式。
func ListOrderExpr(sort string) string {
	switch sort {
	case "oldest":
		return "articles.is_pinned DESC, articles.pin_order DESC, articles.published_at ASC NULLS LAST, articles.id ASC"
	case "popular":
		return "articles.is_pinned DESC, articles.pin_order DESC, articles.view_count DESC, articles.id DESC"
	case "newest", "":
		fallthrough
	default:
		return "articles.is_pinned DESC, articles.pin_order DESC, articles.published_at DESC NULLS LAST, articles.id DESC"
	}
}

func timeRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func timeRFC3339Ptr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// MapAuthor 映射作者摘要。
func MapAuthor(u models.User) v1.ArticleAuthorItem {
	return v1.ArticleAuthorItem{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

// MapCategory 映射分类摘要。
func MapCategory(c *models.Category) *v1.ArticleCategoryItem {
	if c == nil {
		return nil
	}
	return &v1.ArticleCategoryItem{
		ID:   c.ID,
		Name: c.Name,
		Slug: c.Slug,
	}
}

// MapTags 映射标签列表。
func MapTags(tags []models.Tag) []v1.ArticleTagItem {
	out := make([]v1.ArticleTagItem, 0, len(tags))
	for i := range tags {
		t := tags[i]
		out = append(out, v1.ArticleTagItem{
			ID:    t.ID,
			Name:  t.Name,
			Slug:  t.Slug,
			Color: t.Color,
		})
	}
	return out
}

// MapListItem 映射文章列表项。
func MapListItem(a *models.Article) v1.ArticleListItem {
	return v1.ArticleListItem{
		ID:           a.ID,
		Title:        a.Title,
		Slug:         a.Slug,
		Summary:      a.Summary,
		CoverImage:   a.CoverImage,
		IsPinned:     a.IsPinned,
		PinOrder:     a.PinOrder,
		ViewCount:    a.ViewCount,
		LikeCount:    a.LikeCount,
		CommentCount: a.CommentCount,
		PublishedAt:  timeRFC3339Ptr(a.PublishedAt),
		UpdatedAt:    timeRFC3339(a.UpdatedAt),
		Author:       MapAuthor(a.Author),
		Category:     MapCategory(a.Category),
		Tags:         MapTags(a.Tags),
	}
}

// ListPublishedPage 在已构造好的查询链上分页列出文章（会再 Preload 关联）。
func ListPublishedPage(ctx context.Context, db *gorm.DB, q *gorm.DB, page, pageSize int, sort string) ([]models.Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		klog.ErrorS(err, "articlequery ListPublishedPage count")
		return nil, 0, errors.New("system error")
	}

	orderExpr := ListOrderExpr(sort)
	offset := (page - 1) * pageSize
	var rows []models.Article
	listQ := q.Session(&gorm.Session{}).
		Preload("Author").Preload("Category").Preload("Tags").
		Order(orderExpr).Offset(offset).Limit(pageSize)
	if err := listQ.Find(&rows).Error; err != nil {
		klog.ErrorS(err, "articlequery ListPublishedPage find")
		return nil, 0, errors.New("system error")
	}
	return rows, total, nil
}
