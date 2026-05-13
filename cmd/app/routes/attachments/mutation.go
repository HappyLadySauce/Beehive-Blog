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

// Patch handles PATCH /api/v1/attachments/:id (admin).
// Patch 处理 PATCH /api/v1/attachments/:id（管理员）。
func (h *AttachmentsController) Patch(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var req v1.AttachmentPatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	out, err := h.patch(ctx.Request.Context(), actorFromClaims(ctx), id, pkgattachment.PatchInput{
		OriginalName: req.OriginalName,
		Status:       req.Status,
		AccessScope:  req.AccessScope,
		CategoryIDs:  req.CategoryIDs,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toAttachmentResponse(out, nil))
}

// Delete handles DELETE /api/v1/attachments/:id (admin).
// Delete 处理 DELETE /api/v1/attachments/:id（管理员）。
func (h *AttachmentsController) Delete(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	if err := h.delete(ctx.Request.Context(), actorFromClaims(ctx), id); err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}

// ReplaceCategories handles PUT /api/v1/attachments/:id/categories (admin).
// ReplaceCategories 处理 PUT /api/v1/attachments/:id/categories（管理员）。
func (h *AttachmentsController) ReplaceCategories(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var req v1.AttachmentCategoryReplaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	if err := h.replaceCategories(ctx.Request.Context(), actorFromClaims(ctx), id, req.CategoryIDs); err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{})
}

// patch updates mutable attachment fields.
// patch 更新附件可变字段。
func (h *AttachmentsController) patch(ctx context.Context, actor pkgattachment.Actor, id int64, in pkgattachment.PatchInput) (model.Attachment, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.Attachment{}, err
	}
	var out model.Attachment
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, "id = ?", id).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		updates := map[string]interface{}{"updated_at": time.Now()}
		if in.OriginalName != nil {
			updates["original_name"] = strings.TrimSpace(*in.OriginalName)
		}
		if in.Status != nil {
			if !pkgattachment.StatusKnown(*in.Status) {
				return fmt.Errorf("%w: invalid status", pkgattachment.ErrInvalid)
			}
			updates["status"] = strings.TrimSpace(*in.Status)
		}
		if in.AccessScope != nil {
			if !pkgattachment.AccessScopeKnown(*in.AccessScope) {
				return fmt.Errorf("%w: invalid access_scope", pkgattachment.ErrInvalid)
			}
			if *in.AccessScope == pkgattachment.AccessPublic && out.UploadStatus != pkgattachment.UploadReady {
				return fmt.Errorf("%w: public attachments must be ready", pkgattachment.ErrInvalid)
			}
			updates["access_scope"] = strings.TrimSpace(*in.AccessScope)
		}
		if err := tx.Model(&out).Updates(updates).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		if in.CategoryIDs != nil {
			if err := replaceCategoriesTx(tx, out.ID, *in.CategoryIDs); err != nil {
				return err
			}
		}
		return tx.First(&out, "id = ?", id).Error
	})
	if err != nil {
		return model.Attachment{}, err
	}
	return out, nil
}

// delete soft-deletes an attachment.
// delete 软删附件。
func (h *AttachmentsController) delete(ctx context.Context, actor pkgattachment.Actor, id int64) error {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return err
	}
	res := h.db.WithContext(ctx).Delete(&model.Attachment{}, "id = ?", id)
	if res.Error != nil {
		return pkgattachment.MapDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return pkgattachment.ErrNotFound
	}
	return nil
}

// replaceCategories replaces all category bindings for an attachment.
// replaceCategories 替换附件的全部分类绑定。
func (h *AttachmentsController) replaceCategories(ctx context.Context, actor pkgattachment.Actor, attachmentID int64, categoryIDs []int64) error {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var attachment model.Attachment
		if err := tx.First(&attachment, "id = ?", attachmentID).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		return replaceCategoriesTx(tx, attachmentID, categoryIDs)
	})
}

func replaceCategoriesTx(tx *gorm.DB, attachmentID int64, categoryIDs []int64) error {
	if err := tx.Where("attachment_id = ?", attachmentID).Delete(&model.AttachmentCategoryBinding{}).Error; err != nil {
		return pkgattachment.MapDBError(err)
	}
	if len(categoryIDs) == 0 {
		return nil
	}
	uniqueCategoryIDs := pkgattachment.UniqueInt64(categoryIDs)
	var count int64
	if err := tx.Model(&model.AttachmentCategory{}).
		Where("id IN ? AND status = ?", uniqueCategoryIDs, pkgattachment.CategoryStatusActive).
		Count(&count).Error; err != nil {
		return pkgattachment.MapDBError(err)
	}
	if count != int64(len(uniqueCategoryIDs)) || len(uniqueCategoryIDs) != len(categoryIDs) {
		return fmt.Errorf("%w: category_ids contain missing, disabled or duplicate categories", pkgattachment.ErrInvalid)
	}
	now := time.Now()
	rows := make([]model.AttachmentCategoryBinding, 0, len(uniqueCategoryIDs))
	for _, categoryID := range uniqueCategoryIDs {
		rows = append(rows, model.AttachmentCategoryBinding{
			AttachmentID: attachmentID,
			CategoryID:   categoryID,
			CreatedAt:    now,
		})
	}
	if err := tx.Create(&rows).Error; err != nil {
		return pkgattachment.MapDBError(err)
	}
	return nil
}
