package logic

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpsertDocumentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpsertDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertDocumentLogic {
	return &UpsertDocumentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpsertDocumentLogic) UpsertDocument(in *pb.UpsertDocumentRequest) (*pb.IndexDocument, error) {
	if in != nil {
		in.Title = strings.TrimSpace(in.GetTitle())
		in.Slug = strings.TrimSpace(in.GetSlug())
	}
	return l.svcCtx.Store.UpsertDocument(l.ctx, in)
}
