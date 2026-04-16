// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.TokenData, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	fields := map[string]any{
		"account": maskAccount(req.Account),
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "auth.login", "identity.Login", startedAt, err, fields)
	}()

	if strings.TrimSpace(req.Account) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "account and password are required")
	}

	out, err := l.svcCtx.Identity.Login(l.ctx, &identityrpc.LoginRequest{
		Account:  req.Account,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return toTokenData(out), nil
}
