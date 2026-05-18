package attachments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
	in := pkgattachment.ListInput{
		OwnerUserID:     ownerUserID,
		Purpose:         strings.TrimSpace(ctx.Query("purpose")),
		Status:          strings.TrimSpace(ctx.Query("status")),
		CategoryID:      categoryID,
		CategoryMode:    strings.TrimSpace(ctx.Query("category_mode")),
		Search:          strings.TrimSpace(ctx.Query("search")),
		ReferenceStatus: strings.TrimSpace(ctx.Query("reference_status")),
	}

	if hasPageQuery(ctx) {
		page, err := optionalPage(ctx.Query("page"))
		if err != nil {
			common.Fail(ctx, common.NewBadRequest("invalid page", err))
			return
		}
		pageSize, err := optionalPageSize(ctx.Query("page_size"))
		if err != nil {
			common.Fail(ctx, common.NewBadRequest("invalid page_size", err))
			return
		}
		in.Page = page
		in.PageSize = pageSize
		rows, total, page, pageSize, err := h.listOffset(ctx.Request.Context(), actorFromClaims(ctx), in)
		if err != nil {
			writeAttachmentError(ctx, err)
			return
		}
		items, err := h.toAttachmentListResponses(ctx.Request.Context(), rows)
		if err != nil {
			writeAttachmentError(ctx, err)
			return
		}
		common.Success(ctx, v1.AttachmentListResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
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
	in.CursorID = cursorID
	in.Limit = limit
	rows, next, err := h.listCursor(ctx.Request.Context(), actorFromClaims(ctx), in)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	items, err := h.toAttachmentListResponses(ctx.Request.Context(), rows)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
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
	ownerNames, err := h.ownerUsernames(ctx.Request.Context(), []model.Attachment{row})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toAttachmentResponseWithOwner(row, categoryIDs, ownerNameFor(row, ownerNames)))
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

func (h *AttachmentsController) buildListQuery(ctx context.Context, in pkgattachment.ListInput) (*gorm.DB, error) {
	q := h.db.WithContext(ctx).Model(&model.Attachment{})
	if in.OwnerUserID != nil {
		q = q.Where("owner_user_id = ?", *in.OwnerUserID)
	}
	if in.Purpose != "" {
		if !pkgattachment.PurposeKnown(in.Purpose) {
			return nil, fmt.Errorf("%w: invalid purpose", pkgattachment.ErrInvalid)
		}
		q = q.Where("purpose = ?", in.Purpose)
	}
	if in.Status != "" {
		if !pkgattachment.StatusKnown(in.Status) {
			return nil, fmt.Errorf("%w: invalid status", pkgattachment.ErrInvalid)
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
			return nil, fmt.Errorf("%w: invalid reference_status", pkgattachment.ErrInvalid)
		}
	}
	if in.CategoryID != nil {
		q = q.Joins("JOIN attachment.attachment_categories ac ON ac.attachment_id = attachment.attachments.id AND ac.category_id = ?", *in.CategoryID)
	} else if in.CategoryMode != "" {
		switch in.CategoryMode {
		case "unassigned":
			q = q.Where("NOT EXISTS (SELECT 1 FROM attachment.attachment_categories ac WHERE ac.attachment_id = attachment.attachments.id)")
		default:
			return nil, fmt.Errorf("%w: invalid category_mode", pkgattachment.ErrInvalid)
		}
	}
	return q, nil
}

func (h *AttachmentsController) toAttachmentListResponses(ctx context.Context, rows []model.Attachment) ([]v1.AttachmentResponse, error) {
	ownerNames, err := h.ownerUsernames(ctx, rows)
	if err != nil {
		return nil, err
	}
	items := make([]v1.AttachmentResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAttachmentResponseWithOwner(row, nil, ownerNameFor(row, ownerNames)))
	}
	return items, nil
}

func (h *AttachmentsController) ownerUsernames(ctx context.Context, rows []model.Attachment) (map[int64]string, error) {
	ids := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, row := range rows {
		if row.OwnerUserID == nil {
			continue
		}
		if _, ok := seen[*row.OwnerUserID]; ok {
			continue
		}
		seen[*row.OwnerUserID] = struct{}{}
		ids = append(ids, *row.OwnerUserID)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	if err := h.db.WithContext(ctx).Select("id", "username").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, pkgattachment.MapDBError(err)
	}
	names := make(map[int64]string, len(users))
	for _, user := range users {
		names[user.ID] = user.Username
	}
	return names, nil
}

func ownerNameFor(row model.Attachment, names map[int64]string) *string {
	if row.OwnerUserID == nil || names == nil {
		return nil
	}
	name, ok := names[*row.OwnerUserID]
	if !ok {
		return nil
	}
	return &name
}

func (h *AttachmentsController) listCursor(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.ListInput) ([]model.Attachment, string, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return nil, "", err
	}
	limit := in.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q, err := h.buildListQuery(ctx, in)
	if err != nil {
		return nil, "", err
	}
	if in.CursorID > 0 {
		q = q.Where("id < ?", in.CursorID)
	}
	q = q.Order("id DESC").Limit(limit + 1)
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

func (h *AttachmentsController) listOffset(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.ListInput) ([]model.Attachment, int64, int, int, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return nil, 0, 0, 0, err
	}
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	q, err := h.buildListQuery(ctx, in)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, pkgattachment.MapDBError(err)
	}
	var rows []model.Attachment
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, 0, 0, pkgattachment.MapDBError(err)
	}
	return rows, total, page, pageSize, nil
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
