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

	// Public routes — uses optional actor to expand scope for admins.
	// 公开路由——通过 optionalActor 为管理员扩展数据范围。
	contents := router.V1().Group("/contents")
	contents.GET("", c.List)
	contents.GET("/:id", c.Get)
	contents.GET("/:id/relations", c.GetRelations)
	contents.GET("/:id/tags", c.GetContentTags)

	// Admin-only routes.
	// 管理员专用路由。
	adminContents := contents.Group("")
	adminContents.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminContents.POST("", c.Create)
	adminContents.PATCH("/:id", c.Update)
	adminContents.DELETE("/:id", c.Delete)
	adminContents.PATCH("/:id/status", c.TransitionStatus)
	adminContents.GET("/:id/versions", c.ListVersions)
	adminContents.POST("/:id/versions", c.CreateVersion)
	adminContents.POST("/:id/relations", c.AddRelation)
	adminContents.DELETE("/:id/relations/:relationId", c.RemoveRelation)
	adminContents.PUT("/:id/tags", c.SetTags)

	return nil
}
