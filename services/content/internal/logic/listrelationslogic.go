package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListRelationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRelationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRelationsLogic {
	return &ListRelationsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListRelationsLogic) ListRelations(in *pb.ListRelationsRequest) (*pb.ListRelationsResponse, error) {
	if in == nil || in.ContentId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "content_id is required")
	}
	out, err := l.svcCtx.Store.ListRelations(l.ctx, in.ContentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return out, nil
}
