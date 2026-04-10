package categories

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
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/slug"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 分类业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs categories Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

func toBrief(c *models.Category) v1.CategoryBrief {
	return v1.CategoryBrief{
		ID:           c.ID,
		Name:         c.Name,
		Slug:         c.Slug,
		Description:  c.Description,
		ArticleCount: c.ArticleCount,
		SortOrder:    c.SortOrder,
	}
}

func (s *Service) categoryNameTaken(ctx context.Context, name string, excludeID int64) (bool, error) {
	name = strings.TrimSpace(name)
	q := s.svc.DB.WithContext(ctx).Model(&models.Category{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) categorySlugTaken(ctx context.Context, sl string, excludeID int64) (bool, error) {
	q := s.svc.DB.WithContext(ctx).Model(&models.Category{}).Where("slug = ?", sl)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) allocUniqueCategorySlug(ctx context.Context, base string, excludeID int64) (string, error) {
	if base == "" {
		base = slug.Fallback(time.Now().UnixNano())
	}
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		taken, err := s.categorySlugTaken(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func (s *Service) countArticlesInCategory(ctx context.Context, id int64) (int64, error) {
	var n int64
	err := s.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("category_id = ? AND deleted_at IS NULL", id).Count(&n).Error
	return n, err
}

// PublicList 公开一级分类列表。
func (s *Service) PublicList(ctx context.Context) (*v1.CategoryListResponse, int, error) {
	var rows []models.Category
	if err := s.svc.DB.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		klog.ErrorS(err, "PublicList categories")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.CategoryBrief, 0, len(rows))
	for i := range rows {
		list = append(list, toBrief(&rows[i]))
	}
	return &v1.CategoryListResponse{List: list}, http.StatusOK, nil
}

// PublicDetail 按 slug 查询分类及该分类下已发布文章分页。
func (s *Service) PublicDetail(ctx context.Context, categorySlug string, req *v1.CategoryDetailRequest, _ *http.Request) (*v1.CategoryDetailResponse, int, error) {
	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return nil, http.StatusBadRequest, errors.New("invalid slug")
	}
	var c models.Category
	if err := s.svc.DB.WithContext(ctx).Where("slug = ?", categorySlug).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("category not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	resp := &v1.CategoryDetailResponse{
		CategoryBrief: toBrief(&c),
	}

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
	q := articlequery.PublishedArticleQuery(db, "", categorySlug, "", "")
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
	return resp, http.StatusOK, nil
}

// AdminList 管理员扁平列表。
func (s *Service) AdminList(ctx context.Context, req *v1.AdminCategoryListRequest, _ *http.Request) (*v1.AdminCategoryListResponse, int, error) {
	if req == nil {
		req = &v1.AdminCategoryListRequest{}
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := s.svc.DB.WithContext(ctx).Model(&models.Category{}).Order("sort_order ASC, article_count DESC, id DESC")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	var rows []models.Category
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.CategoryBrief, 0, len(rows))
	for i := range rows {
		list = append(list, toBrief(&rows[i]))
	}
	return &v1.AdminCategoryListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// AdminCreate 创建分类。
func (s *Service) AdminCreate(ctx context.Context, req *v1.CreateCategoryRequest, _ *http.Request) (*v1.CategoryBrief, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if taken, err := s.categoryNameTaken(ctx, req.Name, 0); err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	} else if taken {
		return nil, http.StatusConflict, errors.New("category name already exists")
	}

	slugStr := strings.TrimSpace(req.Slug)
	if slugStr != "" {
		var ok bool
		slugStr, ok = slug.Normalize(slugStr)
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := s.categorySlugTaken(ctx, slugStr, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
	} else {
		base := slug.FromTitle(req.Name)
		var err error
		slugStr, err = s.allocUniqueCategorySlug(ctx, base, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	c := &models.Category{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slugStr,
		Description: req.Description,
		SortOrder:   sortOrder,
	}
	if err := s.svc.DB.WithContext(ctx).Create(c).Error; err != nil {
		klog.ErrorS(err, "AdminCreate category")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	b := toBrief(c)
	return &b, http.StatusOK, nil
}

// AdminUpdate 更新分类。
func (s *Service) AdminUpdate(ctx context.Context, id int64, req *v1.UpdateCategoryRequest, _ *http.Request) (*v1.CategoryBrief, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var c models.Category
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("category not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, http.StatusBadRequest, errors.New("invalid name")
		}
		if taken, err := s.categoryNameTaken(ctx, name, id); err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		} else if taken {
			return nil, http.StatusConflict, errors.New("category name already exists")
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
	if req.Slug != nil {
		sl, ok := slug.Normalize(strings.TrimSpace(*req.Slug))
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := s.categorySlugTaken(ctx, sl, id)
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
		if err := s.svc.DB.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			klog.ErrorS(err, "AdminUpdate category")
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		wrote = true
	}

	if err := s.svc.DB.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if wrote {
		apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	}
	b := toBrief(&c)
	return &b, http.StatusOK, nil
}

// AdminDelete 删除分类；有关联文章时需 force=true。
func (s *Service) AdminDelete(ctx context.Context, id int64, force bool, _ *http.Request) (*v1.DeleteCategoryResponse, int, error) {
	nArt, err := s.countArticlesInCategory(ctx, id)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if nArt > 0 && !force {
		return nil, http.StatusConflict, errors.New("category has articles")
	}

	tx := s.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if nArt > 0 && force {
		if err := tx.Model(&models.Article{}).Where("category_id = ?", id).Update("category_id", nil).Error; err != nil {
			tx.Rollback()
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := tx.Delete(&models.Category{}, id).Error; err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	apphexo.WriteHexoTaxonomyFile(ctx, s.svc)
	return &v1.DeleteCategoryResponse{ID: id}, http.StatusOK, nil
}
