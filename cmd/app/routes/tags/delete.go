package tags

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// del soft-deletes a tag after checking for existing content references.
// del 在检查现有内容引用后软删除标签。
func (t *TagsController) del(ctx context.Context, id int64) error {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return mapFirstError(err, "tag not found", "failed to fetch tag")
	}

	var refCount int64
	if err := t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", id).Count(&refCount).Error; err != nil {
		return common.NewInternal("failed to check tag references", err)
	}
	if refCount > 0 {
		return common.NewConflict(
			fmt.Sprintf("tag is referenced by %d content item(s); remove references first", refCount),
			fmt.Errorf("tag %d has %d content references", id, refCount),
		)
	}

	if err := t.svc.DB.WithContext(ctx).Delete(&tag).Error; err != nil {
		return common.NewInternal("failed to delete tag", err)
	}
	return nil
}

// Delete handles DELETE /api/v1/tags/:id (admin).
// Delete 处理 DELETE /api/v1/tags/:id（管理员）。
//
//	@Summary		Delete tag
//	@Description	Soft-deletes a tag when not referenced by content; removed tags no longer appear in listings. Admin only. 中文：无内容引用时软删除标签，删除后列表不可见（仅管理员）。
//	@Tags			tags
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Tag ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		409	{object}	common.BaseResponse	"Tag still referenced"
//	@Router			/api/v1/tags/{id} [delete]
func (t *TagsController) Delete(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	if err := t.del(ctx.Request.Context(), id); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}
