package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

// AuthMiddleware returns a Gin handler that validates the Bearer token
// and injects parsed claims into the request context.
// AuthMiddleware 返回验证 Bearer 令牌并将解析结果注入请求上下文的 Gin 处理器。
func AuthMiddleware(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
				"data":    nil,
			})
			return
		}

		if !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "authorization header must use Bearer scheme",
				"data":    nil,
			})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "empty bearer token",
				"data":    nil,
			})
			return
		}

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
