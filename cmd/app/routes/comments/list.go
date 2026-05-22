package comments

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

// list returns published comments for a public content item.
// list 返回公开内容下已发布的评论。
func (c *CommentsController) list(ctx context.Context, contentID int64, req *v1.ListCommentsRequest) (*v1.ListCommentsResponse, error) {
	if err := c.db.WithContext(ctx).Where("id = ? AND status = ? AND visibility = ?", contentID, "published", "public").
		First(&model.Content{}).Error; err != nil {
		return nil, mapCommentFirstError(err)
	}

	page, pageSize := pagination.NormalizeOffset(req.Page, req.PageSize)
	query := c.db.WithContext(ctx).Model(&model.Comment{}).
		Where("content_id = ? AND status = ?", contentID, "published")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list comments", fmt.Errorf("count: %w", err))
	}

	var rows []model.Comment
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, common.NewInternal("failed to list comments", fmt.Errorf("find: %w", err))
	}

	items := make([]v1.CommentItem, len(rows))
	for i, row := range rows {
		items[i] = toCommentItem(row)
	}
	return &v1.ListCommentsResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// List handles GET /api/v1/contents/:id/comments.
// List 处理 GET /api/v1/contents/:id/comments。
func (c *CommentsController) List(ctx *gin.Context) {
	contentID, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.ListCommentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		writeCommentError(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	resp, err := c.list(ctx.Request.Context(), contentID, &req)
	if err != nil {
		writeCommentError(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
