package middlewares

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	maxAge = 12
)

// Cors add cors headers.
// 浏览器端需带具体 Origin；AllowCredentials 为 true 时不能使用 "*"，故用 AllowOriginFunc 按前缀放行本地与常见开发来源。
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:  []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:  []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		AllowOriginFunc: func(origin string) bool {
			if origin == "" {
				return true
			}
			if origin == "https://github.com" {
				return true
			}
			u, err := url.Parse(origin)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return false
			}
			host := strings.ToLower(u.Hostname())
			if host == "localhost" || host == "127.0.0.1" || host == "::1" {
				return true
			}
			return false
		},
		AllowCredentials: true,
		MaxAge:           maxAge * time.Hour,
	})
}
