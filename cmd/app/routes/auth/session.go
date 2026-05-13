package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

// Session returns claims from an access token already verified by AuthMiddleware.
// Session 返回已经由 AuthMiddleware 验证过的 access token 声明。
//
// @Summary      Get current auth session
// @Description  Verifies the presented access token and returns the signed session claims used by browser-side BFF guards. 中文：校验传入的 access token，并返回浏览器 BFF 守卫使用的已签名会话声明。
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  common.BaseResponse{data=v1.AuthSessionResponse}
// @Failure      401  {object}  common.BaseResponse
// @Router       /api/v1/auth/session [get]
func (a *AuthController) Session(ctx *gin.Context) {
	claims := middleware.GetClaims(ctx)
	if claims == nil || claims.UID <= 0 || claims.Role == "" || claims.ExpiresAt == nil {
		common.Fail(ctx, common.NewUnauthorized("invalid or expired access token", nil))
		return
	}

	common.Success(ctx, v1.AuthSessionResponse{
		UID:  claims.UID,
		Role: claims.Role,
		Exp:  claims.ExpiresAt.Unix(),
		SID:  claims.SID,
	})
}
