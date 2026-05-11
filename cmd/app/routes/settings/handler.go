package settings

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// SettingsController serves admin settings HTTP endpoints.
// SettingsController 提供管理员设置 HTTP 接口。
type SettingsController struct {
	svc *svc.ServiceContext
}

// NewSettingsController constructs a SettingsController.
// NewSettingsController 构造 SettingsController。
func NewSettingsController(svcCtx *svc.ServiceContext) *SettingsController {
	return &SettingsController{svc: svcCtx}
}

// Init registers /api/v1/settings routes (admin only).
// Init 注册 /api/v1/settings 路由（仅管理员）。
func Init(svcCtx *svc.ServiceContext) {
	h := NewSettingsController(svcCtx)
	g := router.V1().Group("/settings")
	g.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))

	email := g.Group("/email")
	email.GET("", h.GetEmailSettings)
	email.PATCH("", h.PatchEmailSettings)
	email.POST("/test", h.TestEmail)
}
