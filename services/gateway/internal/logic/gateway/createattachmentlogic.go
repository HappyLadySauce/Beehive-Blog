package gateway

import (
	"context"
	"strings"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateAttachmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAttachmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAttachmentLogic {
	return &CreateAttachmentLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateAttachmentLogic) CreateAttachment(req *types.CreateAttachmentRequest) (resp *types.Attachment, err error) {
	if req == nil || req.ContentId <= 0 || strings.TrimSpace(req.ObjectKey) == "" || strings.TrimSpace(req.FileName) == "" {
		return nil, status.Error(codes.InvalidArgument, "contentId,objectKey,fileName are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "content_id": req.ContentId, "file_name": maskText(req.FileName, 2, 2)}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.create_attachment", "content.CreateAttachment", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.CreateAttachment(l.ctx, &contentrpc.CreateAttachmentRequest{
		ContentId:       req.ContentId,
		StorageProvider: req.StorageProvider,
		Bucket:          req.Bucket,
		ObjectKey:       req.ObjectKey,
		FileName:        req.FileName,
		MimeType:        req.MimeType,
		Ext:             req.Ext,
		SizeBytes:       req.SizeBytes,
		UsageType:       req.UsageType,
	})
	if err != nil {
		return nil, err
	}
	return toAttachment(out), nil
}
