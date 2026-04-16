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

type CreateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateTagLogic) CreateTag(req *types.CreateTagRequest) (resp *types.Tag, err error) {
	if req == nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Slug) == "" {
		return nil, status.Error(codes.InvalidArgument, "name and slug are required")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{"user_id": userID, "slug": req.Slug}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.create_tag", "content.CreateTag", startedAt, err, fields)
	}()

	out, err := l.svcCtx.Content.CreateTag(l.ctx, &contentrpc.CreateTagRequest{
		Name:        req.Name,
		Slug:        req.Slug,
		Color:       req.Color,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return toTag(out), nil
}
