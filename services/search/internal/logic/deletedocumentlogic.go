package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDocumentLogic {
	return &DeleteDocumentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDocumentLogic) DeleteDocument(in *pb.DeleteDocumentRequest) (*pb.Empty, error) {
	if err := l.svcCtx.Store.DeleteDocument(l.ctx, in.GetContentId()); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
