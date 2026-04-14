package logic

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/identity/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/identity/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginRequest) (*pb.TokenReply, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if strings.TrimSpace(in.Account) == "" || strings.TrimSpace(in.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "account and password are required")
	}

	user, accessToken, refreshToken, expiresIn, err := l.svcCtx.Store.Login(in.Account, in.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: &pb.UserProfile{
			Id:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}
