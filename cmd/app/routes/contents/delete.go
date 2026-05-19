package contents

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// del soft-deletes a content row.
// del 软删除内容行。
func (c *ContentsController) del(ctx context.Context, id int64) error {
	var content model.Content
	if err := c.svc.DB.WithContext(ctx).First(&content, id).Error; err != nil {
		return common.NewNotFound("content not found", err)
	}
	if err := c.svc.DB.WithContext(ctx).Delete(&content).Error; err != nil {
		return common.NewInternal("failed to delete content", err)
	}
	return nil
}

// Delete handles DELETE /api/v1/contents/:id (admin).
// Delete 处理 DELETE /api/v1/contents/:id（管理员）。
func (c *ContentsController) Delete(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	if err := c.del(ctx.Request.Context(), id); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}
