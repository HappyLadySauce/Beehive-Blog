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

type ListRelationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRelationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRelationsLogic {
	return &ListRelationsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListRelationsLogic) ListRelations(req *types.ContentRelationsPathRequest) (*types.RelationListResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	out, err := l.svcCtx.Content.ListRelations(l.ctx, &contentrpc.ListRelationsRequest{ContentId: req.Id})
	if err != nil {
		return nil, err
	}
	return toRelationListResponse(out), nil
}
