package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateContentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentLogic {
	return &UpdateContentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateContentLogic) UpdateContent(in *pb.UpdateContentRequest) (*pb.ContentDetail, error) {
	out, err := l.svcCtx.Store.Update(l.ctx, in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
