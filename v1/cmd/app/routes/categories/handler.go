package categories

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/common"
	"github.com/gin-gonic/gin"
)

type handlers struct {
	svc *Service
}

// Init 注册公开分类路由（/api/v1/categories）。
func Init(svcCtx *svc.ServiceContext) {
	h := &handlers{svc: NewService(svcCtx)}
	g := router.V1()
	g.GET("/categories", h.handlePublicList)
	g.GET("/categories/:slug", h.handlePublicDetail)
}

// RegisterAdminRoutes 在已鉴权的管理员分组上注册分类 CRUD。
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := &handlers{svc: NewService(svcCtx)}
	g.GET("/categories", h.handleAdminList)
	g.POST("/categories", h.handleAdminCreate)
	g.PUT("/categories/:id", h.handleAdminUpdate)
	g.DELETE("/categories/:id", h.handleAdminDelete)
}

// handlePublicList godoc
//
//	@Summary		分类列表
//	@Description	公开；返回一级分类列表（按 sortOrder、id 排序）。公开详情见 GET /api/v1/categories/{slug}；管理员增删改使用数字 id。
//	@Tags			categories
//	@Produce		json
//	@Success		200	{object}	common.BaseResponse{data=v1.CategoryListResponse}
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/categories [get]
func (h *handlers) handlePublicList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.PublicList(ctx)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handlePublicDetail godoc
//
//	@Summary		分类详情
//	@Description	公开；按 slug（非数字 id）查询分类及该分类下已发布文章分页；不含子分类（系统仅支持一级分类）
//	@Tags			categories
//	@Produce		json
//	@Param			slug		path		string	true	"分类 slug"
//	@Param			page		query		int		false	"页码"
//	@Param			pageSize	query		int		false	"每页条数"
//	@Success		200			{object}	common.BaseResponse{data=v1.CategoryDetailResponse}
//	@Failure		400			{object}	common.BaseResponse
//	@Failure		404			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/categories/{slug} [get]
func (h *handlers) handlePublicDetail(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	var req v1.CategoryDetailRequest
	_ = c.ShouldBindQuery(&req)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.PublicDetail(ctx, slug, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleAdminList godoc
//
//	@Summary		管理员分类列表
//	@Description	需管理员；扁平分页列表
//	@Tags			admin
//	@Produce		json
//	@Param			page		query	int	false	"页码"
//	@Param			pageSize	query	int	false	"每页条数"
//	@Success		200			{object}	common.BaseResponse{data=v1.AdminCategoryListResponse}
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/admin/categories [get]
func (h *handlers) handleAdminList(c *gin.Context) {
	var req v1.AdminCategoryListRequest
	_ = c.ShouldBindQuery(&req)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminList(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleAdminCreate godoc
//
//	@Summary		创建分类
//	@Description	需管理员
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.CreateCategoryRequest	true	"分类"
//	@Success		200		{object}	common.BaseResponse{data=v1.CategoryBrief}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/categories [post]
func (h *handlers) handleAdminCreate(c *gin.Context) {
	var req v1.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminCreate(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleAdminUpdate godoc
//
//	@Summary		更新分类
//	@Description	需管理员
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"分类 ID"
//	@Param			request	body		v1.UpdateCategoryRequest	true	"更新字段"
//	@Success		200		{object}	common.BaseResponse{data=v1.CategoryBrief}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/categories/{id} [put]
func (h *handlers) handleAdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid category id")
		return
	}
	var req v1.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminUpdate(ctx, id, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleAdminDelete godoc
//
//	@Summary		删除分类
//	@Description	需管理员；有关联文章时需传 force=true（先将文章 category_id 置空后删除分类）
//	@Tags			admin
//	@Produce		json
//	@Param			id		path	int		true	"分类 ID"
//	@Param			force	query	bool	false	"强制删除（解除文章分类关联后删除）"
//	@Success		200		{object}	common.BaseResponse{data=v1.DeleteCategoryResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/categories/{id} [delete]
func (h *handlers) handleAdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid category id")
		return
	}
	force := strings.EqualFold(strings.TrimSpace(c.Query("force")), "true")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminDelete(ctx, id, force, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
