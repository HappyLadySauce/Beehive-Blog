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

type ApproveReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveReviewLogic {
	return &ApproveReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveReviewLogic) ApproveReview(req *types.ReviewDecisionRequest) (resp *types.ReviewTask, err error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "review id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":   userID,
		"review_id": req.Id,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.approve_review", "content.ApproveReview", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.ApproveReview(l.ctx, &contentrpc.ApproveReviewRequest{
		Id:             req.Id,
		Reason:         req.Reason,
		ReviewerUserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return toReviewTask(out), nil
}
