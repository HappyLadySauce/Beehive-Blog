package settings

import (
	"fmt"

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

// Init validates shared handles and registers /api/v1/settings routes (admin only).
// Init 校验共享句柄并注册 /api/v1/settings 路由（仅管理员）。
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
	if svcCtx.Settings == nil {
		return fmt.Errorf("settings provider is nil")
	}
	if svcCtx.Token == nil {
		return fmt.Errorf("jwt issuer is nil")
	}

	h := NewSettingsController(svcCtx)
	g := router.V1().Group("/settings")
	g.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))

	email := g.Group("/email")
	email.GET("", h.GetEmailSettings)
	email.PATCH("", h.PatchEmailSettings)
	email.POST("/test", h.TestEmail)
	return nil
}
