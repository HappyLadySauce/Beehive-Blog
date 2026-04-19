package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListRevisionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRevisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRevisionsLogic {
	return &ListRevisionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRevisionsLogic) ListRevisions(in *pb.RevisionListRequest) (*pb.RevisionListResponse, error) {
	out, err := l.svcCtx.Store.ListRevisions(l.ctx, in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
