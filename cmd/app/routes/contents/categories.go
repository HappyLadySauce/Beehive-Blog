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

// getContentCategories returns categories attached to a content item.
// getContentCategories 返回内容所关联的分类。
func (c *ContentsController) getContentCategories(ctx context.Context, contentID int64, admin bool) ([]v1.CategoryItem, error) {
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

	return loadContentCategories(ctx, c, contentID)
}

// GetContentCategories handles GET /api/v1/contents/:id/categories.
// GetContentCategories 处理 GET /api/v1/contents/:id/categories。
//
//	@Summary		List content categories
//	@Description	Returns categories attached to a content item. Without token only for public published content; admin Bearer for any content. 中文：返回内容关联的分类；无 token 仅公开已发布内容，管理员 Bearer 可访问任意内容。
//	@Tags			contents
//	@Produce		json
//	@Param			id	path		int	true	"Content ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/categories [get]
func (c *ContentsController) GetContentCategories(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	claims := middleware.GetClaims(ctx)
	admin := claims != nil && claims.Role == "admin"
	categories, err := c.getContentCategories(ctx.Request.Context(), id, admin)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, categories)
}

// setContentCategories replaces all categories on a content item atomically.
// setContentCategories 原子性地替换内容的全部分类。
func (c *ContentsController) setContentCategories(ctx context.Context, contentID int64, req *v1.SetContentCategoriesRequest) error {
	catIDs := uniqueInt64(req.CategoryIDs)
	if len(catIDs) != len(req.CategoryIDs) {
		return common.NewBadRequest("duplicate category IDs are not allowed", fmt.Errorf("duplicate category ids"))
	}

	var count int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.Content{}).Where("id = ?", contentID).Count(&count).Error; err != nil {
		return common.NewInternal("failed to check content", err)
	}
	if count == 0 {
		return common.NewNotFound("content not found", fmt.Errorf("content %d not found", contentID))
	}

	if len(catIDs) > 0 {
		var catCount int64
		if err := c.svc.DB.WithContext(ctx).Model(&model.Category{}).Where("id IN ?", catIDs).Count(&catCount).Error; err != nil {
			return common.NewInternal("failed to validate category IDs", err)
		}
		if catCount != int64(len(catIDs)) {
			return common.NewBadRequest("one or more category IDs do not exist", fmt.Errorf("category validation failed"))
		}
	}

	tx := c.svc.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return common.NewInternal("failed to begin transaction", tx.Error)
	}

	if err := tx.Where("content_id = ?", contentID).Delete(&model.ContentCategory{}).Error; err != nil {
		tx.Rollback()
		return common.NewInternal("failed to clear content categories", err)
	}

	for _, catID := range catIDs {
		cc := model.ContentCategory{ContentID: contentID, CategoryID: catID}
		if err := tx.Create(&cc).Error; err != nil {
			tx.Rollback()
			return common.NewInternal("failed to set content categories", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return common.NewInternal("failed to commit category changes", err)
	}
	return nil
}

// SetCategories handles PUT /api/v1/contents/:id/categories (admin).
// SetCategories 处理 PUT /api/v1/contents/:id/categories（管理员）。
//
//	@Summary		Set content categories
//	@Description	Replaces all categories on a content item atomically. Admin only. 中文：原子性替换内容的全部分类（仅管理员）。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"Content ID"
//	@Param			body	body		v1.SetContentCategoriesRequest	true	"Category ID list"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/categories [put]
func (c *ContentsController) SetCategories(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.SetContentCategoriesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	if err := c.setContentCategories(ctx.Request.Context(), id, &req); err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, nil)
}
