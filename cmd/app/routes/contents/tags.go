package contents

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// getContentTags returns tags attached to a content item.
// getContentTags 返回内容所关联的标签。
func (c *ContentsController) getContentTags(ctx context.Context, contentID int64) ([]v1.TagItem, error) {
	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil || count == 0 {
		return nil, common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	tags := loadContentTags(ctx, c, contentID)
	if tags == nil {
		tags = []v1.TagItem{}
	}
	return tags, nil
}

// GetContentTags handles GET /api/v1/contents/:id/tags.
// GetContentTags 处理 GET /api/v1/contents/:id/tags。
func (c *ContentsController) GetContentTags(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	tags, err := c.getContentTags(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, tags)
}

// setContentTags replaces all tags on a content item atomically.
// setContentTags 原子性地替换内容的全部标签。
func (c *ContentsController) setContentTags(ctx context.Context, contentID int64, req *v1.SetContentTagsRequest) error {
	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil || count == 0 {
		return common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	if len(req.TagIDs) > 0 {
		var tagCount int64
		c.svc.DB.WithContext(ctx).Model(&model.Tag{}).Where("id IN ?", req.TagIDs).Count(&tagCount)
		if tagCount != int64(len(req.TagIDs)) {
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

	for _, tagID := range req.TagIDs {
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
