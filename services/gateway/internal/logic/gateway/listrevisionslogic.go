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

type ListRevisionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRevisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRevisionsLogic {
	return &ListRevisionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRevisionsLogic) ListRevisions(req *types.RevisionListRequest) (resp *types.RevisionListResponse, err error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "content id is required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":    userID,
		"content_id": req.Id,
		"page":       req.Page,
		"page_size":  req.PageSize,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.list_revisions", "content.ListRevisions", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.ListRevisions(l.ctx, &contentrpc.RevisionListRequest{
		ContentId: req.Id,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return toRevisionListResponse(out), nil
}
