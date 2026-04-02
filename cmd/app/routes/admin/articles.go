package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/passwd"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/utils/slug"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ArticleAdmin 管理员文章业务。
type ArticleAdmin struct {
	svc *svc.ServiceContext
}

func newArticleAdmin(svcCtx *svc.ServiceContext) *ArticleAdmin {
	return &ArticleAdmin{svc: svcCtx}
}

func parseRFC3339Ptr(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (a *ArticleAdmin) slugTaken(ctx context.Context, s string, excludeID int64) (bool, error) {
	q := a.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("slug = ? AND deleted_at IS NULL", s)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var c int64
	if err := q.Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (a *ArticleAdmin) allocUniqueSlug(ctx context.Context, base string, excludeID int64) (string, error) {
	if base == "" {
		base = slug.Fallback(time.Now().UnixNano())
	}
	for i := 0; i < 50; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		taken, err := a.slugTaken(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func (a *ArticleAdmin) replaceTags(ctx context.Context, tx *gorm.DB, articleID int64, tagIDs []int64) error {
	if err := tx.WithContext(ctx).Where("article_id = ?", articleID).Delete(&models.ArticleTag{}).Error; err != nil {
		return err
	}
	for _, tid := range tagIDs {
		if tid <= 0 {
			continue
		}
		if err := tx.WithContext(ctx).Create(&models.ArticleTag{ArticleID: articleID, TagID: tid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (a *ArticleAdmin) validateCategory(ctx context.Context, id *int64) error {
	if id == nil {
		return nil
	}
	var c int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Category{}).Where("id = ?", *id).Count(&c).Error; err != nil {
		return err
	}
	if c == 0 {
		return errors.New("category not found")
	}
	return nil
}

func (a *ArticleAdmin) validateTags(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	var c int64
	if err := a.svc.DB.WithContext(ctx).Model(&models.Tag{}).Where("id IN ?", ids).Count(&c).Error; err != nil {
		return err
	}
	if int(c) != len(ids) {
		return errors.New("one or more tags not found")
	}
	return nil
}

func mapArticleToAdminDetail(a *models.Article) *v1.ArticleDetailResponse {
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
		PublishedAt: func() string {
			if a.PublishedAt == nil {
				return ""
			}
			return a.PublishedAt.UTC().Format(time.RFC3339)
		}(),
		UpdatedAt: a.UpdatedAt.UTC().Format(time.RFC3339),
		Author: v1.ArticleAuthorItem{
			ID: a.Author.ID, Username: a.Author.Username, Nickname: a.Author.Nickname, Avatar: a.Author.Avatar,
		},
		Tags: make([]v1.ArticleTagItem, 0, len(a.Tags)),
	}
	if a.Category != nil {
		item.Category = &v1.ArticleCategoryItem{ID: a.Category.ID, Name: a.Category.Name, Slug: a.Category.Slug}
	}
	for _, t := range a.Tags {
		item.Tags = append(item.Tags, v1.ArticleTagItem{ID: t.ID, Name: t.Name, Slug: t.Slug, Color: t.Color})
	}
	return &v1.ArticleDetailResponse{
		ArticleListItem: item,
		Content:         a.Content,
		Protected:       a.Password != "",
	}
}

func (a *ArticleAdmin) loadArticleDetail(ctx context.Context, id int64) (*v1.ArticleDetailResponse, int, error) {
	var art models.Article
	if err := a.svc.DB.WithContext(ctx).Preload("Author").Preload("Category").Preload("Tags").
		Where("id = ? AND deleted_at IS NULL", id).First(&art).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return mapArticleToAdminDetail(&art), http.StatusOK, nil
}

// CreateArticle 创建文章。
func (a *ArticleAdmin) CreateArticle(ctx context.Context, adminUserID int64, req *v1.CreateArticleRequest, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if err := a.validateCategory(ctx, req.CategoryID); err != nil {
		if err.Error() == "category not found" {
			return nil, http.StatusBadRequest, err
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := a.validateTags(ctx, req.TagIDs); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, http.StatusBadRequest, err
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	st := models.ArticleStatusDraft
	if req.Status != "" {
		st = models.ArticleStatus(req.Status)
	}

	slugStr := strings.TrimSpace(req.Slug)
	if slugStr != "" {
		var ok bool
		slugStr, ok = slug.Normalize(slugStr)
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := a.slugTaken(ctx, slugStr, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
	} else {
		base := slug.FromTitle(req.Title)
		var err error
		slugStr, err = a.allocUniqueSlug(ctx, base, 0)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	pubAt, err := parseRFC3339Ptr(req.PublishedAt)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid publishedAt")
	}
	if st == models.ArticleStatusPublished && pubAt == nil {
		now := time.Now()
		pubAt = &now
	}

	art := &models.Article{
		Title:       req.Title,
		Slug:        slugStr,
		Content:     req.Content,
		Summary:     req.Summary,
		CoverImage:  req.CoverImage,
		Status:      st,
		AuthorID:    adminUserID,
		CategoryID:  req.CategoryID,
		PublishedAt: pubAt,
	}

	tx := a.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(art).Error; err != nil {
		tx.Rollback()
		klog.ErrorS(err, "CreateArticle")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := a.replaceTags(ctx, tx, art.ID, req.TagIDs); err != nil {
		tx.Rollback()
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	maybeHexoSyncSingle(a.svc, art.ID)
	return a.loadArticleDetail(ctx, art.ID)
}

// UpdateArticle 更新文章。
func (a *ArticleAdmin) UpdateArticle(ctx context.Context, articleID int64, req *v1.UpdateArticleRequest, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	var art models.Article
	if err := a.svc.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", articleID).First(&art).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if req.CategoryID != nil {
		if err := a.validateCategory(ctx, req.CategoryID); err != nil {
			if err.Error() == "category not found" {
				return nil, http.StatusBadRequest, err
			}
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	tagUpdate := len(req.TagIDs) > 0
	if tagUpdate {
		if err := a.validateTags(ctx, req.TagIDs); err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusBadRequest, err
			}
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
		s, ok := slug.Normalize(*req.Slug)
		if !ok {
			return nil, http.StatusBadRequest, errors.New("invalid slug")
		}
		taken, err := a.slugTaken(ctx, s, articleID)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		if taken {
			return nil, http.StatusConflict, errors.New("slug already exists")
		}
		updates["slug"] = s
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	pubAt, err := parseRFC3339Ptr(req.PublishedAt)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid publishedAt")
	}
	if pubAt != nil {
		updates["published_at"] = pubAt
	}
	if req.CategoryID != nil {
		updates["category_id"] = req.CategoryID
	}

	if len(updates) == 0 && !tagUpdate {
		return a.loadArticleDetail(ctx, articleID)
	}

	tx := a.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if len(updates) > 0 {
		if err := tx.Model(&models.Article{}).Where("id = ?", articleID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if tagUpdate {
		if err := a.replaceTags(ctx, tx, articleID, req.TagIDs); err != nil {
			tx.Rollback()
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	maybeHexoSyncSingle(a.svc, articleID)
	return a.loadArticleDetail(ctx, articleID)
}

// DeleteArticle 软删除。
func (a *ArticleAdmin) DeleteArticle(ctx context.Context, articleID int64, _ *http.Request) (*v1.DeleteArticleResponse, int, error) {
	res := a.svc.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", articleID).Delete(&models.Article{})
	if res.Error != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if res.RowsAffected == 0 {
		return nil, http.StatusNotFound, errors.New("article not found")
	}
	_ = a.svc.DB.WithContext(ctx).Where("article_id = ?", articleID).Delete(&models.ArticleTag{})
	maybeHexoDeletePost(a.svc, articleID)
	return &v1.DeleteArticleResponse{ID: articleID}, http.StatusOK, nil
}

// UpdateArticleStatus 更新状态。
func (a *ArticleAdmin) UpdateArticleStatus(ctx context.Context, articleID int64, req *v1.UpdateArticleStatusRequest, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	pubAt, err := parseRFC3339Ptr(req.PublishedAt)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid publishedAt")
	}
	updates := map[string]interface{}{"status": req.Status}
	if models.ArticleStatus(req.Status) == models.ArticleStatusPublished && pubAt == nil {
		now := time.Now()
		pubAt = &now
	}
	if pubAt != nil {
		updates["published_at"] = pubAt
	}
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("id = ? AND deleted_at IS NULL", articleID).Updates(updates).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	maybeHexoSyncSingle(a.svc, articleID)
	return a.loadArticleDetail(ctx, articleID)
}

// UpdateArticleSlug 更新 slug。
func (a *ArticleAdmin) UpdateArticleSlug(ctx context.Context, articleID int64, req *v1.UpdateArticleSlugRequest, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	s, ok := slug.Normalize(req.Slug)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("invalid slug")
	}
	taken, err := a.slugTaken(ctx, s, articleID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if taken {
		return nil, http.StatusConflict, errors.New("slug already exists")
	}
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("id = ? AND deleted_at IS NULL", articleID).Update("slug", s).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	maybeHexoSyncSingle(a.svc, articleID)
	return a.loadArticleDetail(ctx, articleID)
}

// UpdateArticlePassword 设置或清除密码（bcrypt 存储）。
func (a *ArticleAdmin) UpdateArticlePassword(ctx context.Context, articleID int64, req *v1.UpdateArticlePasswordRequest, _ *http.Request) (*v1.ArticleSecurityResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	pw := strings.TrimSpace(req.Password)
	if pw != "" && (len(pw) < 4 || len(pw) > 20) {
		return nil, http.StatusBadRequest, errors.New("invalid password length")
	}
	var hash string
	if pw != "" {
		h, err := passwd.HashPassword(pw)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		hash = h
	}
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("id = ? AND deleted_at IS NULL", articleID).Update("password", hash).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	protected := hash != ""
	return &v1.ArticleSecurityResponse{Protected: protected}, http.StatusOK, nil
}

// UpdateArticlePin 置顶。
func (a *ArticleAdmin) UpdateArticlePin(ctx context.Context, articleID int64, req *v1.UpdateArticlePinRequest, _ *http.Request) (*v1.ArticleDetailResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	po := req.PinOrder
	if !req.IsPinned {
		po = 0
	}
	if err := a.svc.DB.WithContext(ctx).Model(&models.Article{}).Where("id = ? AND deleted_at IS NULL", articleID).
		Updates(map[string]interface{}{"is_pinned": req.IsPinned, "pin_order": po}).Error; err != nil {
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	maybeHexoSyncSingle(a.svc, articleID)
	return a.loadArticleDetail(ctx, articleID)
}

// --- Handlers ---

func registerArticleAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	s := newArticleAdmin(svcCtx)
	g.POST("/articles", s.handleCreateArticle)
	g.PUT("/articles/:id", s.handleUpdateArticle)
	g.DELETE("/articles/:id", s.handleDeleteArticle)
	g.PUT("/articles/:id/status", s.handleUpdateArticleStatus)
	g.PUT("/articles/:id/slug", s.handleUpdateArticleSlug)
	g.PUT("/articles/:id/password", s.handleUpdateArticlePassword)
	g.PUT("/articles/:id/pin", s.handleUpdateArticlePin)
}

func (s *ArticleAdmin) handleCreateArticle(c *gin.Context) {
	uid, ok := middlewares.GetCurrentUserID(c)
	if !ok || uid <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req v1.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.CreateArticle(ctx, uid, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleUpdateArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.UpdateArticle(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleDeleteArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.DeleteArticle(ctx, id, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleUpdateArticleStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateArticleStatus(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleUpdateArticleSlug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleSlugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateArticleSlug(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleUpdateArticlePassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticlePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateArticlePassword(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

func (s *ArticleAdmin) handleUpdateArticlePin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticlePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateArticlePin(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
