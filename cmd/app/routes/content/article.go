package content

import (
	"context"
	"errors"
	"net/http"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/articlequery"
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

	db := s.svc.DB.WithContext(ctx)
	q := articlequery.PublishedArticleQuery(db, req.Keyword, req.Category, req.Author, req.Tag)
	rows, total, err := articlequery.ListPublishedPage(ctx, db, q, page, pageSize, req.Sort)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	list := make([]v1.ArticleListItem, 0, len(rows))
	for i := range rows {
		list = append(list, articlequery.MapListItem(&rows[i]))
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

	base := articlequery.MapListItem(&a)
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
