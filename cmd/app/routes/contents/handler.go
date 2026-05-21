package contents

import (
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// ContentsController handles HTTP routes for the content domain.
// ContentsController 处理内容域相关 HTTP 路由。
type ContentsController struct {
	svc *svc.ServiceContext
}

// NewContentsController builds a ContentsController bound to the given service context.
// NewContentsController 基于给定 ServiceContext 构造 ContentsController。
func NewContentsController(svcCtx *svc.ServiceContext) *ContentsController {
	return &ContentsController{svc: svcCtx}
}

// Init validates shared handles and registers HTTP routes for the contents domain.
// Init 校验共享句柄并注册 contents 域的 HTTP 路由。
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

	c := NewContentsController(svcCtx)

	// Public routes — OptionalAuthMiddleware injects claims for admin preview.
	// 公开路由——OptionalAuthMiddleware 为管理员预览注入 claims。
	publicContents := router.V1().Group("/contents")
	publicContents.Use(middleware.OptionalAuthMiddleware(svcCtx))
	publicContents.GET("", c.List)
	publicContents.GET("/:id", c.Get)
	publicContents.GET("/:id/relations", c.GetRelations)
	publicContents.GET("/:id/tags", c.GetContentTags)
	publicContents.GET("/:id/categories", c.GetContentCategories)

	// Admin-only routes (separate group avoids stacking with OptionalAuth).
	// 管理员专用路由（独立分组，避免与 OptionalAuth 叠加）。
	adminContents := router.V1().Group("/contents")
	adminContents.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminContents.POST("", c.Create)
	adminContents.PATCH("/:id", c.Update)
	adminContents.DELETE("/:id", c.Delete)
	adminContents.PATCH("/:id/status", c.TransitionStatus)
	adminContents.GET("/:id/versions", c.ListVersions)
	adminContents.POST("/:id/versions", c.CreateVersion)
	adminContents.DELETE("/:id/versions/:versionNumber", c.DeleteVersion)
	adminContents.POST("/:id/versions/:versionNumber/restore", c.RestoreVersion)
	adminContents.POST("/:id/relations", c.AddRelation)
	adminContents.DELETE("/:id/relations/:relationId", c.RemoveRelation)
	adminContents.PUT("/:id/tags", c.SetTags)
	adminContents.PUT("/:id/categories", c.SetCategories)

	return nil
}
