package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *HealthLogic) Health(in *pb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Service: "indexer",
		Status:  "ok",
	}, nil
}
