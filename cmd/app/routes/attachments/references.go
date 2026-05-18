package attachments

import (
	"context"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// GetReferences handles GET /api/v1/attachments/:id/references (admin).
// GetReferences 处理 GET /api/v1/attachments/:id/references（管理员）。
func (h *AttachmentsController) GetReferences(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	items, err := h.listReferences(ctx.Request.Context(), actorFromClaims(ctx), &id)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, v1.AttachmentReferenceListResponse{Items: items})
}

// ListReferences handles GET /api/v1/attachments/references (admin).
// ListReferences 处理 GET /api/v1/attachments/references（管理员）。
func (h *AttachmentsController) ListReferences(ctx *gin.Context) {
	attachmentID, err := optionalInt64Query(ctx, "attachment_id")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid attachment_id", err))
		return
	}
	items, err := h.listReferences(ctx.Request.Context(), actorFromClaims(ctx), attachmentID)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, v1.AttachmentReferenceListResponse{Items: items})
}

func (h *AttachmentsController) listReferences(ctx context.Context, actor pkgattachment.Actor, attachmentID *int64) ([]v1.AttachmentReferenceResponse, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return nil, err
	}
	q := h.db.WithContext(ctx).Model(&model.User{}).Where("avatar_attachment_id IS NOT NULL")
	if attachmentID != nil {
		q = q.Where("avatar_attachment_id = ?", *attachmentID)
	}
	var users []model.User
	if err := q.Order("updated_at DESC").Find(&users).Error; err != nil {
		return nil, pkgattachment.MapDBError(err)
	}
	items := make([]v1.AttachmentReferenceResponse, 0, len(users))
	for _, user := range users {
		if user.AvatarAttachmentID == nil {
			continue
		}
		items = append(items, toUserAvatarReference(*user.AvatarAttachmentID, user))
	}
	return items, nil
}

func (h *AttachmentsController) hasReferences(ctx context.Context, attachmentID int64) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&model.User{}).
		Where("avatar_attachment_id = ?", attachmentID).
		Count(&count).Error; err != nil {
		return false, pkgattachment.MapDBError(err)
	}
	return count > 0, nil
}

func toUserAvatarReference(attachmentID int64, user model.User) v1.AttachmentReferenceResponse {
	title := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		title = *user.Nickname
	}
	return v1.AttachmentReferenceResponse{
		AttachmentID: attachmentID,
		SourceType:   "user",
		SourceID:     user.ID,
		SourceTitle:  title,
		Relation:     "avatar",
		Status:       user.Status,
		UpdatedAt:    user.UpdatedAt,
	}
}
