package archives

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
)

type articleHandlers struct {
	svc *ArticleAdmin
}

// RegisterArticleAdminRoutes 在已挂载管理员鉴权的 RouterGroup 上注册文章管理路由（路径前缀为 /api/v1/admin）。
func RegisterArticleAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := &articleHandlers{svc: newArticleAdmin(svcCtx)}
	g.POST("/articles", h.handleCreateArticle)
	g.PUT("/articles/:id", h.handleUpdateArticle)
	g.DELETE("/articles/:id", h.handleDeleteArticle)
	g.PUT("/articles/:id/status", h.handleUpdateArticleStatus)
	g.PUT("/articles/:id/slug", h.handleUpdateArticleSlug)
	g.PUT("/articles/:id/password", h.handleUpdateArticlePassword)
	g.PUT("/articles/:id/pin", h.handleUpdateArticlePin)
}

// handleCreateArticle godoc
//
//	@Summary		管理员创建文章
//	@Description	需管理员；创建草稿或已发布文章，可指定 slug 或由标题生成唯一 slug
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.CreateArticleRequest	true	"文章正文"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles [post]
func (h *articleHandlers) handleCreateArticle(c *gin.Context) {
	uid, ok := middlewares.GetCurrentUserID(c)
	if !ok || uid <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req v1.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.CreateArticle(ctx, uid, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdateArticle godoc
//
//	@Summary		管理员更新文章
//	@Description	需管理员；部分字段更新，含分类、标签、slug 等
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"文章 ID"
//	@Param			request	body		v1.UpdateArticleRequest	true	"更新字段"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id} [put]
func (h *articleHandlers) handleUpdateArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdateArticle(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleDeleteArticle godoc
//
//	@Summary		管理员删除文章
//	@Description	需管理员；软删除文章并清理关联标签；开启 Hexo auto_sync 时异步删 md
//	@Tags			admin
//	@Produce		json
//	@Param			id	path	int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.DeleteArticleResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id} [delete]
func (h *articleHandlers) handleDeleteArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.DeleteArticle(ctx, id, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdateArticleStatus godoc
//
//	@Summary		更新文章发布状态
//	@Description	需管理员；如发布且未传 publishedAt 则使用当前时间
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"文章 ID"
//	@Param			request	body	v1.UpdateArticleStatusRequest	true	"状态"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/status [put]
func (h *articleHandlers) handleUpdateArticleStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdateArticleStatus(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdateArticleSlug godoc
//
//	@Summary		更新文章 slug
//	@Description	需管理员；slug 须全局唯一
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int						true	"文章 ID"
//	@Param			request	body	v1.UpdateArticleSlugRequest	true	"新 slug"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/slug [put]
func (h *articleHandlers) handleUpdateArticleSlug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticleSlugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdateArticleSlug(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdateArticlePassword godoc
//
//	@Summary		设置或清除文章访问密码
//	@Description	需管理员；非空密码长度 4–20；空字符串表示清除密码
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int								true	"文章 ID"
//	@Param			request	body	v1.UpdateArticlePasswordRequest	true	"密码（可空）"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleSecurityResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/password [put]
func (h *articleHandlers) handleUpdateArticlePassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticlePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdateArticlePassword(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdateArticlePin godoc
//
//	@Summary		更新文章置顶
//	@Description	需管理员；取消置顶时 pin_order 归零
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"文章 ID"
//	@Param			request	body	v1.UpdateArticlePinRequest	true	"置顶与排序"
//	@Success		200		{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/pin [put]
func (h *articleHandlers) handleUpdateArticlePin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	var req v1.UpdateArticlePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdateArticlePin(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
