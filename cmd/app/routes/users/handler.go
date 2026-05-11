package users

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// UsersController handles HTTP routes for users.
// UsersController 处理用户相关 HTTP 路由。
type UsersController struct {
	svc *svc.ServiceContext
}

// NewUsersController builds a UsersController bound to the given service context.
// NewUsersController 基于给定 ServiceContext 构造 UsersController。
func NewUsersController(svcCtx *svc.ServiceContext) *UsersController {
	return &UsersController{
		svc: svcCtx,
	}
}

// Init registers HTTP routes for the users domain.
// Init 注册 users 域的 HTTP 路由。
func Init(svcCtx *svc.ServiceContext) {
	u := NewUsersController(svcCtx)
	rl := middleware.NewAuthPublicRateLimiter(10.0/60.0, 12)

	users := router.V1().Group("/users")

	users.POST("/register", rl.GinMiddleware(), u.Register)
}
