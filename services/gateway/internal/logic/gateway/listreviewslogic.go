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

type ListReviewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListReviewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListReviewsLogic {
	return &ListReviewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListReviewsLogic) ListReviews(req *types.ReviewListRequest) (resp *types.ReviewListResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":   userID,
		"status":    req.Status,
		"page":      req.Page,
		"page_size": req.PageSize,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.list_reviews", "content.ListReviews", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.ListReviews(l.ctx, &contentrpc.ReviewListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	return toReviewListResponse(out), nil
}
