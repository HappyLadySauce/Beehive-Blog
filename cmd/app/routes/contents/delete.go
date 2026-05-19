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
		return mapFirstError(err, "content not found", "failed to fetch content")
	}
	if err := c.svc.DB.WithContext(ctx).Delete(&content).Error; err != nil {
		return common.NewInternal("failed to delete content", err)
	}
	return nil
}

// Delete handles DELETE /api/v1/contents/:id (admin).
// Delete 处理 DELETE /api/v1/contents/:id（管理员）。
//
//	@Summary		Delete content
//	@Description	Soft-deletes a content item. Admin only. 中文：软删除内容（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Content ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id} [delete]
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
