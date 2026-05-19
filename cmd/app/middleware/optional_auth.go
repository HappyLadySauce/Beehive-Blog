package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

// OptionalAuthMiddleware parses a Bearer token when present and injects claims.
// Missing Authorization is allowed; invalid or malformed tokens abort with 401.
// OptionalAuthMiddleware 在存在 Bearer 时解析并注入 claims；无 Authorization 放行；无效令牌返回 401。
func OptionalAuthMiddleware(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if header == "" {
			ctx.Next()
			return
		}

		scheme, tokenString, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenString) == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header must use Bearer scheme",
				"data":    nil,
			})
			return
		}
		tokenString = strings.TrimSpace(tokenString)

		claims, err := svcCtx.Token.ParseAccess(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid or expired access token",
				"data":    nil,
			})
			return
		}

		ctx.Set(jwt.ClaimsKey, claims)
		ctx.Set(jwt.UIDKey, claims.UID)
		ctx.Set(jwt.RoleKey, claims.Role)
		ctx.Next()
	}
}
