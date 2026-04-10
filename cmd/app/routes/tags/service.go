package tags

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	apphexo "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/hexo"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/articlequery"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/color"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/slug"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 标签业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs tags Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

const defaultTagColor = "#3B82F6"

func toItem(t *models.Tag) v1.TagListItem {
	return v1.TagListItem{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		Color:        t.Color,
		Description:  t.Description,
		ArticleCount: t.ArticleCount,
		SortOrder:    t.SortOrder,
		CreatedAt:    t.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func normalizeColor(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultTagColor
	}
	if h := color.NormalizeToHex(s); h != "" {
		return h
	}
	return ""
}

func (s *Service) tagSlugTaken(ctx context.Context, sl string, excludeID int64) (bool, error) {
	q := s.svc.DB.WithContext(ctx).Model(&models.Tag{}).Where("slug = ?", sl)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) tagNameTaken(ctx context.Context, name string, excludeID int64) (bool, error) {
	name = strings.TrimSpace(name)
	q := s.svc.DB.WithContext(ctx).Model(&models.Tag{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) allocUniqueTagSlug(ctx context.Context, base string, excludeID int64) (string, error) {
	if base == "" {
		base = slug.Fallback(time.Now().UnixNano())
	}
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		taken, err := s.tagSlugTaken(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func (s *Service) applyTagListQuery(q *gorm.DB, keyword string) *gorm.DB {
	if kw := strings.TrimSpace(keyword); kw != "" {
		pat := "%" + kw + "%"
		q = q.Where("(tags.name ILIKE ? OR tags.slug ILIKE ?)", pat, pat)
	}
	return q
}

func tagOrderClause(sort string) string {
	switch sort {
	case "name":
		return "tags.name ASC, tags.id ASC"
	case "newest":
		return "tags.created_at DESC, tags.id DESC"
	case "count", "":
		fallthrough
	default:
		return "tags.article_count DESC, tags.id DESC"
	}
}

// PublicList 公开标签分页。
func (s *Service) PublicList(ctx context.Context, req *v1.TagListRequest, _ *http.Request) (*v1.TagListResponse, int, error) {
	if req == nil {
		req = &v1.TagListRequest{}
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := s.svc.DB.WithContext(ctx).Model(&models.Tag{})
	q = s.applyTagListQuery(q, req.Keyword)
	order := tagOrderClause(req.Sort)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	var rows []models.Tag
	offset := (page - 1) * pageSize
	if err := q.Order(order).Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.TagListItem, 0, len(rows))
	for i := range rows {
		list = append(list, toItem(&rows[i]))
	}
	return &v1.TagListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// PublicCloud 标签云。
func (s *Service) PublicCloud(ctx context.Context, req *v1.TagCloudRequest, _ *http.Request) (*v1.TagCloudResponse, int, error) {
	if req == nil {
		req = &v1.TagCloudRequest{}
	}
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}
	var rows []models.Tag
	if err := s.svc.DB.WithContext(ctx).Model(&models.Tag{}).
		Order("article_count DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.TagListItem, 0, len(rows))
	for i := range rows {
		list = append(list, toItem(&rows[i]))
	}
	return &v1.TagCloudResponse{Tags: list}, http.StatusOK, nil
}

// PublicDetail 按 slug 返回标签、文章分页与共现标签。
func (s *Service) PublicDetail(ctx context.Context, tagSlug string, req *v1.TagDetailRequest, _ *http.Request) (*v1.TagDetailResponse, int, error) {
	tagSlug = strings.TrimSpace(tagSlug)
	if tagSlug == "" {
		return nil, http.StatusBadRequest, errors.New("invalid slug")
	}
	var t models.Tag
	if err := s.svc.DB.WithContext(ctx).Where("slug = ?", tagSlug).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("tag not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	resp := &v1.TagDetailResponse{TagListItem: toItem(&t)}

	page, pageSize := 1, 10
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.PageSize > 0 {
			pageSize = req.PageSize
		}
	}
	db := s.svc.DB.WithContext(ctx)
	q := articlequery.PublishedArticleQuery(db, "", "", "", tagSlug)
	rows, total, err := articlequery.ListPublishedPage(ctx, db, q, page, pageSize, "newest")
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	list := make([]v1.ArticleListItem, 0, len(rows))
	for i := range rows {
		list = append(list, articlequery.MapListItem(&rows[i]))
	}
	resp.Articles = &v1.ArticleListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	type relRow struct {
		ID           int64
		Name         string
		Slug         string
		Color        string
		Description  string
		ArticleCount int64
		SortOrder    int
		CreatedAt    time.Time
		CoCount      int64
	}
	var rel []relRow
	if err := s.svc.DB.WithContext(ctx).Raw(`
SELECT t.id, t.name, t.slug, t.color, t.description, t.article_count, t.sort_order, t.created_at, COUNT(*) AS co_count
FROM article_tags at1
JOIN article_tags at2 ON at1.article_id = at2.article_id AND at1.tag_id <> at2.tag_id
JOIN tags t ON t.id = at2.tag_id
JOIN articles a ON a.id = at1.article_id AND a.deleted_at IS NULL AND a.status = ?
WHERE at1.tag_id = ?
GROUP BY t.id, t.name, t.slug, t.color, t.description, t.article_count, t.sort_order, t.created_at
ORDER BY co_count DESC
LIMIT 15
`, models.ArticleStatusPublished, t.ID).Scan(&rel).Error; err != nil {
		klog.ErrorS(err, "PublicDetail related tags")
	} else {
		resp.RelatedTags = make([]v1.RelatedTagItem, 0, len(rel))
		for _, r := range rel {
			resp.RelatedTags = append(resp.RelatedTags, v1.RelatedTagItem{
				TagListItem: v1.TagListItem{
					ID:           r.ID,
					Name:         r.Name,
					Slug:         r.Slug,
					Color:        r.Color,
					Description:  r.Description,
					ArticleCount: r.ArticleCount,
					SortOrder:    r.SortOrder,
					CreatedAt:    r.CreatedAt.UTC().Format(time.RFC3339),
				},
				CoCount: r.CoCount,
			})
		}
	}

	return resp, http.StatusOK, nil
}

// AdminList 管理员标签列表。
func (s *Service) AdminList(ctx context.Context, req *v1.AdminTagListRequest, _ *http.Request) (*v1.TagListResponse, int, error) {
	pub := &v1.TagListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Sort:     req.Sort,
	}
	return s.PublicList(ctx, pub, nil)
}

// AdminCreate 创建标签。
func (s *Service) AdminCreate(ctx context.Context, req *v1.CreateTagRequest, _ *http.Request) (*v1.TagListItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	name := strings.TrimSpace(req.Name)
	if taken, err := s.tagNameTaken(ctx, name, 0); err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	} else if taken {
		return nil, http.StatusConflict, errors.New("tag name already exists")
	}

	col := normalizeColor(req.Color)
	if col == "" {
		return nil, http.StatusBadRequest, errors.New("invalid color")
	}

	slugStr := strings.TrimSpace(req.Slug)
	if slugStr != "" {
		var ok bool
		slugStr, ok = slug.Normalize(slugStr)
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := s.tagSlugTaken(ctx, slugStr, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
	} else {
		base := slug.FromTitle(name)
		var err error
		slugStr, err = s.allocUniqueTagSlug(ctx, base, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	t := &models.Tag{
		Name:        name,
		Slug:        slugStr,
		Color:       col,
		Description: req.Description,
		SortOrder:   sortOrder,
	}
	if err := s.svc.DB.WithContext(ctx).Create(t).Error; err != nil {
		klog.ErrorS(err, "AdminCreate tag")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	item := toItem(t)
	return &item, http.StatusOK, nil
}

// AdminUpdate 更新标签。
func (s *Service) AdminUpdate(ctx context.Context, id int64, req *v1.UpdateTagRequest, _ *http.Request) (*v1.TagListItem, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var t models.Tag
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("tag not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, http.StatusBadRequest, errors.New("invalid name")
		}
		if taken, err := s.tagNameTaken(ctx, name, id); err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		} else if taken {
			return nil, http.StatusConflict, errors.New("tag name already exists")
		}
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Color != nil {
		col := normalizeColor(*req.Color)
		if col == "" {
			return nil, http.StatusBadRequest, errors.New("invalid color")
		}
		updates["color"] = col
	}
	if req.Slug != nil {
		sl, ok := slug.Normalize(strings.TrimSpace(*req.Slug))
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := s.tagSlugTaken(ctx, sl, id)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
		updates["slug"] = sl
	}

	wrote := false
	if len(updates) > 0 {
		if err := s.svc.DB.WithContext(ctx).Model(&models.Tag{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			klog.ErrorS(err, "AdminUpdate tag")
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		wrote = true
	}
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if wrote {
		apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	}
	item := toItem(&t)
	return &item, http.StatusOK, nil
}

func (s *Service) countArticlesWithTag(ctx context.Context, tagID int64) (int64, error) {
	var n int64
	err := s.svc.DB.WithContext(ctx).Model(&models.ArticleTag{}).Where("tag_id = ?", tagID).Count(&n).Error
	return n, err
}

// AdminDelete 删除标签；有关联文章时除 force 外拒绝。
func (s *Service) AdminDelete(ctx context.Context, id int64, force bool, _ *http.Request) (*v1.DeleteTagResponse, int, error) {
	nLinks, err := s.countArticlesWithTag(ctx, id)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if nLinks > 0 && !force {
		return nil, http.StatusConflict, errors.New("tag has articles")
	}

	tx := s.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if nLinks > 0 && force {
		if err := tx.Where("tag_id = ?", id).Delete(&models.ArticleTag{}).Error; err != nil {
			tx.Rollback()
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := tx.Delete(&models.Tag{}, id).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	return &v1.DeleteTagResponse{ID: id}, http.StatusOK, nil
}
