package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateContentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentStatusLogic {
	return &UpdateContentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateContentStatusLogic) UpdateContentStatus(in *pb.UpdateStatusRequest) (*pb.ContentDetail, error) {
	out, err := l.svcCtx.Store.UpdateStatus(l.ctx, in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
