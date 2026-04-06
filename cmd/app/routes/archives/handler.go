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
	// 批量操作须在 /:id 参数路由之前注册，避免路径冲突
	g.POST("/articles/batch", h.handleBatchArticles)
	g.GET("/articles", h.handleListArticles)
	g.POST("/articles", h.handleCreateArticle)
	g.PUT("/articles/:id", h.handleUpdateArticle)
	g.DELETE("/articles/:id", h.handleDeleteArticle)
	g.PUT("/articles/:id/status", h.handleUpdateArticleStatus)
	g.PUT("/articles/:id/slug", h.handleUpdateArticleSlug)
	g.PUT("/articles/:id/password", h.handleUpdateArticlePassword)
	g.PUT("/articles/:id/pin", h.handleUpdateArticlePin)
	g.GET("/articles/:id/versions", h.handleListVersions)
	g.POST("/articles/:id/versions/:versionId/restore", h.handleRestoreVersion)
	g.GET("/articles/:id/export", h.handleExportArticle)
}

// handleListArticles godoc
//
//	@Summary		管理员文章列表
//	@Description	需管理员；分页列出未软删文章，支持 keyword/category slug/tag/author 筛选；status 为逗号分隔多状态，省略则不限状态（含草稿）
//	@Tags			admin
//	@Produce		json
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Param			keyword		query	string	false	"关键词"
//	@Param			category	query	string	false	"分类 slug"
//	@Param			tag			query	string	false	"标签 slug，逗号分隔多标签交集"
//	@Param			author		query	string	false	"作者用户名"
//	@Param			status		query	string	false	"状态，逗号分隔 draft,published,archived,private,scheduled"
//	@Param			sort		query	string	false	"排序 newest|oldest|popular"
//	@Success		200	{object}	common.BaseResponse{data=v1.AdminArticleListResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles [get]
func (h *articleHandlers) handleListArticles(c *gin.Context) {
	var req v1.AdminArticleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminListArticles(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleCreateArticle godoc
//
//	@Summary		管理员创建文章
//	@Description	需管理员；创建草稿或已发布文章，可指定 slug 或由标题生成唯一 slug。categoryId：省略或 ≤0 时自动归入 slug=default 的默认分类；传正整数必须为已存在分类主键。tagIds：已存在标签主键列表，可省略；≤0 与重复 id 会被忽略；列表内 id 必须全部存在否则 400。
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
//	@Description	需管理员；部分字段更新。categoryId：仅当请求体包含该字段时更新；传 null/省略不改变；传 ≤0 表示置空分类；传正整数须为已存在分类。tagIds：仅当请求体包含非空 tagIds 数组时整表替换关联；须全部为已存在标签主键（会去重并忽略 ≤0）。
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

// handleBatchArticles godoc
//
//	@Summary		文章批量操作
//	@Description	需管理员；支持 delete/set_status/set_category/set_tags
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.BatchArticleRequest	true	"批量操作参数"
//	@Success		200		{object}	common.BaseResponse{data=v1.BatchArticleResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/batch [post]
func (h *articleHandlers) handleBatchArticles(c *gin.Context) {
	var req v1.BatchArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	resp, code, err := h.svc.BatchArticles(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleExportArticle godoc
//
//	@Summary		导出文章
//	@Description	需管理员；format=markdown 返回原始 Markdown，format=html 返回 HTML 包装，format=pdf 返回 501
//	@Tags			admin
//	@Produce		application/octet-stream
//	@Param			id		path	int		true	"文章 ID"
//	@Param			format	query	string	false	"导出格式 markdown|html|pdf，默认 markdown"
//	@Success		200		{file}	binary
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		501		{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/export [get]
func (h *articleHandlers) handleExportArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	format := c.DefaultQuery("format", "markdown")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	data, contentType, code, err := h.svc.ExportArticle(ctx, id, format)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	ext := ".md"
	if format == "html" {
		ext = ".html"
	}
	c.Header("Content-Disposition", "attachment; filename=\"article-"+strconv.FormatInt(id, 10)+ext+"\"")
	c.Data(http.StatusOK, contentType, data)
}

// handleListVersions godoc
//
//	@Summary		文章版本历史列表
//	@Description	需管理员；列出指定文章的历史版本（最多 50 条，按版本号倒序）
//	@Tags			admin
//	@Produce		json
//	@Param			id	path		int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.ArticleVersionListResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/versions [get]
func (h *articleHandlers) handleListVersions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.ListVersions(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleRestoreVersion godoc
//
//	@Summary		恢复文章版本
//	@Description	需管理员；将文章内容恢复到指定历史版本，当前内容自动保存为新版本
//	@Tags			admin
//	@Produce		json
//	@Param			id			path	int	true	"文章 ID"
//	@Param			versionId	path	int	true	"版本 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/articles/{id}/versions/{versionId}/restore [post]
func (h *articleHandlers) handleRestoreVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	versionID, err := strconv.ParseInt(c.Param("versionId"), 10, 64)
	if err != nil || versionID <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid version id")
		return
	}
	operatorID, ok := middlewares.GetCurrentUserID(c)
	if !ok {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.RestoreVersion(ctx, id, versionID, operatorID)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
