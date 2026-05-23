package comments

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

// CommentsController handles reader-facing comment routes.
// CommentsController 处理读者侧评论接口。
type CommentsController struct {
	svc *svc.ServiceContext
	db  *gorm.DB
}

// NewCommentsController builds a comments controller.
// NewCommentsController 构造评论控制器。
func NewCommentsController(svcCtx *svc.ServiceContext) (*CommentsController, error) {
	if svcCtx == nil {
		return nil, fmt.Errorf("service context is nil")
	}
	if svcCtx.DB == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	return &CommentsController{svc: svcCtx, db: svcCtx.DB}, nil
}

// Init registers public comment routes.
// Init 注册公开评论路由。
func Init(svcCtx *svc.ServiceContext) error {
	c, err := NewCommentsController(svcCtx)
	if err != nil {
		return err
	}

	group := router.V1().Group("/contents/:id/comments")
	group.GET("", c.List)
	group.POST("", c.Create)

	admin := router.V1().Group("/comments")
	admin.Use(middleware.AuthMiddleware(svcCtx), middleware.RequireRole("admin"))
	admin.GET("", c.ListAdmin)
	admin.PATCH("/:commentId/status", c.UpdateStatus)
	admin.DELETE("/:commentId", c.Delete)
	return nil
}

func writeCommentError(ctx *gin.Context, err error) {
	common.Fail(ctx, err)
}
