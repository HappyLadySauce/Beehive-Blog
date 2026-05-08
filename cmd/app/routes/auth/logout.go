package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
)

// Logout revokes the current access-token-bound session; repeated calls are safe.
// Logout 撤销当前 access token 绑定的会话；重复调用是安全的。
func (a *AuthController) Logout(ctx *gin.Context) {
	claims := jwt.GetClaims(ctx)
	if claims == nil || claims.UID <= 0 || claims.SID <= 0 {
		common.Fail(ctx, common.NewUnauthorized("invalid or expired access token", nil))
		return
	}
	if err := authsession.RevokeSession(a.svc.DB, claims.SID, claims.UID, "logout"); err != nil {
		common.Fail(ctx, common.NewInternal("failed to logout", err))
		return
	}
	common.Success(ctx, gin.H{})
}
