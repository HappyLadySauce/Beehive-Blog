package pages

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
)

type pageHandlers struct {
	svc *PageAdmin
}

// RegisterAdminRoutes 在已挂载管理员鉴权的 RouterGroup 上注册独立页面路由。
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := &pageHandlers{svc: newPageAdmin(svcCtx)}
	g.GET("/pages/trash", h.handleListTrashedPages)
	g.GET("/pages", h.handleListPages)
	g.POST("/pages", h.handleCreatePage)
	g.GET("/pages/:id", h.handleGetPage)
	g.PUT("/pages/:id", h.handleUpdatePage)
	g.PUT("/pages/:id/status", h.handleUpdatePageStatus)
	g.DELETE("/pages/:id", h.handleDeletePage)
	g.POST("/pages/:id/restore", h.handleRestorePage)
	g.DELETE("/pages/:id/permanent", h.handlePermanentDeletePage)
}

// handleListPages godoc
//
//	@Summary		管理员独立页面列表
//	@Description	分页、关键词、状态筛选、排序 newest|oldest|popular
//	@Tags			admin
//	@Produce		json
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Param			keyword		query	string	false	"关键词"
//	@Param			status		query	string	false	"状态，逗号分隔"
//	@Param			sort		query	string	false	"newest|oldest|popular"
//	@Success		200	{object}	common.BaseResponse{data=v1.AdminPageListResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/pages [get]
func (h *pageHandlers) handleListPages(c *gin.Context) {
	var req v1.AdminPageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.AdminListPages(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleListTrashedPages godoc
//
//	@Summary		独立页面回收站
//	@Tags			admin
//	@Produce		json
//	@Param			page		query	int	false	"页码"
//	@Param			pageSize	query	int	false	"每页条数"
//	@Param			keyword		query	string	false	"关键词"
//	@Param			sort		query	string	false	"排序"
//	@Success		200	{object}	common.BaseResponse{data=v1.AdminPageListResponse}
//	@Router			/api/v1/admin/pages/trash [get]
func (h *pageHandlers) handleListTrashedPages(c *gin.Context) {
	var req v1.AdminPageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.ListTrashedPages(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleCreatePage godoc
//
//	@Summary		创建独立页面
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.CreatePageRequest	true	"正文"
//	@Success		200		{object}	common.BaseResponse{data=v1.PageDetailResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/admin/pages [post]
func (h *pageHandlers) handleCreatePage(c *gin.Context) {
	var req v1.CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.CreatePage(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleGetPage godoc
//
//	@Summary		获取独立页面详情
//	@Tags			admin
//	@Produce		json
//	@Param			id	path	int	true	"页面 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.PageDetailResponse}
//	@Router			/api/v1/admin/pages/{id} [get]
func (h *pageHandlers) handleGetPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.GetPage(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdatePage godoc
//
//	@Summary		更新独立页面
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int						true	"页面 ID"
//	@Param			request	body	v1.UpdatePageRequest	true	"字段"
//	@Success		200		{object}	common.BaseResponse{data=v1.PageDetailResponse}
//	@Router			/api/v1/admin/pages/{id} [put]
func (h *pageHandlers) handleUpdatePage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	var req v1.UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdatePage(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleUpdatePageStatus godoc
//
//	@Summary		更新独立页面状态
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"页面 ID"
//	@Param			request	body	v1.UpdatePageStatusRequest	true	"状态"
//	@Success		200		{object}	common.BaseResponse{data=v1.PageDetailResponse}
//	@Router			/api/v1/admin/pages/{id}/status [put]
func (h *pageHandlers) handleUpdatePageStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	var req v1.UpdatePageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.UpdatePageStatus(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleDeletePage godoc
//
//	@Summary		软删除独立页面
//	@Tags			admin
//	@Produce		json
//	@Param			id	path	int	true	"页面 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.DeletePageResponse}
//	@Router			/api/v1/admin/pages/{id} [delete]
func (h *pageHandlers) handleDeletePage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.DeletePage(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handleRestorePage godoc
//
//	@Summary		从回收站恢复独立页面
//	@Tags			admin
//	@Produce		json
//	@Param			id	path	int	true	"页面 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.DeletePageResponse}
//	@Router			/api/v1/admin/pages/{id}/restore [post]
func (h *pageHandlers) handleRestorePage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.RestorePage(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handlePermanentDeletePage godoc
//
//	@Summary		永久删除独立页面
//	@Tags			admin
//	@Produce		json
//	@Param			id	path	int	true	"页面 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.DeletePageResponse}
//	@Router			/api/v1/admin/pages/{id}/permanent [delete]
func (h *pageHandlers) handlePermanentDeletePage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid page id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.PermanentDeletePage(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
