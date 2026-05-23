package comments

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

type adminCommentRow struct {
	ID           int64
	ContentID    int64
	ContentTitle string
	ContentSlug  string
	Nickname     string
	Website      *string
	Body         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// listAdmin returns moderated comments for Studio management.
// listAdmin 返回供 Studio 管理的审核评论列表。
func (c *CommentsController) listAdmin(ctx context.Context, req *v1.ListAdminCommentsRequest) (*v1.ListAdminCommentsResponse, error) {
	page, pageSize := pagination.NormalizeOffset(req.Page, req.PageSize)
	query := c.db.WithContext(ctx).Table("content.comments AS comments").
		Select("comments.id, comments.content_id, contents.title AS content_title, contents.slug AS content_slug, comments.nickname, comments.website, comments.body, comments.status, comments.created_at, comments.updated_at").
		Joins("JOIN content.contents AS contents ON contents.id = comments.content_id").
		Where("comments.deleted_at IS NULL")
	if req.Status != "" {
		query = query.Where("comments.status = ?", req.Status)
	}
	if req.ContentID > 0 {
		query = query.Where("comments.content_id = ?", req.ContentID)
	}
	if keyword := strings.TrimSpace(req.Search); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("comments.nickname ILIKE ? OR comments.body ILIKE ? OR contents.title ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list comments", fmt.Errorf("count: %w", err))
	}
	var rows []adminCommentRow
	if err := query.Order("comments.created_at DESC, comments.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, common.NewInternal("failed to list comments", fmt.Errorf("find: %w", err))
	}
	items := make([]v1.AdminCommentItem, len(rows))
	for i, row := range rows {
		items[i] = v1.AdminCommentItem{
			ID:           row.ID,
			ContentID:    row.ContentID,
			ContentTitle: row.ContentTitle,
			ContentSlug:  row.ContentSlug,
			Nickname:     row.Nickname,
			Website:      row.Website,
			Body:         row.Body,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}
	return &v1.ListAdminCommentsResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// ListAdmin handles GET /api/v1/comments.
// ListAdmin 处理 GET /api/v1/comments。
func (c *CommentsController) ListAdmin(ctx *gin.Context) {
	var req v1.ListAdminCommentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeCommentError(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	resp, err := c.listAdmin(ctx.Request.Context(), &req)
	if err != nil {
		writeCommentError(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

func (c *CommentsController) updateStatus(ctx context.Context, id int64, status string) (*v1.AdminCommentItem, error) {
	var row model.Comment
	if err := c.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, mapCommentFirstError(err)
	}
	if err := c.db.WithContext(ctx).Model(&row).Updates(map[string]interface{}{"status": status, "updated_at": nowUTC()}).Error; err != nil {
		return nil, common.NewInternal("failed to update comment status", err)
	}
	resp, err := c.listAdmin(ctx, &v1.ListAdminCommentsRequest{ContentID: row.ContentID, Page: 1, PageSize: 100})
	if err != nil {
		return nil, err
	}
	for _, item := range resp.Items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, common.NewInternal("failed to load updated comment", gorm.ErrRecordNotFound)
}

// UpdateStatus handles PATCH /api/v1/comments/:commentId/status.
// UpdateStatus 处理 PATCH /api/v1/comments/:commentId/status。
func (c *CommentsController) UpdateStatus(ctx *gin.Context) {
	id, ok := parseCommentID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateCommentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeCommentError(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.updateStatus(ctx.Request.Context(), id, req.Status)
	if err != nil {
		writeCommentError(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

func (c *CommentsController) delete(ctx context.Context, id int64) error {
	result := c.db.WithContext(ctx).Delete(&model.Comment{}, id)
	if result.Error != nil {
		return common.NewInternal("failed to delete comment", result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFound("comment not found", gorm.ErrRecordNotFound)
	}
	return nil
}

// Delete handles DELETE /api/v1/comments/:commentId.
// Delete 处理 DELETE /api/v1/comments/:commentId。
func (c *CommentsController) Delete(ctx *gin.Context) {
	id, ok := parseCommentID(ctx)
	if !ok {
		return
	}
	if err := c.delete(ctx.Request.Context(), id); err != nil {
		writeCommentError(ctx, err)
		return
	}
	common.Success(ctx, map[string]any{})
}

func parseCommentID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("commentId")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		if err == nil {
			err = errors.New("id must be positive")
		}
		common.Fail(ctx, common.NewBadRequest("invalid comment id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}
