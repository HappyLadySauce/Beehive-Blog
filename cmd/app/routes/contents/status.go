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
		return nil, common.NewNotFound("content not found", err)
	}

	if content.Status == req.Status {
		return c.get(ctx, id, true)
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
	return c.get(ctx, id, true)
}

// TransitionStatus handles PATCH /api/v1/contents/:id/status (admin).
// TransitionStatus 处理 PATCH /api/v1/contents/:id/status（管理员）。
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
