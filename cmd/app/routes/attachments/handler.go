package attachments

import (
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	attachmentstorage "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/storage"
)

// AttachmentsController serves attachment HTTP endpoints.
// AttachmentsController 提供附件 HTTP 接口。
type AttachmentsController struct {
	svc *svc.ServiceContext
}

// NewAttachmentsController builds an AttachmentsController bound to the given service context.
// NewAttachmentsController 基于给定 ServiceContext 构造 AttachmentsController。
func NewAttachmentsController(svcCtx *svc.ServiceContext) *AttachmentsController {
	return &AttachmentsController{svc: svcCtx}
}

// Init initializes attachment services and registers attachment routes.
// Init 初始化附件服务并注册附件路由。
func Init(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return fmt.Errorf("service context is nil")
	}
	if svcCtx.Config == nil {
		return fmt.Errorf("config is nil")
	}
	if svcCtx.Config.Attachment == nil {
		return fmt.Errorf("attachment config is nil")
	}
	if _, err := attachmentstorage.NewRegistry(svcCtx.Config.Attachment); err != nil {
		return fmt.Errorf("init attachment storage: %w", err)
	}
	
	h := NewAttachmentsController(svcCtx)

	attachments := router.V1().Group("/attachments")
	attachments.GET("/:id", h.GetAttachment)
	attachments.GET("/:id/content", h.GetAttachmentContent)

	adminAttachments := attachments.Group("")
	adminAttachments.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	adminAttachments.POST("", h.UploadLocal)
	adminAttachments.POST("/upload-url", h.PresignRemote)
	adminAttachments.POST("/:id/complete", h.CompleteRemote)
	adminAttachments.GET("", h.List)
	adminAttachments.PATCH("/:id", h.Patch)
	adminAttachments.DELETE("/:id", h.Delete)
	adminAttachments.PUT("/:id/categories", h.ReplaceCategories)

	categories := router.V1().Group("/attachment/categories")
	categories.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	categories.POST("", h.CreateCategory)
	categories.GET("", h.ListCategories)
	categories.GET("/:id", h.GetCategory)
	categories.PATCH("/:id", h.PatchCategory)
	categories.DELETE("/:id", h.DeleteCategory)

	return nil
}

func (h *AttachmentsController) storageRegistry() (*attachmentstorage.Registry, error) {
	return attachmentstorage.NewRegistry(h.svc.Config.Attachment)
}
