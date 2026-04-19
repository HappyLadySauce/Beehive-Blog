// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package gateway

import (
	"context"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubmitReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReviewLogic {
	return &SubmitReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReviewLogic) SubmitReview(req *types.ContentReviewSubmitRequest) (resp *types.ReviewTask, err error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "content id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":    userID,
		"content_id": req.Id,
		"priority":   req.Priority,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.submit_review", "content.SubmitReview", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.SubmitReview(l.ctx, &contentrpc.SubmitReviewRequest{
		ContentId:       req.Id,
		Note:            req.Note,
		Priority:        req.Priority,
		SubmitterUserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return toReviewTask(out), nil
}
