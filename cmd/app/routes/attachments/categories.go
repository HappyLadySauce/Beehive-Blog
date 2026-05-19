package attachments

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
)

// CreateCategory handles POST /api/v1/attachment/categories (admin).
// CreateCategory 处理 POST /api/v1/attachment/categories（管理员）。
//
//	@Summary		Create attachment category
//	@Description	Creates an attachment category node. Admin only. 中文：创建附件分类节点（仅管理员）。
//	@Tags			attachments
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		v1.AttachmentCategoryCreateRequest	true	"Category creation request"
//	@Success		200		{object}	common.BaseResponse{data=v1.AttachmentCategoryResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/attachment/categories [post]
func (h *AttachmentsController) CreateCategory(ctx *gin.Context) {
	var req v1.AttachmentCategoryCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	row, err := h.categorySvc.Create(ctx.Request.Context(), actorFromClaims(ctx), pkgattachment.CategoryCreateInput{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toCategoryResponse(row))
}

// ListCategories handles GET /api/v1/attachment/categories (admin).
// ListCategories 处理 GET /api/v1/attachment/categories（管理员）。
//
//	@Summary		List attachment categories
//	@Description	Paginated list of attachment categories with optional filters. Admin only. 中文：分页列出附件分类，支持筛选（仅管理员）。
//	@Tags			attachments
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			page_size	query		int		false	"Items per page (default 20, max 100)"
//	@Param			status		query		string	false	"Filter by status"	Enums(active, disabled)
//	@Param			search		query		string	false	"Search name or slug"
//	@Success		200			{object}	common.BaseResponse{data=v1.AttachmentCategoryListResponse}
//	@Failure		400			{object}	common.BaseResponse
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Router			/api/v1/attachment/categories [get]
func (h *AttachmentsController) ListCategories(ctx *gin.Context) {
	var req v1.ListAttachmentCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid query parameters", err))
		return
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	rows, total, err := h.categorySvc.List(ctx.Request.Context(), actorFromClaims(ctx),
		page, pageSize, req.Status, req.Search)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	items := make([]v1.AttachmentCategoryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCategoryResponse(row))
	}
	common.Success(ctx, v1.AttachmentCategoryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetCategory handles GET /api/v1/attachment/categories/:id (admin).
// GetCategory 处理 GET /api/v1/attachment/categories/:id（管理员）。
//
//	@Summary		Get attachment category
//	@Description	Returns one attachment category by ID. Admin only. 中文：按 ID 返回附件分类（仅管理员）。
//	@Tags			attachments
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Category ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.AttachmentCategoryResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/attachment/categories/{id} [get]
func (h *AttachmentsController) GetCategory(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	row, err := h.categorySvc.Get(ctx.Request.Context(), actorFromClaims(ctx), id)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toCategoryResponse(row))
}

// PatchCategory handles PATCH /api/v1/attachment/categories/:id (admin).
// PatchCategory 处理 PATCH /api/v1/attachment/categories/:id（管理员）。
//
//	@Summary		Patch attachment category
//	@Description	Updates an attachment category. Admin only. 中文：更新附件分类（仅管理员）。
//	@Tags			attachments
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int									true	"Category ID"
//	@Param			body	body		v1.AttachmentCategoryPatchRequest	true	"Category patch request"
//	@Success		200		{object}	common.BaseResponse{data=v1.AttachmentCategoryResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Router			/api/v1/attachment/categories/{id} [patch]
func (h *AttachmentsController) PatchCategory(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var req v1.AttachmentCategoryPatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	row, err := h.categorySvc.Patch(ctx.Request.Context(), actorFromClaims(ctx), id, pkgattachment.CategoryPatchInput{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toCategoryResponse(row))
}

// DeleteCategory handles DELETE /api/v1/attachment/categories/:id (admin).
// DeleteCategory 处理 DELETE /api/v1/attachment/categories/:id（管理员）。
//
//	@Summary		Delete attachment category
//	@Description	Soft-deletes an attachment category. Admin only. 中文：软删除附件分类（仅管理员）。
//	@Tags			attachments
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Category ID"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		409	{object}	common.BaseResponse
//	@Router			/api/v1/attachment/categories/{id} [delete]
func (h *AttachmentsController) DeleteCategory(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	if err := h.categorySvc.Delete(ctx.Request.Context(), actorFromClaims(ctx), id); err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}
