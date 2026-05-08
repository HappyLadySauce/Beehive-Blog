package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Refresh mints a new access token from a valid refresh token (refresh JWT is not rotated here).
// Refresh 使用有效 refresh 令牌签发新的 access 令牌（此处不轮换 refresh JWT）。
func (a *AuthController) Refresh(ctx *gin.Context, req *v1.RefreshRequest) (*v1.RefreshResponse, error) {
	claims, err := a.svc.Token.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, common.NewUnauthorized("invalid or expired refresh token", nil)
	}

	// Refresh tokens must be re-authorized against current user state before minting access.
	// 刷新令牌签发 access 前必须基于当前用户状态重新授权。
	var user model.User
	if err := a.svc.DB.First(&user, claims.UID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NewUnauthorized("invalid or expired refresh token", nil)
		}
		return nil, common.NewInternal("failed to refresh token", fmt.Errorf("query user: %w", err))
	}
	if err := assertUserMayLogin(&user); err != nil {
		return nil, err
	}

	access, err := a.svc.Token.IssueAccess(user.ID, user.Role)
	if err != nil {
		return nil, common.NewInternal("failed to issue access token", err)
	}

	return &v1.RefreshResponse{
		Token: v1.AuthToken{
			AccessToken: access.Token,
			TokenType:   jwt.TokenTypeBearer,
			ExpiresIn:   access.ExpiresIn,
		},
	}, nil
}
