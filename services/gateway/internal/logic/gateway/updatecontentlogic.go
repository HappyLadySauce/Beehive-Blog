// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"
	"strings"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentLogic {
	return &UpdateContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateContentLogic) UpdateContent(req *types.ContentUpdateRequest) (resp *types.ContentDetail, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":    userID,
		"content_id": req.Id,
	}
	defer func() {
		if err != nil {
			auditFailure(l.ctx, "content.update", err, fields)
			return
		}
		auditSuccess(l.ctx, "content.update", fields)
	}()

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if strings.TrimSpace(req.Title) == "" &&
		strings.TrimSpace(req.Summary) == "" &&
		strings.TrimSpace(req.BodyMarkdown) == "" &&
		strings.TrimSpace(req.Visibility) == "" &&
		strings.TrimSpace(req.AiAccess) == "" {
		return nil, status.Error(codes.InvalidArgument, "no updates provided")
	}

	out, err := l.svcCtx.Content.UpdateContent(l.ctx, &contentrpc.UpdateContentRequest{
		Id:           req.Id,
		Title:        req.Title,
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
