package contents

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// transitionStatus validates and executes a content status transition.
// transitionStatus 校验并执行内容状态流转。
func (c *ContentsController) transitionStatus(ctx context.Context, id int64, req *v1.TransitionStatusRequest) (*v1.ContentDetailResponse, error) {
	var content model.Content
	if err := c.svc.DB.WithContext(ctx).First(&content, id).Error; err != nil {
		return nil, mapFirstError(err, "content not found", "failed to fetch content")
	}

	if content.Status == req.Status {
		resp, err := c.get(ctx, id, true)
		if err != nil {
			return nil, err
		}
		return resp.(*v1.ContentDetailResponse), nil
	}

	if !validStatusTransition(content.Status, req.Status) {
		return nil, common.NewBadRequest(
			fmt.Sprintf("invalid status transition from %q to %q", content.Status, req.Status),
			fmt.Errorf("status transition: %s -> %s", content.Status, req.Status),
		)
	}

	updates := map[string]interface{}{"status": req.Status}

	if req.Status == "published" && content.PublishedAt == nil {
		now := time.Now()
		updates["published_at"] = now
	}

	if err := c.svc.DB.WithContext(ctx).Model(&content).Updates(updates).Error; err != nil {
		return nil, common.NewInternal("failed to update content status", err)
	}
	resp, err := c.get(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return resp.(*v1.ContentDetailResponse), nil
}

// TransitionStatus handles PATCH /api/v1/contents/:id/status (admin).
// TransitionStatus 处理 PATCH /api/v1/contents/:id/status（管理员）。
//
//	@Summary		Transition content status
//	@Description	Transitions a content item's status (e.g., draft → review → published). Admin only. 中文：流转内容状态（如草稿→审核→发布），仅管理员。
//	@Tags			contents
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Content ID"
//	@Param			body	body		v1.TransitionStatusRequest	true	"New status (draft, review, published, archived)"
//	@Success		200		{object}	common.BaseResponse{data=v1.ContentDetailResponse}
//	@Failure		400		{object}	common.BaseResponse	"Invalid status transition"
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Router			/api/v1/contents/{id}/status [patch]
func (c *ContentsController) TransitionStatus(ctx *gin.Context) {
	id, ok := parseContentID(ctx)
	if !ok {
		return
	}
	var req v1.TransitionStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	resp, err := c.transitionStatus(ctx.Request.Context(), id, &req)
	if err != nil {
		common.Fail(ctx, err)
		return
	}
	common.Success(ctx, resp)
}
