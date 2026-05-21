package contents

import (
	"context"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// get returns a single content by ID.
// get 根据 ID 返回单个内容。
func (c *ContentsController) get(ctx context.Context, id int64, admin bool) (interface{}, error) {
	query := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", id)
	if !admin {
		query = query.Where("status = ? AND visibility = ?", "published", "public")
	}

	var content model.Content
	if err := query.First(&content).Error; err != nil {
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	var user model.User
	if err := c.svc.DB.WithContext(ctx).Select("username").First(&user, content.AuthorID).Error; err == nil {
		// authorUsername used below in both branches.
	}

	if !admin {
		if result := c.svc.DB.WithContext(ctx).
			Exec("UPDATE content.contents SET view_count = view_count + 1 WHERE id = ?", id); result.Error != nil {
			klog.ErrorS(result.Error, "failed to increment view count", "content_id", id)
		}

		item := toPublicContentItem(content)
		item.AuthorUsername = user.Username
		tags, err := loadContentTags(ctx, c, content.ID)
		if err != nil {
			return nil, common.NewInternal("failed to load content tags", err)
		}
		item.Tags = tags

		return &v1.PublicContentDetailResponse{
			PublicContentItem: item,
			Body:              content.Body,
		}, nil
	}

	item := toContentItem(content)
	item.AuthorUsername = user.Username
	tags, err := loadContentTags(ctx, c, content.ID)
	if err != nil {
		return nil, common.NewInternal("failed to load content tags", err)
	}
	item.Tags = tags

	return &v1.ContentDetailResponse{
		ContentItem: item,
		Body:        content.Body,
	}, nil
}

// Get handles GET /api/v1/contents/:id.
// Get 处理 GET /api/v1/contents/:id。
//
//	@Summary		Get content detail
//	@Description	Returns a single content by ID. Without token: data is v1.PublicContentDetailResponse. With admin Bearer: data is v1.ContentDetailResponse. 中文：无 token 时 data 为 PublicContentDetailResponse；管理员 Bearer 时 data 为 ContentDetailResponse（含草稿等）。
//	@Tags			contents
//	@Produce		json
//	@Param			id	path		int	true	"Content ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.ContentDetailResponse}	"Admin view; anonymous uses PublicContentDetailResponse shape"
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id} [get]
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
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	updates := map[string]interface{}{}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
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
		resp, err := c.get(ctx, id, true)
		if err != nil {
			return nil, err
		}
		return resp.(*v1.ContentDetailResponse), nil
	}

	if err := c.svc.DB.WithContext(ctx).Model(&content).Updates(updates).Error; err != nil {
		return nil, mapContentUpdateUniqueViolation(err)
	}
	resp, err := c.get(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return resp.(*v1.ContentDetailResponse), nil
}

// Update handles PATCH /api/v1/contents/:id (admin).
// Update 处理 PATCH /api/v1/contents/:id（管理员）。
//
//	@Summary		Update content
//	@Description	Updates a content item's fields. Admin only. 中文：更新内容字段（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Content ID"
//	@Param			body	body		v1.UpdateContentRequest	true	"Content update request (all fields optional)"
//	@Success		200		{object}	common.BaseResponse{data=v1.ContentDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse	"Slug conflict"
//	@Router			/api/v1/contents/{id} [patch]
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
