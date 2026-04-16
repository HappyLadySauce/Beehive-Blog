package gateway

import (
	"context"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListAttachmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAttachmentsLogic {
	return &ListAttachmentsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListAttachmentsLogic) ListAttachments(req *types.AttachmentListRequest) (*types.AttachmentListResponse, error) {
	out, err := l.svcCtx.Content.ListAttachments(l.ctx, &contentrpc.ListAttachmentsRequest{
		ContentId: req.ContentId,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return toAttachmentListResponse(out), nil
}
