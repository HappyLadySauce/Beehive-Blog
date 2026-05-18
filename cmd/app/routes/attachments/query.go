package attachments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// List handles GET /api/v1/attachments (admin).
// List 处理 GET /api/v1/attachments（管理员）。
func (h *AttachmentsController) List(ctx *gin.Context) {
	ownerUserID, err := optionalInt64Query(ctx, "owner_user_id")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid owner_user_id", err))
		return
	}
	categoryID, err := optionalInt64Query(ctx, "category_id")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid category_id", err))
		return
	}
	cursorID, err := optionalCursor(ctx.Query("cursor"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid cursor", err))
		return
	}
	limit, err := optionalLimit(ctx.Query("limit"))
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid limit", err))
		return
	}
	rows, next, err := h.list(ctx.Request.Context(), actorFromClaims(ctx), pkgattachment.ListInput{
		OwnerUserID:     ownerUserID,
		Purpose:         strings.TrimSpace(ctx.Query("purpose")),
		Status:          strings.TrimSpace(ctx.Query("status")),
		CategoryID:      categoryID,
		CategoryMode:    strings.TrimSpace(ctx.Query("category_mode")),
		Search:          strings.TrimSpace(ctx.Query("search")),
		ReferenceStatus: strings.TrimSpace(ctx.Query("reference_status")),
		CursorID:        cursorID,
		Limit:           limit,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	items := make([]v1.AttachmentResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAttachmentResponse(row, nil))
	}
	common.Success(ctx, v1.AttachmentListResponse{Items: items, NextCursor: next})
}

// GetAttachment handles GET /api/v1/attachments/:id.
// GetAttachment 处理 GET /api/v1/attachments/:id。
func (h *AttachmentsController) GetAttachment(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	actor, hasAuth, err := h.optionalActor(ctx)
	if err != nil {
		common.Fail(ctx, common.NewUnauthorized("invalid or expired access token", err))
		return
	}
	if !hasAuth || actor.Role != pkgattachment.RoleAdmin {
		content, err := h.getPublicContent(ctx.Request.Context(), id, false)
		if err != nil {
			writeAttachmentError(ctx, err)
			return
		}
		common.Success(ctx, toPublicAttachmentResponse(content.Attachment))
		return
	}
	row, categoryIDs, err := h.getAdmin(ctx.Request.Context(), actor, id)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toAttachmentResponse(row, categoryIDs))
}

// GetAttachmentContent handles GET /api/v1/attachments/:id/content.
// GetAttachmentContent 处理 GET /api/v1/attachments/:id/content。
func (h *AttachmentsController) GetAttachmentContent(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	actor, hasAuth, err := h.optionalActor(ctx)
	if err != nil {
		common.Fail(ctx, common.NewUnauthorized("invalid or expired access token", err))
		return
	}
	admin := hasAuth && actor.Role == pkgattachment.RoleAdmin
	out, err := h.getPublicContent(ctx.Request.Context(), id, admin)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	if out.RedirectURL != "" {
		ctx.Redirect(http.StatusFound, out.RedirectURL)
		return
	}
	ctx.FileAttachment(out.LocalPath, out.Attachment.Filename)
}

func (h *AttachmentsController) list(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.ListInput) ([]model.Attachment, string, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return nil, "", err
	}
	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := h.db.WithContext(ctx).Model(&model.Attachment{}).Order("id DESC").Limit(limit + 1)
	if in.OwnerUserID != nil {
		q = q.Where("owner_user_id = ?", *in.OwnerUserID)
	}
	if in.Purpose != "" {
		if !pkgattachment.PurposeKnown(in.Purpose) {
			return nil, "", fmt.Errorf("%w: invalid purpose", pkgattachment.ErrInvalid)
		}
		q = q.Where("purpose = ?", in.Purpose)
	}
	if in.Status != "" {
		if !pkgattachment.StatusKnown(in.Status) {
			return nil, "", fmt.Errorf("%w: invalid status", pkgattachment.ErrInvalid)
		}
		q = q.Where("status = ?", in.Status)
	}
	if in.Search != "" {
		like := "%" + strings.ToLower(in.Search) + "%"
		q = q.Where(
			"LOWER(filename) LIKE ? OR LOWER(COALESCE(original_name, '')) LIKE ? OR LOWER(object_key) LIKE ? OR LOWER(mime_type) LIKE ?",
			like,
			like,
			like,
			like,
		)
	}
	if in.ReferenceStatus != "" {
		switch in.ReferenceStatus {
		case "referenced":
			q = q.Where("EXISTS (SELECT 1 FROM identity.users u WHERE u.avatar_attachment_id = attachment.attachments.id AND u.deleted_at IS NULL)")
		case "orphan":
			q = q.Where("NOT EXISTS (SELECT 1 FROM identity.users u WHERE u.avatar_attachment_id = attachment.attachments.id AND u.deleted_at IS NULL)")
		default:
			return nil, "", fmt.Errorf("%w: invalid reference_status", pkgattachment.ErrInvalid)
		}
	}
	if in.CursorID > 0 {
		q = q.Where("id < ?", in.CursorID)
	}
	if in.CategoryID != nil {
		q = q.Joins("JOIN attachment.attachment_categories ac ON ac.attachment_id = attachment.attachments.id AND ac.category_id = ?", *in.CategoryID)
	} else if in.CategoryMode != "" {
		switch in.CategoryMode {
		case "unassigned":
			q = q.Where("NOT EXISTS (SELECT 1 FROM attachment.attachment_categories ac WHERE ac.attachment_id = attachment.attachments.id)")
		default:
			return nil, "", fmt.Errorf("%w: invalid category_mode", pkgattachment.ErrInvalid)
		}
	}
	var rows []model.Attachment
	if err := q.Find(&rows).Error; err != nil {
		return nil, "", pkgattachment.MapDBError(err)
	}
	next := ""
	if len(rows) > limit {
		next = strconv.FormatInt(rows[limit-1].ID, 10)
		rows = rows[:limit]
	}
	return rows, next, nil
}

func (h *AttachmentsController) getAdmin(ctx context.Context, actor pkgattachment.Actor, id int64) (model.Attachment, []int64, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.Attachment{}, nil, err
	}
	return h.getWithCategories(ctx, id)
}

// getPublicContent resolves content for public rows or private rows when admin is true.
// getPublicContent 解析公开内容，或在 admin 为真时解析私有内容。
func (h *AttachmentsController) getPublicContent(ctx context.Context, id int64, admin bool) (pkgattachment.ContentResult, error) {
	attachment, _, err := h.getWithCategories(ctx, id)
	if err != nil {
		return pkgattachment.ContentResult{}, err
	}
	if attachment.UploadStatus != pkgattachment.UploadReady || attachment.Status != pkgattachment.StatusActive {
		return pkgattachment.ContentResult{}, pkgattachment.ErrNotFound
	}
	if !admin && attachment.AccessScope != pkgattachment.AccessPublic {
		return pkgattachment.ContentResult{}, pkgattachment.ErrForbidden
	}

	mount, be, err := driver.ResolveMountForRead(ctx, h.driverStore, h.driverRegistry, attachment.StorageMountID)
	if err != nil {
		return pkgattachment.ContentResult{}, err
	}
	if attachment.ObjectKey == "" {
		return pkgattachment.ContentResult{}, fmt.Errorf("%w: object key is missing", pkgattachment.ErrInvalid)
	}
	if mount.DriverName == driver.DriverLocal {
		localPath, err := be.LocalFilePath(attachment.ObjectKey)
		if err != nil {
			return pkgattachment.ContentResult{}, err
		}
		return pkgattachment.ContentResult{Attachment: attachment, LocalPath: localPath}, nil
	}
	presigned, err := be.PresignDownload(ctx, attachment.ObjectKey, time.Duration(pkgattachment.PresignTTLSeconds)*time.Second)
	if err != nil {
		return pkgattachment.ContentResult{}, err
	}
	return pkgattachment.ContentResult{Attachment: attachment, RedirectURL: presigned.URL}, nil
}

func (h *AttachmentsController) getWithCategories(ctx context.Context, id int64) (model.Attachment, []int64, error) {
	var attachment model.Attachment
	if err := h.db.WithContext(ctx).First(&attachment, "id = ?", id).Error; err != nil {
		return model.Attachment{}, nil, pkgattachment.MapDBError(err)
	}
	var bindings []model.AttachmentCategoryBinding
	if err := h.db.WithContext(ctx).Where("attachment_id = ?", id).Find(&bindings).Error; err != nil {
		return model.Attachment{}, nil, pkgattachment.MapDBError(err)
	}
	categoryIDs := make([]int64, 0, len(bindings))
	for _, binding := range bindings {
		categoryIDs = append(categoryIDs, binding.CategoryID)
	}
	return attachment, categoryIDs, nil
}
