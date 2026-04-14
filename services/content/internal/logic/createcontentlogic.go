package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContentLogic {
	return &CreateContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateContentLogic) CreateContent(in *pb.CreateContentRequest) (*pb.ContentDetail, error) {
	out, err := l.svcCtx.Store.Create(l.ctx, in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
