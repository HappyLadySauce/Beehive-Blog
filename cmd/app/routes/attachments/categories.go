package attachments

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// CreateCategory handles POST /api/v1/attachment/categories (admin).
// CreateCategory 处理 POST /api/v1/attachment/categories（管理员）。
func (h *AttachmentsController) CreateCategory(ctx *gin.Context) {
	var req v1.AttachmentCategoryCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	row, err := h.createCategory(ctx.Request.Context(), actorFromClaims(ctx), pkgattachment.CategoryCreateInput{
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
func (h *AttachmentsController) ListCategories(ctx *gin.Context) {
	rows, err := h.listCategories(ctx.Request.Context(), actorFromClaims(ctx))
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	items := make([]v1.AttachmentCategoryResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCategoryResponse(row))
	}
	common.Success(ctx, v1.AttachmentCategoryListResponse{Items: items})
}

// GetCategory handles GET /api/v1/attachment/categories/:id (admin).
// GetCategory 处理 GET /api/v1/attachment/categories/:id（管理员）。
func (h *AttachmentsController) GetCategory(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	row, err := h.getCategory(ctx.Request.Context(), actorFromClaims(ctx), id)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toCategoryResponse(row))
}

// PatchCategory handles PATCH /api/v1/attachment/categories/:id (admin).
// PatchCategory 处理 PATCH /api/v1/attachment/categories/:id（管理员）。
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
	row, err := h.patchCategory(ctx.Request.Context(), actorFromClaims(ctx), id, pkgattachment.CategoryPatchInput{
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
func (h *AttachmentsController) DeleteCategory(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	if err := h.deleteCategory(ctx.Request.Context(), actorFromClaims(ctx), id); err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}

// createCategory creates an attachment category.
// createCategory 创建附件分类。
func (h *AttachmentsController) createCategory(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.CategoryCreateInput) (model.AttachmentCategory, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Slug) == "" {
		return model.AttachmentCategory{}, fmt.Errorf("%w: name and slug are required", pkgattachment.ErrInvalid)
	}
	status := in.Status
	if status == "" {
		status = pkgattachment.CategoryStatusActive
	}
	if !pkgattachment.CategoryStatusKnown(status) {
		return model.AttachmentCategory{}, fmt.Errorf("%w: invalid category status", pkgattachment.ErrInvalid)
	}
	row := model.AttachmentCategory{
		ParentID:    in.ParentID,
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.TrimSpace(in.Slug),
		Description: pkgattachment.CleanOptional(in.Description),
		Icon:        pkgattachment.CleanOptional(in.Icon),
		SortOrder:   in.SortOrder,
		Status:      status,
	}
	if err := h.svc.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return model.AttachmentCategory{}, pkgattachment.MapDBError(err)
	}
	return row, nil
}

// listCategories returns live categories ordered for tree rendering.
// listCategories 返回按树展示排序的未软删分类。
func (h *AttachmentsController) listCategories(ctx context.Context, actor pkgattachment.Actor) ([]model.AttachmentCategory, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return nil, err
	}
	var rows []model.AttachmentCategory
	if err := h.svc.DB.WithContext(ctx).Order("path ASC").Find(&rows).Error; err != nil {
		return nil, pkgattachment.MapDBError(err)
	}
	return rows, nil
}

// getCategory returns one category.
// getCategory 返回单个分类。
func (h *AttachmentsController) getCategory(ctx context.Context, actor pkgattachment.Actor, id int64) (model.AttachmentCategory, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	var row model.AttachmentCategory
	if err := h.svc.DB.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return model.AttachmentCategory{}, pkgattachment.MapDBError(err)
	}
	return row, nil
}

// patchCategory updates a category.
// patchCategory 更新分类。
func (h *AttachmentsController) patchCategory(ctx context.Context, actor pkgattachment.Actor, id int64, in pkgattachment.CategoryPatchInput) (model.AttachmentCategory, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.AttachmentCategory{}, err
	}
	var out model.AttachmentCategory
	err := h.svc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, "id = ?", id).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		updates := map[string]interface{}{"updated_at": time.Now()}
		if in.ParentID != nil {
			updates["parent_id"] = *in.ParentID
		}
		if in.Name != nil {
			updates["name"] = strings.TrimSpace(*in.Name)
		}
		if in.Slug != nil {
			updates["slug"] = strings.TrimSpace(*in.Slug)
		}
		if in.Description != nil {
			updates["description"] = strings.TrimSpace(*in.Description)
		}
		if in.Icon != nil {
			updates["icon"] = strings.TrimSpace(*in.Icon)
		}
		if in.SortOrder != nil {
			updates["sort_order"] = *in.SortOrder
		}
		if in.Status != nil {
			if !pkgattachment.CategoryStatusKnown(*in.Status) {
				return fmt.Errorf("%w: invalid category status", pkgattachment.ErrInvalid)
			}
			updates["status"] = strings.TrimSpace(*in.Status)
		}
		if err := tx.Model(&out).Updates(updates).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		return tx.First(&out, "id = ?", id).Error
	})
	if err != nil {
		return model.AttachmentCategory{}, err
	}
	return out, nil
}

// deleteCategory soft-deletes a category.
// deleteCategory 软删分类。
func (h *AttachmentsController) deleteCategory(ctx context.Context, actor pkgattachment.Actor, id int64) error {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return err
	}
	res := h.svc.DB.WithContext(ctx).Delete(&model.AttachmentCategory{}, "id = ?", id)
	if res.Error != nil {
		return pkgattachment.MapDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return pkgattachment.ErrNotFound
	}
	return nil
}
