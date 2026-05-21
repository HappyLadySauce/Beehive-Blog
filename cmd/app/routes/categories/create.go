package categories

import (
	"context"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// create creates a new category.
// create 创建新分类。
func (c *CategoriesController) create(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.CreateCategoryResponse, error) {
	cat := model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}
	if err := c.svc.DB.WithContext(ctx).Create(&cat).Error; err != nil {
		return nil, mapCategoryCreateUniqueViolation(err)
	}
	return &v1.CreateCategoryResponse{ID: cat.ID}, nil
}

// Create handles POST /api/v1/categories (admin).
// Create 处理 POST /api/v1/categories（管理员）。
//
//	@Summary		Create category
//	@Description	Creates a new category. Admin only. 中文：创建新分类（仅管理员）。
//	@Tags			categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		v1.CreateCategoryRequest	true	"Category creation request"
//	@Success		200		{object}	common.BaseResponse{data=v1.CreateCategoryResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse	"Slug conflict"
//	@Router			/api/v1/categories [post]
func (c *CategoriesController) Create(ctx *gin.Context) {
	var req v1.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.create(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
