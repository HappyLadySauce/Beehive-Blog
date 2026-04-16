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

type CreateCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateCommentLogic) CreateComment(req *types.CreateCommentRequest) (resp *types.Comment, err error) {
	if req == nil || req.ContentId <= 0 || strings.TrimSpace(req.BodyMarkdown) == "" {
		return nil, status.Error(codes.InvalidArgument, "contentId and bodyMarkdown are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "content_id": req.ContentId}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.create_comment", "content.CreateComment", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.CreateComment(l.ctx, &contentrpc.CreateCommentRequest{
		ContentId:       req.ContentId,
		ParentCommentId: req.ParentCommentId,
		AuthorName:      req.AuthorName,
		AuthorEmail:     req.AuthorEmail,
		BodyMarkdown:    req.BodyMarkdown,
	})
	if err != nil {
		return nil, err
	}
	return toComment(out), nil
}
