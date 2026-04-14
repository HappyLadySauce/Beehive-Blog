package comments

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/common"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// Init registers public and authenticated comment routes on /api/v1.
func Init(svcCtx *svc.ServiceContext) {
	s := NewService(svcCtx)
	g := router.V1()
	g.GET("/articles/:id/comments", s.HandleListComments)
	g.POST("/articles/:id/comments", middlewares.Auth(svcCtx), s.HandleCreateComment)
}

// RegisterAdminRoutes registers GET/PUT admin comment routes on the admin group.
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	s := NewService(svcCtx)
	g.GET("/comments", s.HandleAdminListComments)
	g.PUT("/comments/:id/status", s.HandleUpdateCommentStatus)
}

// HandleListComments godoc
//
//	@Summary		文章评论列表
//	@Description	公开；仅返回已通过审核的评论
//	@Tags			comments
//	@Produce		json
//	@Param			id			path		int	true	"文章 ID"
//	@Param			page		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页条数"
//	@Success		200			{object}	common.BaseResponse{data=v1.CommentListResponse}
//	@Failure		404			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/articles/{id}/comments [get]
func (s *Service) HandleListComments(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var q v1.CommentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		klog.ErrorS(err, "HandleListComments bind query")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	page := q.Page
	pageSize := q.PageSize
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.ListByArticle(ctx, id, page, pageSize)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleCreateComment godoc
//
//	@Summary		发表评论
//	@Description	需登录；新评论待审核
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"文章 ID"
//	@Param			request	body		v1.CreateCommentRequest	true	"评论"
//	@Success		200		{object}	common.BaseResponse{data=v1.CreateCommentResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/articles/{id}/comments [post]
func (s *Service) HandleCreateComment(c *gin.Context) {
	uid, ok := middlewares.GetCurrentUserID(c)
	if !ok || uid <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.Create(ctx, id, uid, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminListComments godoc
//
//	@Summary		管理员评论列表
//	@Description	需管理员；支持按文章、状态、关键词筛选
//	@Tags			admin
//	@Produce		json
//	@Param			page		query		int		false	"页码"
//	@Param			pageSize	query		int		false	"每页条数"
//	@Param			articleId	query		int		false	"文章 ID"
//	@Param			status		query		string	false	"pending|approved|rejected|spam"
//	@Param			keyword		query		string	false	"内容关键词"
//	@Success		200			{object}	common.BaseResponse{data=v1.AdminCommentListResponse}
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/admin/comments [get]
func (s *Service) HandleAdminListComments(c *gin.Context) {
	var q v1.AdminCommentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminList(ctx, &q)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleUpdateCommentStatus godoc
//
//	@Summary		更新评论审核状态
//	@Description	需管理员
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"评论 ID"
//	@Param			request	body		v1.UpdateCommentStatusRequest	true	"状态"
//	@Success		200		{object}	common.BaseResponse{data=v1.UpdateCommentStatusResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/comments/{id}/status [put]
func (s *Service) HandleUpdateCommentStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid comment id")
		return
	}
	var req v1.UpdateCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateAdminStatus(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
