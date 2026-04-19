// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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

type RestoreRevisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreRevisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreRevisionLogic {
	return &RestoreRevisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreRevisionLogic) RestoreRevision(req *types.RevisionPathRequest) (resp *types.ContentDetail, err error) {
	if req == nil || req.Id <= 0 || req.RevisionId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "content id and revision id are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":     userID,
		"content_id":  req.Id,
		"revision_id": req.RevisionId,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.restore_revision", "content.RestoreRevision", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.RestoreRevision(l.ctx, &contentrpc.RestoreRevisionRequest{
		ContentId:  req.Id,
		RevisionId: req.RevisionId,
	})
	if err != nil {
		return nil, err
	}
	triggerAsyncIndexUpsert(l.ctx, l.svcCtx.Search, out)
	return toContentDetail(out), nil
}
