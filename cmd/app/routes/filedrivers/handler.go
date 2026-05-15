package filedrivers

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
)

// FileDriversController serves storage driver and mount HTTP endpoints.
// FileDriversController 提供存储驱动与挂载项的 HTTP 接口。
type FileDriversController struct {
	svc            *svc.ServiceContext
	db             *gorm.DB
	driverStore    *driver.Store
	driverRegistry *driver.DriverRegistry
}

// Init registers all file driver and storage mount routes.
// Init 注册所有文件驱动与存储挂载项路由。
func Init(svcCtx *svc.ServiceContext) error {
	h, err := NewFileDriversController(svcCtx)
	if err != nil {
		return err
	}

	drivers := router.V1().Group("/file-drivers")
	drivers.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	drivers.GET("", h.ListDrivers)
	drivers.GET("/:name", h.GetDriver)

	mounts := router.V1().Group("/storage-mounts")
	mounts.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	mounts.GET("", h.ListMounts)
	mounts.POST("", h.CreateMount)
	mounts.GET("/:id", h.GetMount)
	mounts.PATCH("/:id", h.PatchMount)
	mounts.POST("/:id/enable", h.EnableMount)
	mounts.POST("/:id/disable", h.DisableMount)
	mounts.POST("/:id/check", h.CheckMount)
	mounts.DELETE("/:id", h.DeleteMount)

	return nil
}

// NewFileDriversController builds a FileDriversController from the service context.
// NewFileDriversController 基于 ServiceContext 构造 FileDriversController。
func NewFileDriversController(svcCtx *svc.ServiceContext) (*FileDriversController, error) {
	if err := validateDeps(svcCtx); err != nil {
		return nil, err
	}
	return &FileDriversController{
		svc:            svcCtx,
		db:             svcCtx.DB,
		driverStore:    svcCtx.DriverStore,
		driverRegistry: svcCtx.DriverRegistry,
	}, nil
}

func validateDeps(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return fmt.Errorf("service context is nil")
	}
	if svcCtx.DB == nil {
		return fmt.Errorf("database handle is nil")
	}
	if svcCtx.DriverStore == nil {
		return fmt.Errorf("driver store is nil")
	}
	if svcCtx.DriverRegistry == nil {
		return fmt.Errorf("driver registry is nil")
	}
	return nil
}

// parseIDParam parses the :id URL parameter as int64.
// parseIDParam 将 :id URL 参数解析为 int64。
func parseIDParam(param string) (int64, error) {
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", param)
	}
	return id, nil
}
