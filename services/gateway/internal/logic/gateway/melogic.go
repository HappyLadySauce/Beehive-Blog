// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MeLogic) Me() (resp *types.UserProfile, err error) {
	userID, err := parseAccessTokenUserIDFromContext(l.ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	out, err := l.svcCtx.Identity.GetUser(l.ctx, &identityrpc.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return toUserProfile(out), nil
}
