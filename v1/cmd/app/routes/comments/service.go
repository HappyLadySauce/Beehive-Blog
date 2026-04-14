package comments

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Service 评论业务。
type Service struct {
	svc *svc.ServiceContext
}

// NewService constructs Service.
func NewService(svcCtx *svc.ServiceContext) *Service {
	return &Service{svc: svcCtx}
}

func clientIP(r *http.Request) string {
	if r == nil {
		return "unknown"
	}
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		ip := strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
		if ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return "unknown"
	}
	parsedHost, _, err := net.SplitHostPort(host)
	if err == nil && parsedHost != "" {
		return parsedHost
	}
	return host
}

func mapCommentAuthor(u *models.User) v1.CommentAuthorItem {
	if u == nil {
		return v1.CommentAuthorItem{}
	}
	return v1.CommentAuthorItem{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

// ListByArticle 公开：仅已通过审核的评论。
func (s *Service) ListByArticle(ctx context.Context, articleID int64, page, pageSize int) (*v1.CommentListResponse, int, error) {
	if articleID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid article id")
	}
	var a models.Article
	err := s.svc.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL AND status = ?", articleID, models.ArticleStatusPublished).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		klog.ErrorS(err, "ListByArticle load article", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	_ = a

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	base := s.svc.DB.WithContext(ctx).Model(&models.Comment{}).
		Where("article_id = ? AND status = ?", articleID, models.CommentStatusApproved)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		klog.ErrorS(err, "ListByArticle count")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var rows []models.Comment
	if err := s.svc.DB.WithContext(ctx).
		Preload("User").
		Where("article_id = ? AND status = ?", articleID, models.CommentStatusApproved).
		Order("created_at ASC").
		Offset(offset).Limit(pageSize).
		Find(&rows).Error; err != nil {
		klog.ErrorS(err, "ListByArticle find")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.CommentItem, 0, len(rows))
	for i := range rows {
		c := &rows[i]
		items = append(items, v1.CommentItem{
			ID:        c.ID,
			Content:   c.Content,
			ParentID:  c.ParentID,
			CreatedAt: c.CreatedAt,
			Author:    mapCommentAuthor(c.User),
		})
	}
	return &v1.CommentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// Create 登录用户发表评论（待审核）。
func (s *Service) Create(ctx context.Context, articleID, userID int64, req *v1.CreateCommentRequest, r *http.Request) (*v1.CreateCommentResponse, int, error) {
	if req == nil || r == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if articleID <= 0 || userID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid article id")
	}
	var a models.Article
	if err := s.svc.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL AND status = ?", articleID, models.ArticleStatusPublished).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("article not found")
		}
		klog.ErrorS(err, "Create load article", "articleID", articleID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var parentID *int64
	if req.ParentID != nil && *req.ParentID > 0 {
		pid := *req.ParentID
		var parent models.Comment
		if err := s.svc.DB.WithContext(ctx).
			Where("id = ? AND article_id = ? AND status = ?", pid, articleID, models.CommentStatusApproved).
			First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, http.StatusBadRequest, errors.New("invalid parent comment")
			}
			klog.ErrorS(err, "Create load parent", "parentID", pid)
			return nil, http.StatusInternalServerError, errors.New("system error")
		}
		parentID = &pid
	}

	ip := clientIP(r)
	c := models.Comment{
		Content:   strings.TrimSpace(req.Content),
		Status:    models.CommentStatusPending,
		ArticleID: articleID,
		UserID:    &userID,
		ParentID:  parentID,
		AuthorIP:  ip,
	}
	if err := s.svc.DB.WithContext(ctx).Create(&c).Error; err != nil {
		klog.ErrorS(err, "Create comment")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	if a.AuthorID != userID {
		n := models.Notification{
			UserID:     a.AuthorID,
			Type:       models.NotificationTypeComment,
			Title:      "New comment on your article",
			Content:    fmt.Sprintf("Article id %d received a new comment pending review.", articleID),
			IsRead:     false,
			SourceID:   fmt.Sprintf("%d", c.ID),
			SourceType: "comment",
		}
		if err := s.svc.DB.WithContext(ctx).Create(&n).Error; err != nil {
			klog.ErrorS(err, "Create notification for comment", "commentID", c.ID)
		}
	}

	return &v1.CreateCommentResponse{ID: c.ID}, http.StatusOK, nil
}

func (s *Service) adminCommentBaseQuery(ctx context.Context, q *v1.AdminCommentListQuery) *gorm.DB {
	listQ := s.svc.DB.WithContext(ctx).Model(&models.Comment{})
	if q.ArticleID > 0 {
		listQ = listQ.Where("article_id = ?", q.ArticleID)
	}
	if strings.TrimSpace(q.Status) != "" {
		listQ = listQ.Where("status = ?", models.CommentStatus(q.Status))
	}
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		pat := "%" + kw + "%"
		listQ = listQ.Where("content ILIKE ?", pat)
	}
	return listQ
}

// AdminList 管理员评论列表。
func (s *Service) AdminList(ctx context.Context, q *v1.AdminCommentListQuery) (*v1.AdminCommentListResponse, int, error) {
	if q == nil {
		q = &v1.AdminCommentListQuery{}
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := s.adminCommentBaseQuery(ctx, q).Count(&total).Error; err != nil {
		klog.ErrorS(err, "AdminList count")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	var rows []models.Comment
	if err := s.adminCommentBaseQuery(ctx, q).Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&rows).Error; err != nil {
		klog.ErrorS(err, "AdminList find")
		return nil, http.StatusInternalServerError, errors.New("system error")
	}

	items := make([]v1.AdminCommentItem, 0, len(rows))
	for i := range rows {
		c := &rows[i]
		items = append(items, v1.AdminCommentItem{
			ID:        c.ID,
			Content:   c.Content,
			Status:    string(c.Status),
			ArticleID: c.ArticleID,
			UserID:    c.UserID,
			ParentID:  c.ParentID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Author:    mapCommentAuthor(c.User),
		})
	}
	return &v1.AdminCommentListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, http.StatusOK, nil
}

// UpdateAdminStatus 管理员更新评论状态。
func (s *Service) UpdateAdminStatus(ctx context.Context, commentID int64, req *v1.UpdateCommentStatusRequest) (*v1.UpdateCommentStatusResponse, int, error) {
	if req == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	if commentID <= 0 {
		return nil, http.StatusBadRequest, errors.New("invalid comment id")
	}
	st := models.CommentStatus(req.Status)
	var c models.Comment
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", commentID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("comment not found")
		}
		klog.ErrorS(err, "UpdateAdminStatus load", "id", commentID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	if err := s.svc.DB.WithContext(ctx).Model(&c).Updates(map[string]interface{}{
		"status":     st,
		"updated_at": time.Now().UTC(),
	}).Error; err != nil {
		klog.ErrorS(err, "UpdateAdminStatus save", "id", commentID)
		return nil, http.StatusInternalServerError, errors.New("system error")
	}
	return &v1.UpdateCommentStatusResponse{ID: c.ID, Status: string(st)}, http.StatusOK, nil
}
