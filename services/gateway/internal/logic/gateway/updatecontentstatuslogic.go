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

type UpdateContentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateContentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentStatusLogic {
	return &UpdateContentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateContentStatusLogic) UpdateContentStatus(req *types.StatusUpdateRequest) (resp *types.ContentDetail, err error) {
	if req == nil || req.Id <= 0 || strings.TrimSpace(req.Status) == "" {
		return nil, status.Error(codes.InvalidArgument, "id and status are required")
	}

	out, err := l.svcCtx.Content.UpdateContentStatus(l.ctx, &contentrpc.UpdateStatusRequest{
		Id:     req.Id,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return toContentDetail(out), nil
}
