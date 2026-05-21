package categories

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// del soft-deletes a category after checking for existing content references.
// del 在检查现有内容引用后软删除分类。
func (c *CategoriesController) del(ctx context.Context, id int64) error {
	var cat model.Category
	if err := c.svc.DB.WithContext(ctx).First(&cat, id).Error; err != nil {
		return mapFirstError(err, "category not found", "failed to fetch category")
	}

	var refCount int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.ContentCategory{}).Where("category_id = ?", id).Count(&refCount).Error; err != nil {
		return common.NewInternal("failed to check category references", err)
	}
	if refCount > 0 {
		return common.NewConflict(
			fmt.Sprintf("category is referenced by %d content item(s); remove references first", refCount),
			fmt.Errorf("category %d has %d content references", id, refCount),
		)
	}

	if err := c.svc.DB.WithContext(ctx).Delete(&cat).Error; err != nil {
		return common.NewInternal("failed to delete category", err)
	}
	return nil
}

// Delete handles DELETE /api/v1/categories/:id (admin).
// Delete 处理 DELETE /api/v1/categories/:id（管理员）。
//
//	@Summary		Delete category
//	@Description	Soft-deletes a category when not referenced by content; removed categories no longer appear in listings. Admin only. 中文：无内容引用时软删除分类，删除后列表不可见（仅管理员）。
//	@Tags			categories
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Category ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		409	{object}	common.BaseResponse	"Category still referenced"
//	@Router			/api/v1/categories/{id} [delete]
func (c *CategoriesController) Delete(ctx *gin.Context) {
	id, ok := parseCategoryID(ctx)
	if !ok {
		return
	}
	if err := c.del(ctx.Request.Context(), id); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}
