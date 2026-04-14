package likes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

// HandleLike godoc
//
//	@Summary		文章点赞
//	@Description	登录用户对文章点赞；同一文章只能点赞一次，重复点赞返回 409
//	@Tags			user
//	@Produce		json
//	@Param			id	path		int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		409	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/user/articles/{id}/like [post]
func (s *Service) HandleLike(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || articleID <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.LikeArticle(ctx, articleID, userID)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleUnlike godoc
//
//	@Summary		取消文章点赞
//	@Description	登录用户取消对文章的点赞；未点赞时返回 404
//	@Tags			user
//	@Produce		json
//	@Param			id	path		int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/user/articles/{id}/like [delete]
func (s *Service) HandleUnlike(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || articleID <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UnlikeArticle(ctx, articleID, userID)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// Init 在 /api/v1/user 分组下注册点赞路由（Auth 中间件由 user.Init 已挂载）。
func Init(svcCtx *svc.ServiceContext) {
	svc := NewService(svcCtx)
	// 挂在 /user 分组下，复用已有的 Auth 中间件
	ug := router.V1().Group("/user")
	ug.Use(middlewares.Auth(svcCtx))
	ug.POST("/articles/:id/like", svc.HandleLike)
	ug.DELETE("/articles/:id/like", svc.HandleUnlike)
	klog.InfoS("Likes routes registered")
}
