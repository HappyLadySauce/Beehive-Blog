package gateway

import (
	"context"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteAttachmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAttachmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAttachmentLogic {
	return &DeleteAttachmentLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteAttachmentLogic) DeleteAttachment(req *types.AttachmentPathRequest) (err error) {
	if req == nil || req.Id <= 0 {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "attachment_id": req.Id}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.delete_attachment", "content.DeleteAttachment", startedAt, err, fields)
	}()

	_, err = l.svcCtx.Content.DeleteAttachment(l.ctx, &contentrpc.DeleteAttachmentRequest{Id: req.Id})
	return err
}
