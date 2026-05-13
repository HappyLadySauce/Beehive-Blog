package attachments

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	attachmentstorage "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/storage"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

// AttachmentsController serves attachment HTTP endpoints.
// AttachmentsController 提供附件 HTTP 接口。
type AttachmentsController struct {
	svc               *svc.ServiceContext
	db                *gorm.DB
	attachmentOptions *options.AttachmentOptions
	storage           *attachmentstorage.Registry
}

// NewAttachmentsController builds an AttachmentsController bound to the given service context.
// NewAttachmentsController 基于给定 ServiceContext 构造 AttachmentsController。
func NewAttachmentsController(svcCtx *svc.ServiceContext) (*AttachmentsController, error) {
	if err := validateDependencies(svcCtx); err != nil {
		return nil, err
	}
	storageRegistry, err := attachmentstorage.NewRegistry(svcCtx.Config.Attachment)
	if err != nil {
		return nil, fmt.Errorf("init attachment storage: %w", err)
	}
	return &AttachmentsController{
		svc:               svcCtx,
		db:                svcCtx.DB,
		attachmentOptions: svcCtx.Config.Attachment,
		storage:           storageRegistry,
	}, nil
}

// Init initializes attachment services and registers attachment routes.
// Init 初始化附件服务并注册附件路由。
func Init(svcCtx *svc.ServiceContext) error {
	h, err := NewAttachmentsController(svcCtx)
	if err != nil {
		return err
	}

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

func validateDependencies(svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return fmt.Errorf("service context is nil")
	}
	if svcCtx.Config == nil {
		return fmt.Errorf("config is nil")
	}
	if svcCtx.Config.Attachment == nil {
		return fmt.Errorf("attachment config is nil")
	}
	if svcCtx.DB == nil {
		return fmt.Errorf("database handle is nil")
	}
	return nil
}
