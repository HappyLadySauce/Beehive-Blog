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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.TokenData, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	fields := map[string]any{
		"username": maskAccount(req.Username),
		"email":    maskEmail(req.Email),
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "auth.register", "identity.Register", startedAt, err, fields)
	}()

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email and password are required")
	}

	out, err := l.svcCtx.Identity.Register(l.ctx, &identityrpc.RegisterRequest{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return toTokenData(out), nil
}
