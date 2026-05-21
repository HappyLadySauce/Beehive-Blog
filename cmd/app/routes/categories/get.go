package categories

import (
	"context"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// get returns a single category by ID.
// get 根据 ID 返回单个分类。
func (c *CategoriesController) get(ctx context.Context, id int64) (*v1.CategoryDetailResponse, error) {
	var cat model.Category
	query := c.svc.DB.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id)
	if err := query.First(&cat).Error; err != nil {
		return nil, mapFirstError(err, "category not found", "failed to fetch category")
	}

	var contentCount int64
	if err := c.svc.DB.WithContext(ctx).Model(&model.ContentCategory{}).Where("category_id = ?", cat.ID).Count(&contentCount).Error; err != nil {
		return nil, common.NewInternal("failed to count category content", err)
	}

	item := toCategoryItem(cat)
	item.ContentCount = contentCount
	return &v1.CategoryDetailResponse{CategoryItem: item}, nil
}

// Get handles GET /api/v1/categories/:id.
// Get 处理 GET /api/v1/categories/:id。
//
//	@Summary		Get category detail
//	@Description	Returns a non-deleted category by ID. Soft-deleted categories return 404. 中文：按 ID 返回未软删分类；已软删返回 404。
//	@Tags			categories
//	@Produce		json
//	@Param			id	path		int	true	"Category ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.CategoryDetailResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/categories/{id} [get]
func (c *CategoriesController) Get(ctx *gin.Context) {
	id, ok := parseCategoryID(ctx)
	if !ok {
		return
	}
	resp, err := c.get(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// update performs a PATCH update on a category.
// update 对分类执行 PATCH 更新。
func (c *CategoriesController) update(ctx context.Context, id int64, req *v1.UpdateCategoryRequest) (*v1.CategoryDetailResponse, error) {
	var cat model.Category
	if err := c.svc.DB.WithContext(ctx).First(&cat, id).Error; err != nil {
		return nil, mapFirstError(err, "category not found", "failed to fetch category")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Description != nil {
		if *req.Description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = *req.Description
		}
	}
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			updates["parent_id"] = nil
		} else {
			updates["parent_id"] = *req.ParentID
		}
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		return c.get(ctx, id)
	}

	if err := c.svc.DB.WithContext(ctx).Model(&cat).Updates(updates).Error; err != nil {
		return nil, mapCategoryUpdateUniqueViolation(err)
	}
	return c.get(ctx, id)
}

// Update handles PATCH /api/v1/categories/:id (admin).
// Update 处理 PATCH /api/v1/categories/:id（管理员）。
//
//	@Summary		Update category
//	@Description	Updates a category. Admin only. 中文：更新分类（仅管理员）。
//	@Tags			categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Category ID"
//	@Param			body	body		v1.UpdateCategoryRequest	true	"Category update request"
//	@Success		200		{object}	common.BaseResponse{data=v1.CategoryDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/categories/{id} [patch]
func (c *CategoriesController) Update(ctx *gin.Context) {
	id, ok := parseCategoryID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateCategoryRequest
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
