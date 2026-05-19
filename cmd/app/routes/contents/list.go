package contents

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// list queries contents with pagination and optional filters.
// list 查询内容列表（分页+可选筛选）。
func (c *ContentsController) list(ctx context.Context, req *v1.ListContentsRequest, admin bool) (*v1.ListContentsResponse, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := c.svc.DB.WithContext(ctx).Model(&model.Content{})

	if !admin {
		query = query.Where("status = ? AND visibility = ?", "published", "public")
	} else {
		if req.Status != "" {
			query = query.Where("status = ?", req.Status)
		}
		if req.Visibility != "" {
			query = query.Where("visibility = ?", req.Visibility)
		}
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.TagID > 0 {
		subQuery := c.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).
			Select("content_id").Where("tag_id = ?", req.TagID)
		query = query.Where("id IN (?)", subQuery)
	}
	if req.Search != "" {
		pattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(COALESCE(excerpt, '')) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.NewInternal("failed to list contents", fmt.Errorf("count: %w", err))
	}

	var contents []model.Content
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&contents).Error; err != nil {
		return nil, common.NewInternal("failed to list contents", fmt.Errorf("find: %w", err))
	}

	items := make([]v1.ContentItem, len(contents))
	for i, content := range contents {
		item := toContentItem(content)
		var user model.User
		if err := c.svc.DB.WithContext(ctx).Select("username").First(&user, content.AuthorID).Error; err == nil {
			item.AuthorUsername = user.Username
		}
		item.Tags = loadContentTags(ctx, c, content.ID)
		items[i] = item
	}

	return &v1.ListContentsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List handles GET /api/v1/contents.
// List 处理 GET /api/v1/contents。
func (c *ContentsController) List(ctx *gin.Context) {
	var req v1.ListContentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	resp, err := c.list(ctx.Request.Context(), &req, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// loadContentTags fetches tags for a given content ID.
// loadContentTags 获取指定内容的标签列表。
func loadContentTags(ctx context.Context, ctrl *ContentsController, contentID int64) []v1.TagItem {
	var cts []model.ContentTag
	if err := ctrl.svc.DB.WithContext(ctx).Where("content_id = ?", contentID).Find(&cts).Error; err != nil {
		return nil
	}
	if len(cts) == 0 {
		return nil
	}
	tagIDs := make([]int64, len(cts))
	for i, ct := range cts {
		tagIDs[i] = ct.TagID
	}
	var tags []model.Tag
	if err := ctrl.svc.DB.WithContext(ctx).Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
		return nil
	}
	items := make([]v1.TagItem, len(tags))
	for i, t := range tags {
		items[i] = v1.TagItem{
			ID:    t.ID,
			Name:  t.Name,
			Slug:  t.Slug,
			Color: t.Color,
		}
	}
	return items
}
