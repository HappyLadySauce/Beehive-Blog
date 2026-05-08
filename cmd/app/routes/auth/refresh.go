package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

// Refresh mints a new access token from a valid refresh token (refresh JWT is not rotated here).
// Refresh 使用有效 refresh 令牌签发新的 access 令牌（此处不轮换 refresh JWT）。
func (a *AuthController) Refresh(ctx *gin.Context, req *v1.RefreshRequest) (*v1.RefreshResponse, error) {
	claims, err := a.svc.Token.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	access, err := a.svc.Token.IssueAccess(claims.UID, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	return &v1.RefreshResponse{
		Token: v1.AuthToken{
			AccessToken: access.Token,
			TokenType:   jwt.TokenTypeBearer,
			ExpiresIn:   access.ExpiresIn,
		},
	}, nil
}
