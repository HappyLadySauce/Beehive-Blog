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

type GetRevisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRevisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRevisionLogic {
	return &GetRevisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRevisionLogic) GetRevision(req *types.RevisionPathRequest) (resp *types.RevisionDetail, err error) {
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
		auditRPC(l.ctx, "content.get_revision", "content.GetRevision", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.GetRevision(l.ctx, &contentrpc.GetRevisionRequest{
		ContentId:  req.Id,
		RevisionId: req.RevisionId,
	})
	if err != nil {
		return nil, err
	}
	return toRevisionDetail(out), nil
}
