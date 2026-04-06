package attachments

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

// HandleUpload godoc
//
//	@Summary		上传附件
//	@Description	管理员上传单个附件（multipart/form-data，字段名 file），支持可选 groupId 参数
//	@Tags			admin
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file	true	"附件文件"
//	@Param			groupId		query		integer	false	"附件分组 ID"
//	@Success		200			{object}	common.BaseResponse
//	@Failure		400			{object}	common.BaseResponse
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/attachments/upload [post]
func (s *Service) HandleUpload(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok {
		common.Fail(c, http.StatusUnauthorized, nil)
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		klog.ErrorS(err, "HandleUpload: get form file")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	var groupID *int64
	if gidStr := c.Query("groupId"); gidStr != "" {
		gid, err := strconv.ParseInt(gidStr, 10, 64)
		if err == nil && gid > 0 {
			groupID = &gid
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, status, err := s.Upload(ctx, userID, fh, groupID, false)
	if err != nil {
		common.Fail(c, status, err)
		return
	}
	common.Success(c, resp)
}

// HandleList godoc
//
//	@Summary		附件列表
//	@Description	分页查询管理员上传的附件，支持 type/keyword/groupId 筛选
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		integer	false	"页码 (默认1)"
//	@Param			pageSize	query		integer	false	"每页数量 (默认20)"
//	@Param			type		query		string	false	"文件类型 (image/document/video/audio/other)"
//	@Param			keyword		query		string	false	"关键词搜索"
//	@Param			groupId		query		integer	false	"分组 ID"
//	@Success		200			{object}	common.BaseResponse
//	@Failure		400			{object}	common.BaseResponse
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/attachments [get]
func (s *Service) HandleList(c *gin.Context) {
	var q v1.AttachmentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, status, err := s.List(ctx, &q)
	if err != nil {
		common.Fail(c, status, err)
		return
	}
	common.Success(c, resp)
}

// HandleDelete godoc
//
//	@Summary		删除附件
//	@Description	删除指定 ID 的附件记录及物理文件
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path		integer	true	"附件 ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/attachments/{id} [delete]
func (s *Service) HandleDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		common.Fail(c, http.StatusBadRequest, nil)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, status, err := s.Delete(ctx, id)
	if err != nil {
		common.Fail(c, status, err)
		return
	}
	common.Success(c, resp)
}

// HandleUploadImage godoc
//
//	@Summary		编辑器图片上传
//	@Description	管理员在编辑文章时拖入图片后调用；只接受 image/* 类型；返回 url 和 alt 供编辑器插入 Markdown 引用
//	@Tags			admin
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"图片文件 (image/*)"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/upload-image [post]
func (s *Service) HandleUploadImage(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok {
		common.Fail(c, http.StatusUnauthorized, nil)
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		klog.ErrorS(err, "HandleUploadImage: get form file")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	item, status, err := s.Upload(ctx, userID, fh, nil, true)
	if err != nil {
		common.Fail(c, status, err)
		return
	}
	// derive alt text from original filename without extension
	alt := item.OriginalName
	if idx := len(alt) - len(".") - len(alt[len(alt)-4:]); idx > 0 {
		// strip extension (up to 5 chars) for common image formats
		for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp"} {
			if len(alt) > len(ext) {
				suffix := alt[len(alt)-len(ext):]
				if suffix == ext {
					alt = alt[:len(alt)-len(ext)]
					break
				}
			}
		}
	}

	common.Success(c, v1.UploadImageResponse{
		URL: item.URL,
		Alt: alt,
	})
}

// RegisterAdminRoutes registers attachment-related routes inside the admin group.
// The group is expected to already have Auth + admin-role middleware applied.
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	svc := NewService(svcCtx)
	g.POST("/attachments/upload", svc.HandleUpload)
	g.GET("/attachments", svc.HandleList)
	g.DELETE("/attachments/:id", svc.HandleDelete)
	g.POST("/upload-image", svc.HandleUploadImage)
}
