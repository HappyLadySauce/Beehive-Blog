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

type UpdateCommentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCommentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentStatusLogic {
	return &UpdateCommentStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateCommentStatusLogic) UpdateCommentStatus(req *types.UpdateCommentStatusRequest) (resp *types.Comment, err error) {
	if req == nil || req.Id <= 0 || strings.TrimSpace(req.Status) == "" {
		return nil, status.Error(codes.InvalidArgument, "id and status are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "comment_id": req.Id, "status": req.Status}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.update_comment_status", "content.UpdateCommentStatus", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.UpdateCommentStatus(l.ctx, &contentrpc.UpdateCommentStatusRequest{
		Id:             req.Id,
		Status:         req.Status,
		ModerationNote: req.ModerationNote,
	})
	if err != nil {
		return nil, err
	}
	return toComment(out), nil
}
