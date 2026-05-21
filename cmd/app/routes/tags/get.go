package tags

import (
	"context"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// get returns a single tag by ID.
// get 根据 ID 返回单个标签。
func (t *TagsController) get(ctx context.Context, id int64) (*v1.TagDetailResponse, error) {
	var tag model.Tag
	query := t.svc.DB.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id)
	if err := query.First(&tag).Error; err != nil {
		return nil, mapFirstError(err, "tag not found", "failed to fetch tag")
	}

	var contentCount int64
	if err := t.svc.DB.WithContext(ctx).Model(&model.ContentTag{}).Where("tag_id = ?", tag.ID).Count(&contentCount).Error; err != nil {
		return nil, common.NewInternal("failed to count tag content", err)
	}

	item := toTagItem(tag)
	item.ContentCount = contentCount
	return &v1.TagDetailResponse{TagItem: item}, nil
}

// Get handles GET /api/v1/tags/:id.
// Get 处理 GET /api/v1/tags/:id。
//
//	@Summary		Get tag detail
//	@Description	Returns a non-deleted tag by ID. Soft-deleted tags return 404. 中文：按 ID 返回未软删标签；已软删返回 404。
//	@Tags			tags
//	@Produce		json
//	@Param			id	path		int	true	"Tag ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.TagDetailResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/tags/{id} [get]
func (t *TagsController) Get(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	resp, err := t.get(ctx.Request.Context(), id)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}

// update performs a PATCH update on a tag.
// update 对标签执行 PATCH 更新。
func (t *TagsController) update(ctx context.Context, id int64, req *v1.UpdateTagRequest) (*v1.TagDetailResponse, error) {
	var tag model.Tag
	if err := t.svc.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, mapFirstError(err, "tag not found", "failed to fetch tag")
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
	if req.Color != nil {
		if *req.Color == "" {
			updates["color"] = nil
		} else {
			updates["color"] = *req.Color
		}
	}
	if len(updates) == 0 {
		return t.get(ctx, id)
	}

	if err := t.svc.DB.WithContext(ctx).Model(&tag).Updates(updates).Error; err != nil {
		return nil, mapTagUpdateUniqueViolation(err)
	}
	return t.get(ctx, id)
}

// Update handles PATCH /api/v1/tags/:id (admin).
// Update 处理 PATCH /api/v1/tags/:id（管理员）。
//
//	@Summary		Update tag
//	@Description	Updates a tag. Admin only. 中文：更新标签（仅管理员）。
//	@Tags			tags
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Tag ID"
//	@Param			body	body		v1.UpdateTagRequest	true	"Tag update request"
//	@Success		200		{object}	common.BaseResponse{data=v1.TagDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/tags/{id} [patch]
func (t *TagsController) Update(ctx *gin.Context) {
	id, ok := parseTagID(ctx)
	if !ok {
		return
	}
	var req v1.UpdateTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := t.update(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
