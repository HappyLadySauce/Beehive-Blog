package users

import (
	"fmt"

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

// Init validates shared handles and registers HTTP routes for the users domain.
// Init 校验共享句柄并注册 users 域的 HTTP 路由。
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
	if svcCtx.Token == nil {
		return fmt.Errorf("jwt issuer is nil")
	}

	u := NewUsersController(svcCtx)
	rl := middleware.NewAuthPublicRateLimiter(10.0/60.0, 12)

	users := router.V1().Group("/users")

	users.POST("/register", rl.GinMiddleware(), u.Register)

	adminUsers := users.Group("")
	adminUsers.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminUsers.GET("", u.List)
	adminUsers.POST("", u.Create)
	adminUsers.GET("/:id", u.Get)
	adminUsers.PATCH("/:id", u.Update)
	adminUsers.DELETE("/:id", u.Delete)

	return nil
}
