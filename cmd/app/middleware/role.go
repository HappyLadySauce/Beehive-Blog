package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
)

// RequireRole returns middleware that allows the request only if claims carry the given role.
// Must run after AuthMiddleware so jwt claims exist.
// RequireRole 返回中间件：仅当 claims 携带指定角色时放行；须在 AuthMiddleware 之后执行。
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.GetClaims(c)
		if claims == nil || claims.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "forbidden",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
