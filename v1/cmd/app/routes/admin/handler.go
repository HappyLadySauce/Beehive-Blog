package admin

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/archives"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/attachments"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/categories"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/comments"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/pages"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/site"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/tags"
	userroute "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/user"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"

	"github.com/gin-gonic/gin"
)

// pingResponse is a minimal health-style payload for admin route smoke tests.
type pingResponse struct {
	Message string `json:"message"`
}

// HandlePing godoc
//
//	@Summary		管理员探活
//	@Description	需管理员角色；用于验证 /api/v1/admin 分组与鉴权是否生效
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Router			/api/v1/admin/ping [get]
func HandlePing(c *gin.Context) {
	common.Success(c, pingResponse{Message: "pong"})
}

// Init registers routes under /api/v1/admin with Auth + admin role.
func Init(svcCtx *svc.ServiceContext) {
	g := router.V1().Group("/admin")
	g.Use(middlewares.Auth(svcCtx), middlewares.RequireRoles(models.UserRoleAdmin))
	g.GET("/ping", HandlePing)
	g.POST("/sync/posts", HandleSyncPosts(svcCtx))
	g.GET("/sync/status", HandleSyncStatus(svcCtx))
	archives.RegisterArticleAdminRoutes(g, svcCtx)
	pages.RegisterAdminRoutes(g, svcCtx)
	categories.RegisterAdminRoutes(g, svcCtx)
	tags.RegisterAdminRoutes(g, svcCtx)
	comments.RegisterAdminRoutes(g, svcCtx)
	attachments.RegisterAdminRoutes(g, svcCtx)
	site.RegisterAdminRoutes(g, svcCtx)
	userroute.RegisterAdminRoutes(g, svcCtx)
}
