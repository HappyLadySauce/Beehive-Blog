package categories

import (
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// CategoriesController handles HTTP routes for content categories.
// CategoriesController 处理内容分类相关 HTTP 路由。
type CategoriesController struct {
	svc *svc.ServiceContext
}

// NewCategoriesController builds a CategoriesController bound to the given service context.
// NewCategoriesController 基于给定 ServiceContext 构造 CategoriesController。
func NewCategoriesController(svcCtx *svc.ServiceContext) *CategoriesController {
	return &CategoriesController{svc: svcCtx}
}

// Init validates shared handles and registers HTTP routes for the categories domain.
// Init 校验共享句柄并注册 categories 域的 HTTP 路由。
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

	c := NewCategoriesController(svcCtx)

	publicCategories := router.V1().Group("/categories")
	publicCategories.Use(middleware.OptionalAuthMiddleware(svcCtx))
	publicCategories.GET("", c.List)
	publicCategories.GET("/:id", c.Get)

	adminCategories := router.V1().Group("/categories")
	adminCategories.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminCategories.POST("", c.Create)
	adminCategories.PATCH("/:id", c.Update)
	adminCategories.DELETE("/:id", c.Delete)

	return nil
}
