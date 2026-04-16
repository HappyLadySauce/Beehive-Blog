// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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

type CreateContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContentLogic {
	return &CreateContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateContentLogic) CreateContent(req *types.ContentCreateRequest) (resp *types.ContentDetail, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id": userID,
		"type":    req.ContentType,
		"slug":    req.Slug,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "content.create", "content.CreateContent", startedAt, err, fields)
	}()

	if strings.TrimSpace(req.ContentType) == "" || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Slug) == "" {
		return nil, status.Error(codes.InvalidArgument, "type/title/slug are required")
	}

	out, err := l.svcCtx.Content.CreateContent(l.ctx, &contentrpc.CreateContentRequest{
		Type:         req.ContentType,
		Title:        req.Title,
		Slug:         req.Slug,
		Summary:      req.Summary,
		BodyMarkdown: req.BodyMarkdown,
		Visibility:   req.Visibility,
		AiAccess:     req.AiAccess,
	})
	if err != nil {
		return nil, err
	}
	triggerAsyncIndexUpsert(l.ctx, l.svcCtx.Search, out)
	return toContentDetail(out), nil
}
