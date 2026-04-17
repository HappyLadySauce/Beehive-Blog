package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/worker"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOutboxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOutboxLogic {
	return &SyncOutboxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncOutboxLogic) SyncOutbox(in *pb.SyncOutboxRequest) (*pb.SyncOutboxResponse, error) {
	w := worker.NewIndexerWorker(l.svcCtx)
	result, err := w.ProcessBatch(l.ctx, in.GetBatchSize())
	if err != nil {
		return nil, err
	}

	return &pb.SyncOutboxResponse{
		Picked:    result.Picked,
		Succeeded: result.Succeeded,
		Failed:    result.Failed,
	}, nil
}
