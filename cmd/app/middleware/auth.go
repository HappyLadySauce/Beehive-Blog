package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth"
)

// Context keys set by AuthMiddleware for downstream handlers.
// AuthMiddleware 为下游处理器注入的 Context 键。
const (
	// ClaimsKey stores the parsed *auth.Claims.
	// ClaimsKey 存储解析后的 *auth.Claims。
	ClaimsKey = "claims"
	// UIDKey stores the authenticated user ID as int64.
	// UIDKey 存储已认证用户 ID（int64）。
	UIDKey = "uid"
	// RoleKey stores the authenticated user role as string.
	// RoleKey 存储已认证用户角色（string）。
	RoleKey = "role"
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

		ctx.Set(ClaimsKey, claims)
		ctx.Set(UIDKey, claims.UID)
		ctx.Set(RoleKey, claims.Role)

		ctx.Next()
	}
}

// GetClaims extracts the *auth.Claims from the Gin context; returns nil if absent.
// GetClaims 从 Gin 上下文提取 *auth.Claims；不存在时返回 nil。
func GetClaims(ctx *gin.Context) *auth.Claims {
	v, ok := ctx.Get(ClaimsKey)
	if !ok {
		return nil
	}
	claims, ok := v.(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
