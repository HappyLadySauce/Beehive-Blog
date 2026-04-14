package pages

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	routehexo "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/hexo"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/slug"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// PageAdmin 管理员独立页面业务。
type PageAdmin struct {
	svc *svc.ServiceContext
}

func newPageAdmin(svcCtx *svc.ServiceContext) *PageAdmin {
	return &PageAdmin{svc: svcCtx}
}

func parsePageStatusFilter(raw string) ([]models.ArticleStatus, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	valid := map[string]models.ArticleStatus{
		"draft":     models.ArticleStatusDraft,
		"published": models.ArticleStatusPublished,
		"archived":  models.ArticleStatusArchived,
		"private":   models.ArticleStatusPrivate,
	}
	var out []models.ArticleStatus
	seen := make(map[models.ArticleStatus]struct{})
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		st, ok := valid[p]
		if !ok {
			return nil, errors.New("invalid status filter")
		}
		if _, dup := seen[st]; dup {
			continue
		}
		seen[st] = struct{}{}
		out = append(out, st)
	}
	return out, nil
}

func (p *PageAdmin) slugTakenByPage(ctx context.Context, s string, excludeID int64) (bool, error) {
	q := p.svc.DB.WithContext(ctx).Model(&models.Page{}).Where("slug = ?", s)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var c int64
	if err := q.Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (p *PageAdmin) publishedArticleSlugConflict(ctx context.Context, s string) (bool, error) {
	var c int64
	err := p.svc.DB.WithContext(ctx).Model(&models.Article{}).
		Where("slug = ? AND status = ? AND deleted_at IS NULL", s, models.ArticleStatusPublished).
		Count(&c).Error
	return c > 0, err
}

func (p *PageAdmin) allocUniquePageSlug(ctx context.Context, base string, excludeID int64) (string, error) {
	if base == "" {
		base = slug.Fallback(time.Now().UnixNano())
	}
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		taken, err := p.slugTakenByPage(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if taken {
			continue
		}
		conflict, err := p.publishedArticleSlugConflict(ctx, candidate)
		if err != nil {
			return "", err
		}
		if conflict {
			continue
		}
		return candidate, nil
	}
	return "", errors.New("could not allocate unique slug")
}

func mapPageToListItem(a *models.Page) v1.AdminPageListItem {
	return v1.AdminPageListItem{
		ID:        a.ID,
		Title:     a.Title,
		Slug:      a.Slug,
		Status:    string(a.Status),
		ViewCount: a.ViewCount,
		IsInMenu:  a.IsInMenu,
		SortOrder: a.SortOrder,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapPageToDetail(a *models.Page) *v1.PageDetailResponse {
	it := mapPageToListItem(a)
	return &v1.PageDetailResponse{
		AdminPageListItem: it,
		Content:           a.Content,
	}
}

// AdminListPages 分页列表（未软删）。
func (p *PageAdmin) AdminListPages(ctx context.Context, req *v1.AdminPageListRequest) (*v1.AdminPageListResponse, int, error) {
	if req == nil {
		req = &v1.AdminPageListRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	statuses, err := parsePageStatusFilter(req.Status)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid status filter")
	}

	baseQ := p.svc.DB.WithContext(ctx).Model(&models.Page{})
	kw := strings.TrimSpace(req.Keyword)
	if kw != "" {
		pat := "%" + kw + "%"
		baseQ = baseQ.Where("title ILIKE ? OR content ILIKE ?", pat, pat)
	}
	if len(statuses) > 0 {
		baseQ = baseQ.Where("status IN ?", statuses)
	}
	var total int64
	if err := baseQ.Count(&total).Error; err != nil {
		klog.ErrorS(err, "AdminListPages count")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	order := "updated_at DESC, id DESC"
	switch req.Sort {
	case "oldest":
		order = "created_at ASC, id ASC"
	case "popular":
		order = "view_count DESC, updated_at DESC"
	}
	findQ := p.svc.DB.WithContext(ctx).Model(&models.Page{}).Order(order)
	if kw != "" {
		pat := "%" + kw + "%"
		findQ = findQ.Where("title ILIKE ? OR content ILIKE ?", pat, pat)
	}
	if len(statuses) > 0 {
		findQ = findQ.Where("status IN ?", statuses)
	}
	var rows []models.Page
	offset := (page - 1) * pageSize
	if err := findQ.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		klog.ErrorS(err, "AdminListPages find")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.AdminPageListItem, 0, len(rows))
	for i := range rows {
		list = append(list, mapPageToListItem(&rows[i]))
	}
	return &v1.AdminPageListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// GetPage 详情（未软删）。
func (p *PageAdmin) GetPage(ctx context.Context, id int64) (*v1.PageDetailResponse, int, error) {
	var row models.Page
	if err := p.svc.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("page not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return mapPageToDetail(&row), http.StatusOK, nil
}

// CreatePage 创建页面。
func (p *PageAdmin) CreatePage(ctx context.Context, req *v1.CreatePageRequest) (*v1.PageDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	st := models.ArticleStatusDraft
	if req.Status != "" {
		st = models.ArticleStatus(req.Status)
	}

	var slugStr string
	if strings.TrimSpace(req.Slug) != "" {
		var ok bool
		slugStr, ok = slug.Normalize(strings.TrimSpace(req.Slug))
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := p.slugTakenByPage(ctx, slugStr, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
		conflict, err := p.publishedArticleSlugConflict(ctx, slugStr)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if conflict {
			return nil, http.StatusConflict, errors.New("slug conflicts with a published article")
		}
	} else {
		base := slug.FromTitle(req.Title)
		var err error
		slugStr, err = p.allocUniquePageSlug(ctx, base, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	pg := &models.Page{
		Title:   req.Title,
		Slug:    slugStr,
		Content: req.Content,
		Status:  st,
	}
	if req.IsInMenu != nil {
		pg.IsInMenu = *req.IsInMenu
	}
	if req.SortOrder != nil {
		pg.SortOrder = *req.SortOrder
	}
	if err := p.svc.DB.WithContext(ctx).Create(pg).Error; err != nil {
		klog.ErrorS(err, "CreatePage")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	routehexo.MaybeSyncPage(p.svc, pg.ID)
	return mapPageToDetail(pg), http.StatusOK, nil
}

// UpdatePage 更新页面。
func (p *PageAdmin) UpdatePage(ctx context.Context, id int64, req *v1.UpdatePageRequest) (*v1.PageDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var row models.Page
	if err := p.svc.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("page not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Status != nil {
		updates["status"] = models.ArticleStatus(*req.Status)
	}
	if req.IsInMenu != nil {
		updates["is_in_menu"] = *req.IsInMenu
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Slug != nil {
		s := strings.TrimSpace(*req.Slug)
		if s == "" {
			return nil, http.StatusBadRequest, errors.New("slug cannot be empty")
		}
		norm, ok := slug.Normalize(s)
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := p.slugTakenByPage(ctx, norm, id)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
		conflict, err := p.publishedArticleSlugConflict(ctx, norm)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if conflict {
			return nil, http.StatusConflict, errors.New("slug conflicts with a published article")
		}
		updates["slug"] = norm
	}
	if len(updates) == 0 {
		return mapPageToDetail(&row), http.StatusOK, nil
	}
	if err := p.svc.DB.WithContext(ctx).Model(&models.Page{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		klog.ErrorS(err, "UpdatePage", "id", id)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	routehexo.MaybeSyncPage(p.svc, id)
	var out models.Page
	if err := p.svc.DB.WithContext(ctx).First(&out, id).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return mapPageToDetail(&out), http.StatusOK, nil
}

// UpdatePageStatus 更新状态。
func (p *PageAdmin) UpdatePageStatus(ctx context.Context, id int64, req *v1.UpdatePageStatusRequest) (*v1.PageDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	st := models.ArticleStatus(req.Status)
	res := p.svc.DB.WithContext(ctx).Model(&models.Page{}).Where("id = ?", id).Update("status", st)
	if res.Error != nil {
		klog.ErrorS(res.Error, "UpdatePageStatus", "id", id)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("page not found")
	}
	routehexo.MaybeSyncPage(p.svc, id)
	return p.GetPage(ctx, id)
}

// DeletePage 软删。
func (p *PageAdmin) DeletePage(ctx context.Context, id int64) (*v1.DeletePageResponse, int, error) {
	res := p.svc.DB.WithContext(ctx).Delete(&models.Page{}, id)
	if res.Error != nil {
		klog.ErrorS(res.Error, "DeletePage", "id", id)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("page not found")
	}
	routehexo.MaybeDeletePage(p.svc, id)
	return &v1.DeletePageResponse{ID: id}, http.StatusOK, nil
}

// ListTrashedPages 回收站列表。
func (p *PageAdmin) ListTrashedPages(ctx context.Context, req *v1.AdminPageListRequest) (*v1.AdminPageListResponse, int, error) {
	if req == nil {
		req = &v1.AdminPageListRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	baseQ := p.svc.DB.WithContext(ctx).Unscoped().Model(&models.Page{}).Where("deleted_at IS NOT NULL")
	kw := strings.TrimSpace(req.Keyword)
	if kw != "" {
		pat := "%" + kw + "%"
		baseQ = baseQ.Where("title ILIKE ? OR content ILIKE ?", pat, pat)
	}
	var total int64
	if err := baseQ.Count(&total).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	order := "deleted_at DESC"
	switch req.Sort {
	case "oldest":
		order = "deleted_at ASC"
	case "popular", "newest":
		order = "updated_at DESC, id DESC"
	}
	findQ := p.svc.DB.WithContext(ctx).Unscoped().Model(&models.Page{}).
		Where("deleted_at IS NOT NULL").Order(order)
	if kw != "" {
		pat := "%" + kw + "%"
		findQ = findQ.Where("title ILIKE ? OR content ILIKE ?", pat, pat)
	}
	var rows []models.Page
	offset := (page - 1) * pageSize
	if err := findQ.Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	list := make([]v1.AdminPageListItem, 0, len(rows))
	for i := range rows {
		list = append(list, mapPageToListItem(&rows[i]))
	}
	return &v1.AdminPageListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, http.StatusOK, nil
}

// RestorePage 从回收站恢复。
func (p *PageAdmin) RestorePage(ctx context.Context, id int64) (*v1.DeletePageResponse, int, error) {
	res := p.svc.DB.WithContext(ctx).Unscoped().Model(&models.Page{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if res.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("page not found in trash")
	}
	routehexo.MaybeSyncPage(p.svc, id)
	return &v1.DeletePageResponse{ID: id}, http.StatusOK, nil
}

// PermanentDeletePage 永久删除。
func (p *PageAdmin) PermanentDeletePage(ctx context.Context, id int64) (*v1.DeletePageResponse, int, error) {
	var n int64
	if err := p.svc.DB.WithContext(ctx).Unscoped().Model(&models.Page{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).Count(&n).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if n == 0 {
		return nil, http.StatusNotFound, errors.New("page not found in trash")
	}
	if err := p.svc.DB.WithContext(ctx).Unscoped().Delete(&models.Page{}, id).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	routehexo.MaybeDeletePage(p.svc, id)
	return &v1.DeletePageResponse{ID: id}, http.StatusOK, nil
}
