package settings

import (
	"context"
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
)

// SettingsController serves admin settings HTTP endpoints.
// SettingsController 提供管理员设置 HTTP 接口。
type SettingsController struct {
	svc      *svc.ServiceContext
	store    *pkgsettings.Store
	provider *pkgsettings.Provider
}

// NewSettingsController constructs a SettingsController and loads the initial settings snapshot.
// NewSettingsController 构造 SettingsController 并加载初始设置快照。
func NewSettingsController(ctx context.Context, svcCtx *svc.ServiceContext) (*SettingsController, error) {
	if err := validateDependencies(svcCtx); err != nil {
		return nil, err
	}
	seed, err := svcCtx.Config.Email.ToApplicationSettings()
	if err != nil {
		return nil, fmt.Errorf("email options: %w", err)
	}

	store := pkgsettings.NewStore(svcCtx.DB)
	if err := store.EnsureSingleton(ctx, seed); err != nil {
		return nil, fmt.Errorf("ensure application settings: %w", err)
	}
	h := &SettingsController{
		svc:      svcCtx,
		store:    store,
		provider: pkgsettings.NewProvider(),
	}
	if err := h.refresh(ctx); err != nil {
		return nil, fmt.Errorf("refresh application settings: %w", err)
	}
	return h, nil
}

// Init validates shared handles, initializes route-local settings services, and registers /api/v1/settings routes.
// Init 校验共享句柄、初始化 settings 路由本地服务，并注册 /api/v1/settings 路由。
func Init(ctx context.Context, svcCtx *svc.ServiceContext) error {
	h, err := NewSettingsController(ctx, svcCtx)
	if err != nil {
		return err
	}

	if svcCtx.PostgresDSN != "" {
		pkgsettings.StartNotifyListener(ctx, svcCtx.PostgresDSN, h.refresh)
	}

	g := router.V1().Group("/settings")
	g.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))

	email := g.Group("/email")
	email.GET("", h.GetEmailSettings)
	email.PATCH("", h.PatchEmailSettings)
	email.POST("/test", h.TestEmail)
	return nil
}

func validateDependencies(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return fmt.Errorf("service context is nil")
	}
	if svcCtx.Config == nil {
		return fmt.Errorf("config is nil")
	}
	if svcCtx.DB == nil {
		return fmt.Errorf("database handle is nil")
	}
	if svcCtx.Config.Email == nil {
		return fmt.Errorf("email config is nil")
	}
	if svcCtx.Token == nil {
		return fmt.Errorf("jwt issuer is nil")
	}
	return nil
}

func (h *SettingsController) refresh(ctx context.Context) error {
	s, rev, err := h.store.Load(ctx)
	if err != nil {
		return err
	}
	h.provider.Replace(s, rev)
	return nil
}
