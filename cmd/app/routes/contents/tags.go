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

// getContentTags returns tags attached to a content item.
// getContentTags 返回内容所关联的标签。
func (c *ContentsController) getContentTags(ctx context.Context, contentID int64, admin bool) ([]v1.TagItem, error) {
	existQuery := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID)
	if !admin {
		existQuery = existQuery.Where("status = ? AND visibility = ?", "published", "public")
	}
	var count int64
	if err := existQuery.Count(&count).Error; err != nil {
		return nil, common.NewInternal("failed to check content", err)
	}
	if count == 0 {
		return nil, common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	return loadContentTags(ctx, c, contentID, !admin)
}

// GetContentTags handles GET /api/v1/contents/:id/tags.
// GetContentTags 处理 GET /api/v1/contents/:id/tags。
//
//	@Summary		List content tags
//	@Description	Returns tags attached to a content item. Without token only for public published content; admin Bearer for any content. 中文：返回内容关联的标签；无 token 仅公开已发布内容，管理员 Bearer 可访问任意内容。
//	@Tags			contents
//	@Produce		json
//	@Param			id	path		int	true	"Content ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/tags [get]
func (c *ContentsController) GetContentTags(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	tags, err := c.getContentTags(ctx.Request.Context(), id, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, tags)
}

// setContentTags replaces all tags on a content item atomically.
// setContentTags 原子性地替换内容的全部标签。
func (c *ContentsController) setContentTags(ctx context.Context, contentID int64, req *v1.SetContentTagsRequest) error {
	tagIDs := uniqueInt64(req.TagIDs)
	if len(tagIDs) != len(req.TagIDs) {
		return common.NewBadRequest("duplicate tag IDs are not allowed", fmt.Errorf("duplicate tag ids"))
	}

	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil {
		return common.NewInternal("failed to check content", err)
	}
	if count == 0 {
		return common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	if len(tagIDs) > 0 {
		var tagCount int64
		if err := c.svc.DB.WithContext(ctx).Model(&model.Tag{}).Where("id IN ?", tagIDs).Count(&tagCount).Error; err != nil {
			return common.NewInternal("failed to validate tag IDs", err)
		}
		if tagCount != int64(len(tagIDs)) {
			return common.NewBadRequest("one or more tag IDs do not exist", fmt.Errorf("tag validation failed"))
		}
	}

	tx := c.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return common.NewInternal("failed to begin transaction", tx.Error)
	}

	if err := tx.Where("content_id = ?", contentID).Delete(&model.ContentTag{}).Error; err != nil {
		tx.Rollback()
		return common.NewInternal("failed to clear content tags", err)
	}

	for _, tagID := range tagIDs {
		ct := model.ContentTag{ContentID: contentID, TagID: tagID}
		if err := tx.Create(&ct).Error; err != nil {
			tx.Rollback()
			return common.NewInternal("failed to set content tags", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return common.NewInternal("failed to commit tag changes", err)
	}
	return nil
}

// SetTags handles PUT /api/v1/contents/:id/tags (admin).
// SetTags 处理 PUT /api/v1/contents/:id/tags（管理员）。
//
//	@Summary		Set content tags
//	@Description	Replaces all tags on a content item atomically. Admin only. 中文：原子性替换内容的全部标签（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Content ID"
//	@Param			body	body		v1.SetContentTagsRequest	true	"Tag ID list"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/tags [put]
func (c *ContentsController) SetTags(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.SetContentTagsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	if err := c.setContentTags(ctx.Request.Context(), id, &req); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}
