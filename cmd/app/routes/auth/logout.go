package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	authsession "github.com/HappyLadySauce/Beehive-Blog/pkg/auth/session"
)

// Logout revokes the current access-token-bound session; repeated calls are safe.
// Logout 撤销当前 access token 绑定的会话；重复调用是安全的。
//
// @Summary      Logout (revoke current session)
// @Description  Revokes the server-side session tied to the presented access token. Reuse detection on refresh tokens is unaffected. 中文：撤销当前 access token 对应的服务端会话；与 refresh 轮换检测无关。
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse  "data is an empty JSON object"
// @Failure      401  {object}  common.BaseResponse
// @Failure      500  {object}  common.BaseResponse
// @Router       /api/v1/auth/logout [post]
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
