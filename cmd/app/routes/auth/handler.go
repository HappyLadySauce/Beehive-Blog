package auth

import (
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// AuthController handles HTTP routes for authentication.
// AuthController 处理认证相关 HTTP 路由。
type AuthController struct {
	svc *svc.ServiceContext
}

// NewAuthController builds a AuthController bound to the given service context.
// NewAuthController 基于给定 ServiceContext 构造 AuthController。
func NewAuthController(svcCtx *svc.ServiceContext) *AuthController {
	return &AuthController{
		svc: svcCtx,
	}
}

// Init validates shared handles and registers HTTP routes for the auth domain.
// Init 校验共享句柄并注册 auth 域的 HTTP 路由。
func Init(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return fmt.Errorf("service context is nil")
	}
	if svcCtx.Config == nil {
		return fmt.Errorf("config is nil")
	}
	if svcCtx.DB == nil {
		return fmt.Errorf("database handle is nil")
	}
	if svcCtx.Cache == nil {
		return fmt.Errorf("cache client is nil")
	}
	if svcCtx.Token == nil {
		return fmt.Errorf("jwt issuer is nil")
	}

	auth := NewAuthController(svcCtx)
	// ~20 requests/min sustained per IP with short bursts for login/OAuth/refresh.
	// 约每 IP 每分钟 20 次可持续速率，并允许短时突发（登录 / OAuth / 刷新）。
	rl := middleware.NewAuthPublicRateLimiter(20.0/60.0, 25)

	authGroup := router.V1().Group("/auth")

	authGroup.GET("/github/authorize", rl.GinMiddleware(), auth.GithubOAuthBegin)
	authGroup.POST("/login", rl.GinMiddleware(), auth.Login)
	authGroup.POST("/refresh", rl.GinMiddleware(), auth.Refresh)
	authGroup.GET("/session", middleware.AuthMiddleware(svcCtx), auth.Session)
	authGroup.POST("/logout", middleware.AuthMiddleware(svcCtx), auth.Logout)
	return nil
}
