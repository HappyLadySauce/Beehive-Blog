package content

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 公开内容接口业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs content Service.
func NewService(svc *svc.ServiceContext) *Service {
	return &Service{svc: svc}
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

func mapAuthor(u models.User) v1.ArticleAuthorItem {
	return v1.ArticleAuthorItem{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

func mapCategory(c *models.Category) *v1.ArticleCategoryItem {
	if c == nil {
		return nil
	}
	return &v1.ArticleCategoryItem{
		ID:   c.ID,
		Name: c.Name,
		Slug: c.Slug,
	}
}

func mapTags(tags []models.Tag) []v1.ArticleTagItem {
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

func mapArticleListItem(a *models.Article) v1.ArticleListItem {
	item := v1.ArticleListItem{
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
		Author:       mapAuthor(a.Author),
		Category:     mapCategory(a.Category),
		Tags:         mapTags(a.Tags),
	}
	return item
}

// ListArticles 公开文章列表。
func (s *Service) ListArticles(ctx context.Context, req *v1.ArticleListRequest, _ *http.Request) (*v1.ArticleListResponse, int, error) {
	if req == nil {
		req = &v1.ArticleListRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	q := s.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("articles.deleted_at IS NULL").
		Where("articles.status = ?", models.ArticleStatusPublished)

	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		pat := "%" + kw + "%"
		q = q.Where("(articles.title ILIKE ? OR articles.summary ILIKE ? OR articles.content ILIKE ?)", pat, pat, pat)
	}
	if cat := strings.TrimSpace(req.Category); cat != "" {
		q = q.Joins("JOIN categories ON categories.id = articles.category_id").
			Where("categories.slug = ?", cat)
	}
	if author := strings.TrimSpace(req.Author); author != "" {
		q = q.Joins("JOIN users AS authors ON authors.id = articles.author_id").
			Where("authors.username = ?", author)
	}

	tagPart := strings.TrimSpace(req.Tag)
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
			sub := s.svc.DB.WithContext(ctx).Table("article_tags").
				Select("article_tags.article_id").
				Joins("JOIN tags ON tags.id = article_tags.tag_id AND tags.slug IN ?", slugs).
				Group("article_tags.article_id").
				Having("COUNT(DISTINCT tags.slug) = ?", len(slugs))
			q = q.Where("articles.id IN (?)", sub)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		klog.ErrorS(err, "ListArticles count failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	orderExpr := "articles.is_pinned DESC, articles.pin_order DESC, articles.published_at DESC NULLS LAST, articles.id DESC"
	switch req.Sort {
	case "oldest":
		orderExpr = "articles.is_pinned DESC, articles.pin_order DESC, articles.published_at ASC NULLS LAST, articles.id ASC"
	case "popular":
		orderExpr = "articles.is_pinned DESC, articles.pin_order DESC, articles.view_count DESC, articles.id DESC"
	case "newest", "":
	default:
		orderExpr = "articles.is_pinned DESC, articles.pin_order DESC, articles.published_at DESC NULLS LAST, articles.id DESC"
	}

	var rows []models.Article
	listQ := s.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("articles.deleted_at IS NULL").
		Where("articles.status = ?", models.ArticleStatusPublished)
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		pat := "%" + kw + "%"
		listQ = listQ.Where("(articles.title ILIKE ? OR articles.summary ILIKE ? OR articles.content ILIKE ?)", pat, pat, pat)
	}
	if cat := strings.TrimSpace(req.Category); cat != "" {
		listQ = listQ.Joins("JOIN categories ON categories.id = articles.category_id").
			Where("categories.slug = ?", cat)
	}
	if author := strings.TrimSpace(req.Author); author != "" {
		listQ = listQ.Joins("JOIN users AS authors ON authors.id = articles.author_id").
			Where("authors.username = ?", author)
	}
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
			sub := s.svc.DB.WithContext(ctx).Table("article_tags").
				Select("article_tags.article_id").
				Joins("JOIN tags ON tags.id = article_tags.tag_id AND tags.slug IN ?", slugs).
				Group("article_tags.article_id").
				Having("COUNT(DISTINCT tags.slug) = ?", len(slugs))
			listQ = listQ.Where("articles.id IN (?)", sub)
		}
	}

	offset := (page - 1) * pageSize
	if err := listQ.Preload("Author").Preload("Category").Preload("Tags").Order(orderExpr).Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "ListArticles query failed")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	list := make([]v1.ArticleListItem, 0, len(rows))
	for i := range rows {
		list = append(list, mapArticleListItem(&rows[i]))
	}
	return &v1.ArticleListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// GetArticle 公开文章详情。
func (s *Service) GetArticle(ctx context.Context, articleID int64, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	var a models.Article
	err := s.svc.DB.WithContext(ctx).
		Preload("Author").Preload("Category").Preload("Tags").
		Where("id = ? AND deleted_at IS NULL AND status = ?", articleID, models.ArticleStatusPublished).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		klog.ErrorS(err, "GetArticle failed", "id", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	base := mapArticleListItem(&a)
	detail := &v1.ArticleDetailResponse{
		ArticleListItem: base,
		Content:         a.Content,
		Protected:       a.Password != "",
	}

	// 上一篇：发布时间更早的已发布文章
	if a.PublishedAt != nil {
		var prev models.Article
		q := s.svc.DB.WithContext(ctx).Where("deleted_at IS NULL AND status = ?", models.ArticleStatusPublished)
		q = q.Where("(published_at < ? OR (published_at = ? AND id < ?))", a.PublishedAt, a.PublishedAt, a.ID)
		if err := q.Order("published_at DESC, id DESC").Limit(1).First(&prev).Error; err == nil {
			detail.Previous = &v1.ArticleNavItem{ID: prev.ID, Title: prev.Title, Slug: prev.Slug}
		}
		var next models.Article
		q2 := s.svc.DB.WithContext(ctx).Where("deleted_at IS NULL AND status = ?", models.ArticleStatusPublished)
		q2 = q2.Where("(published_at > ? OR (published_at = ? AND id > ?))", a.PublishedAt, a.PublishedAt, a.ID)
		if err := q2.Order("published_at ASC, id ASC").Limit(1).First(&next).Error; err == nil {
			detail.Next = &v1.ArticleNavItem{ID: next.ID, Title: next.Title, Slug: next.Slug}
		}
	}
	return detail, http.StatusOK, nil
}

// RecordArticleView 浏览量 +1。
func (s *Service) RecordArticleView(ctx context.Context, articleID int64, _ *http.Request) (*v1.RecordArticleViewResponse, int, error) {
	res := s.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id = ? AND deleted_at IS NULL AND status = ?", articleID, models.ArticleStatusPublished).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
	if res.Error != nil {
		klog.ErrorS(res.Error, "RecordArticleView update failed", "id", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
	}
	var vc int64
	if err := s.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("id = ?", articleID).Select("view_count").Scan(&vc).Error; err != nil {
		return &v1.RecordArticleViewResponse{ViewCount: 0}, http.StatusOK, nil
	}
	return &v1.RecordArticleViewResponse{ViewCount: vc}, http.StatusOK, nil
}
