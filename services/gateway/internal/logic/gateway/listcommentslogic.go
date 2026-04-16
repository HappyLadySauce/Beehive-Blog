package gateway

import (
	"context"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentsLogic {
	return &ListCommentsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListCommentsLogic) ListComments(req *types.CommentListRequest) (*types.CommentListResponse, error) {
	if req == nil || req.ContentId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "contentId is required")
	}
	out, err := l.svcCtx.Content.ListComments(l.ctx, &contentrpc.ListCommentsRequest{
		ContentId: req.ContentId,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return toCommentListResponse(out), nil
}
