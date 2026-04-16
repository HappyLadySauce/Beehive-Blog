package logic

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListAttachmentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAttachmentsLogic {
	return &ListAttachmentsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListAttachmentsLogic) ListAttachments(in *pb.ListAttachmentsRequest) (*pb.ListAttachmentsResponse, error) {
	out, err := l.svcCtx.Store.ListAttachments(l.ctx, in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}
