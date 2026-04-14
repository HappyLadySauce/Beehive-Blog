package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListPublicArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPublicArticlesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPublicArticlesLogic {
	return &ListPublicArticlesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPublicArticlesLogic) ListPublicArticles(in *pb.ListContentsRequest) (*pb.ListContentsResponse, error) {
	out, err := l.svcCtx.Store.List(l.ctx, in, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return out, nil
}
