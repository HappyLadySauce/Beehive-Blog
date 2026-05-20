package tags

import (
	"context"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// create creates a new tag.
// create 创建新标签。
func (t *TagsController) create(ctx context.Context, req *v1.CreateTagRequest) (*v1.CreateTagResponse, error) {
	tag := model.Tag{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Status:      "active",
	}
	if err := t.svc.DB.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, mapTagCreateUniqueViolation(err)
	}
	return &v1.CreateTagResponse{ID: tag.ID}, nil
}

// Create handles POST /api/v1/tags (admin).
// Create 处理 POST /api/v1/tags（管理员）。
//
//	@Summary		Create tag
//	@Description	Creates a new tag. Admin only. 中文：创建新标签（仅管理员）。
//	@Tags			tags
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		v1.CreateTagRequest	true	"Tag creation request"
//	@Success		200		{object}	common.BaseResponse{data=v1.CreateTagResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse	"Name or slug conflict"
//	@Router			/api/v1/tags [post]
func (t *TagsController) Create(ctx *gin.Context) {
	var req v1.CreateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := t.create(ctx.Request.Context(), &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
