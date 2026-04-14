package tags

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

// Init 注册公开标签路由（/api/v1/tags）。
func Init(svcCtx *svc.ServiceContext) {
	h := &handlers{svc: NewService(svcCtx)}
	g := router.V1()
	g.GET("/tags", h.handlePublicList)
	g.GET("/tags/cloud", h.handlePublicCloud)
	g.GET("/tags/:slug", h.handlePublicDetail)
}

// RegisterAdminRoutes 在已鉴权的管理员分组上注册标签 CRUD。
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	h := &handlers{svc: NewService(svcCtx)}
	g.GET("/tags", h.handleAdminList)
	g.POST("/tags", h.handleAdminCreate)
	g.PUT("/tags/:id", h.handleAdminUpdate)
	g.DELETE("/tags/:id", h.handleAdminDelete)
}

// handlePublicList godoc
//
//	@Summary		标签列表
//	@Description	公开；分页与 keyword、sort 筛选
//	@Tags			tags
//	@Produce		json
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Param			keyword		query	string	false	"名称或 slug 模糊"
//	@Param			sort		query	string	false	"name|count|newest"
//	@Success		200			{object}	common.BaseResponse{data=v1.TagListResponse}
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/tags [get]
func (h *handlers) handlePublicList(c *gin.Context) {
	var req v1.TagListRequest
	_ = c.ShouldBindQuery(&req)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.PublicList(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handlePublicCloud godoc
//
//	@Summary		标签云
//	@Description	公开；按文章数排序截取前 limit 条
//	@Tags			tags
//	@Produce		json
//	@Param			limit	query	int	false	"条数上限，默认 50"
//	@Success		200		{object}	common.BaseResponse{data=v1.TagCloudResponse}
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/tags/cloud [get]
func (h *handlers) handlePublicCloud(c *gin.Context) {
	var req v1.TagCloudRequest
	_ = c.ShouldBindQuery(&req)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := h.svc.PublicCloud(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// handlePublicDetail godoc
//
//	@Summary		标签详情
//	@Description	公开；按 slug；含文章分页与共现相关标签
//	@Tags			tags
//	@Produce		json
//	@Param			slug		path	string	true	"标签 slug"
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Success		200			{object}	common.BaseResponse{data=v1.TagDetailResponse}
//	@Failure		400			{object}	common.BaseResponse
//	@Failure		404			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/tags/{slug} [get]
func (h *handlers) handlePublicDetail(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	var req v1.TagDetailRequest
	_ = c.ShouldBindQuery(&req)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
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
//	@Summary		管理员标签列表
//	@Description	需管理员
//	@Tags			admin
//	@Produce		json
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Param			keyword		query	string	false	"关键词"
//	@Param			sort		query	string	false	"name|count|newest"
//	@Success		200			{object}	common.BaseResponse{data=v1.TagListResponse}
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/admin/tags [get]
func (h *handlers) handleAdminList(c *gin.Context) {
	var req v1.AdminTagListRequest
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
//	@Summary		创建标签
//	@Description	需管理员；color 可填 #RGB/#RRGGBB、无 # 的 6 位十六进制，或常见颜色名（如 blue、sky blue）；空则默认为 #3B82F6
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body	v1.CreateTagRequest	true	"标签"
//	@Success		200		{object}	common.BaseResponse{data=v1.TagListItem}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/tags [post]
func (h *handlers) handleAdminCreate(c *gin.Context) {
	var req v1.CreateTagRequest
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
//	@Summary		更新标签
//	@Description	需管理员；color 规则同创建（规范化后以 #RRGGBB 存储）
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"标签 ID"
//	@Param			request	body	v1.UpdateTagRequest	true	"更新字段"
//	@Success		200		{object}	common.BaseResponse{data=v1.TagListItem}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/tags/{id} [put]
func (h *handlers) handleAdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid tag id")
		return
	}
	var req v1.UpdateTagRequest
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
//	@Summary		删除标签
//	@Description	需管理员；有关联文章时须 force=true 并先移除 article_tags
//	@Tags			admin
//	@Produce		json
//	@Param			id		path	int		true	"标签 ID"
//	@Param			force	query	bool	false	"强制删除"
//	@Success		200		{object}	common.BaseResponse{data=v1.DeleteTagResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/tags/{id} [delete]
func (h *handlers) handleAdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid tag id")
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
