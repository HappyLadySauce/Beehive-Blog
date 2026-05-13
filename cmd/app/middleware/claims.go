package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

// GetClaims extracts the *jwt.Claims injected by AuthMiddleware.
// GetClaims 提取 AuthMiddleware 注入的 *jwt.Claims。
func GetClaims(ctx *gin.Context) *jwt.Claims {
	v, ok := ctx.Get(jwt.ClaimsKey)
	if !ok {
		return nil
	}
	claims, ok := v.(*jwt.Claims)
	if !ok {
		return nil
	}
	return claims
}
