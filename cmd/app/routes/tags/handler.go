package tags

import (
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

// TagsController handles HTTP routes for content tags.
// TagsController 处理内容标签相关 HTTP 路由。
type TagsController struct {
	svc *svc.ServiceContext
}

// NewTagsController builds a TagsController bound to the given service context.
// NewTagsController 基于给定 ServiceContext 构造 TagsController。
func NewTagsController(svcCtx *svc.ServiceContext) *TagsController {
	return &TagsController{svc: svcCtx}
}

// Init validates shared handles and registers HTTP routes for the tags domain.
// Init 校验共享句柄并注册 tags 域的 HTTP 路由。
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

	t := NewTagsController(svcCtx)

	tags := router.V1().Group("/tags")
	tags.GET("", t.List)
	tags.GET("/:id", t.Get)

	adminTags := tags.Group("")
	adminTags.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminTags.POST("", t.Create)
	adminTags.PATCH("/:id", t.Update)
	adminTags.DELETE("/:id", t.Delete)

	return nil
}
