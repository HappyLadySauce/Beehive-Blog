package auth

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/httpx"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// UsersController handles HTTP routes for users.
// UsersController 处理用户相关 HTTP 路由。
type AuthController struct {
	svc svc.ServiceContext
}

// NewAuthController builds a AuthController bound to the given service context.
// NewAuthController 基于给定 ServiceContext 构造 AuthController。
func NewAuthController(svcCtx *svc.ServiceContext) *AuthController {
	return &AuthController{
		svc: *svcCtx,
	}
}

// Init registers HTTP routes for the auth domain.
// Init 注册 auth 域的 HTTP 路由。
func Init(svcCtx *svc.ServiceContext) {
	auth := NewAuthController(svcCtx)

	authGroup := router.V1().Group("/auth")

	authGroup.POST("/login", httpx.HandleJSON(auth.Login))
}
