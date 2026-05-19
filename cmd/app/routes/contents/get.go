package contents

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// get returns a single content by ID.
// get 根据 ID 返回单个内容。
func (c *ContentsController) get(ctx context.Context, id int64, admin bool) (*v1.ContentDetailResponse, error) {
	query := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", id)
	if !admin {
		query = query.Where("status = ? AND visibility = ?", "published", "public")
	}

	var content model.Content
	if err := query.First(&content).Error; err != nil {
		return nil, common.NewNotFound("content not found", err)
	}

	item := toContentItem(content)
	var user model.User
	if err := c.svc.DB.WithContext(ctx).Select("username").First(&user, content.AuthorID).Error; err == nil {
		item.AuthorUsername = user.Username
	}
	item.Tags = loadContentTags(ctx, c, content.ID)

	if !admin {
		c.svc.DB.WithContext(ctx).Model(&content).UpdateColumn("view_count", content.ViewCount+1)
	}

	return &v1.ContentDetailResponse{
		ContentItem: item,
		Body:        content.Body,
	}, nil
}

// Get handles GET /api/v1/contents/:id.
// Get 处理 GET /api/v1/contents/:id。
func (c *ContentsController) Get(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	resp, err := c.get(ctx.Request.Context(), id, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// update performs a PATCH update on a content row.
// update 对内容行执行 PATCH 更新。
func (c *ContentsController) update(ctx context.Context, id int64, req *v1.UpdateContentRequest) (*v1.ContentDetailResponse, error) {
	var content model.Content
	if err := c.svc.DB.WithContext(ctx).First(&content, id).Error; err != nil {
		return nil, common.NewNotFound("content not found", err)
	}

	updates := map[string]interface{}{}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
		if *req.Slug != content.Slug {
			checkType := content.Type
			if req.Type != nil {
				checkType = *req.Type
			}
			var existing int64
			c.svc.DB.WithContext(ctx).Model(&model.Content{}).
				Where("type = ? AND slug = ? AND deleted_at IS NULL AND id <> ?", checkType, *req.Slug, id).
				Count(&existing)
			if existing > 0 {
				return nil, common.NewConflict(
					fmt.Sprintf("slug %q is already taken for type %q", *req.Slug, checkType),
					fmt.Errorf("slug conflict on update: %s/%s", checkType, *req.Slug),
				)
			}
		}
		updates["slug"] = *req.Slug
	}
	if req.Excerpt != nil {
		if *req.Excerpt == "" {
			updates["excerpt"] = nil
		} else {
			updates["excerpt"] = *req.Excerpt
		}
	}
	if req.Body != nil {
		if *req.Body == "" {
			updates["body"] = nil
		} else {
			updates["body"] = *req.Body
		}
		wc := computeWordCount(req.Body)
		updates["word_count"] = wc
		updates["reading_time_minutes"] = computeReadingTime(wc)
	}
	if req.CoverAttachmentID != nil {
		updates["cover_attachment_id"] = *req.CoverAttachmentID
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.AIAccess != nil {
		updates["ai_access"] = *req.AIAccess
	}
	if req.WordCount != nil {
		updates["word_count"] = *req.WordCount
	}
	if req.ReadingTimeMinutes != nil {
		updates["reading_time_minutes"] = *req.ReadingTimeMinutes
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) == 0 {
		return c.get(ctx, id, true)
	}

	if err := c.svc.DB.WithContext(ctx).Model(&content).Updates(updates).Error; err != nil {
		return nil, mapContentCrudUniqueViolation(err)
	}
	return c.get(ctx, id, true)
}

// Update handles PATCH /api/v1/contents/:id (admin).
// Update 处理 PATCH /api/v1/contents/:id（管理员）。
func (c *ContentsController) Update(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.update(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
