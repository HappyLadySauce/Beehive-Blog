// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(req *types.RefreshRequest) (resp *types.TokenData, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	fields := map[string]any{
		"refresh_token": maskToken(req.RefreshToken),
	}
	defer func() {
		if err != nil {
			auditFailure(l.ctx, "auth.refresh", err, fields)
			return
		}
		auditSuccess(l.ctx, "auth.refresh", fields)
	}()

	if strings.TrimSpace(req.RefreshToken) == "" {
		return nil, status.Error(codes.InvalidArgument, "refreshToken is required")
	}

	out, err := l.svcCtx.Identity.Refresh(l.ctx, &identityrpc.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return toTokenData(out), nil
}
