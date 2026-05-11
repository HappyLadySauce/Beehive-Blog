package attachments

import (
	"errors"

	"github.com/gin-gonic/gin"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

func toAttachmentResponse(row model.Attachment, categoryIDs []int64) v1.AttachmentResponse {
	resp := v1.AttachmentResponse{
		ID:           row.ID,
		OwnerUserID:  row.OwnerUserID,
		Purpose:      row.Purpose,
		Filename:     row.Filename,
		OriginalName: row.OriginalName,
		MimeType:     row.MimeType,
		Size:         row.Size,
		StorageType:  row.StorageType,
		Bucket:       row.Bucket,
		ObjectKey:    row.ObjectKey,
		LocalPath:    row.LocalPath,
		ETag:         row.ETag,
		Checksum:     row.Checksum,
		AccessScope:  row.AccessScope,
		UploadStatus: row.UploadStatus,
		Status:       row.Status,
		CategoryIDs:  categoryIDs,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		resp.DeletedAt = &row.DeletedAt.Time
	}
	return resp
}

func toPublicAttachmentResponse(row model.Attachment) v1.AttachmentResponse {
	return v1.AttachmentResponse{
		ID:           row.ID,
		Purpose:      row.Purpose,
		Filename:     row.Filename,
		OriginalName: row.OriginalName,
		MimeType:     row.MimeType,
		Size:         row.Size,
		AccessScope:  row.AccessScope,
		UploadStatus: row.UploadStatus,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toCategoryResponse(row model.AttachmentCategory) v1.AttachmentCategoryResponse {
	resp := v1.AttachmentCategoryResponse{
		ID:          row.ID,
		ParentID:    row.ParentID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		Icon:        row.Icon,
		Path:        row.Path,
		Depth:       row.Depth,
		SortOrder:   row.SortOrder,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		resp.DeletedAt = &row.DeletedAt.Time
	}
	return resp
}

func writeAttachmentError(ctx *gin.Context, err error) {
	common.Fail(ctx, toAppError(err))
}

func toAppError(err error) error {
	switch {
	case errors.Is(err, pkgattachment.ErrInvalid):
		return common.NewBadRequest("invalid attachment request", err)
	case errors.Is(err, pkgattachment.ErrForbidden):
		return common.NewForbidden("forbidden", err)
	case errors.Is(err, pkgattachment.ErrNotFound):
		return common.NewNotFound("attachment resource not found", err)
	case errors.Is(err, pkgattachment.ErrConflict):
		return common.NewConflict("attachment conflict", err)
	default:
		return common.NewInternal("attachment operation failed", err)
	}
}
