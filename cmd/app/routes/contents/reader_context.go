package contents

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/pagination"
)

// readerContext returns adjacent, related, recent, and comment data for a public content item.
// readerContext 返回公开内容详情页的相邻、相关、最近与评论数据。
func (c *ContentsController) readerContext(ctx context.Context, id int64) (*v1.PublicReaderContextResponse, error) {
	var content model.Content
	if err := c.svc.DB.WithContext(ctx).Where("id = ? AND status = ? AND visibility = ?", id, "published", "public").First(&content).Error; err != nil {
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	var previous, next *v1.PublicContentItem
	publishedAt := content.CreatedAt
	if content.PublishedAt != nil {
		publishedAt = *content.PublishedAt
	}

	if items, err := c.publicContentItems(ctx, publicContentQuery{ContentType: content.Type, ExcludeID: content.ID, Before: &publishedAt, Limit: 1}); err != nil {
		return nil, err
	} else if len(items) > 0 {
		previous = &items[0]
	}
	if items, err := c.publicContentItems(ctx, publicContentQuery{ContentType: content.Type, ExcludeID: content.ID, After: &publishedAt, Limit: 1}); err != nil {
		return nil, err
	} else if len(items) > 0 {
		next = &items[0]
	}

	related, err := c.publicContentItems(ctx, publicContentQuery{ContentType: content.Type, ExcludeID: content.ID, Limit: 4})
	if err != nil {
		return nil, err
	}
	recent, err := c.publicContentItems(ctx, publicContentQuery{ExcludeID: content.ID, Limit: 5})
	if err != nil {
		return nil, err
	}
	comments, err := c.publicComments(ctx, content.ID, 1, 10)
	if err != nil {
		return nil, err
	}
	return &v1.PublicReaderContextResponse{Previous: previous, Next: next, Related: related, Recent: recent, Comments: comments}, nil
}

// ReaderContext handles GET /api/v1/contents/:id/reader-context.
// ReaderContext 处理 GET /api/v1/contents/:id/reader-context。
func (c *ContentsController) ReaderContext(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	resp, err := c.readerContext(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

func (c *ContentsController) publicComments(ctx context.Context, contentID int64, pageInput int, pageSizeInput int) (v1.ListCommentsResponse, error) {
	page, pageSize := pagination.NormalizeOffset(pageInput, pageSizeInput)
	query := c.svc.DB.WithContext(ctx).Model(&model.Comment{}).Where("content_id = ? AND status = ?", contentID, "published")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return v1.ListCommentsResponse{}, common.NewInternal("failed to load comments", fmt.Errorf("count: %w", err))
	}
	var rows []model.Comment
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return v1.ListCommentsResponse{}, common.NewInternal("failed to load comments", fmt.Errorf("find: %w", err))
	}
	items := make([]v1.CommentItem, len(rows))
	for i, row := range rows {
		items[i] = v1.CommentItem{
			ID:        row.ID,
			ContentID: row.ContentID,
			Nickname:  row.Nickname,
			Website:   row.Website,
			Body:      row.Body,
			CreatedAt: row.CreatedAt,
		}
	}
	return v1.ListCommentsResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
