// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"

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
	if req == nil || strings.TrimSpace(req.ContentType) == "" || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Slug) == "" {
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
	return toContentDetail(out), nil
}
