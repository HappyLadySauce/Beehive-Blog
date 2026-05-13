package attachments

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	pkgattachment "github.com/HappyLadySauce/Beehive-Blog/pkg/attachment"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/storage"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

// UploadLocal handles POST /api/v1/attachments multipart uploads (admin).
// UploadLocal 处理 POST /api/v1/attachments multipart 上传（管理员）。
func (h *AttachmentsController) UploadLocal(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, h.attachmentOptions.MaxBytes+1<<20)
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("file field is required", err))
		return
	}
	defer file.Close()

	ownerUserID, err := optionalInt64Form(ctx, "owner_user_id")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid owner_user_id", err))
		return
	}
	categoryIDs, err := int64ListForm(ctx, "category_ids")
	if err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid category_ids", err))
		return
	}
	mimeType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	originalName := strings.TrimSpace(ctx.PostForm("original_name"))
	if originalName == "" {
		originalName = header.Filename
	}
	in := pkgattachment.LocalUploadInput{
		OwnerUserID:  ownerUserID,
		Purpose:      defaultString(ctx.PostForm("purpose"), pkgattachment.PurposeContent),
		Filename:     header.Filename,
		OriginalName: &originalName,
		MimeType:     mimeType,
		Size:         header.Size,
		Reader:       file,
		AccessScope:  defaultString(ctx.PostForm("access_scope"), pkgattachment.AccessPrivate),
		CategoryIDs:  categoryIDs,
	}
	out, err := h.uploadLocal(ctx.Request.Context(), actorFromClaims(ctx), in)
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toAttachmentResponse(out, nil))
}

// PresignRemote handles POST /api/v1/attachments/upload-url (admin).
// PresignRemote 处理 POST /api/v1/attachments/upload-url（管理员）。
func (h *AttachmentsController) PresignRemote(ctx *gin.Context) {
	var req v1.AttachmentPresignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	out, err := h.presignRemote(ctx.Request.Context(), actorFromClaims(ctx), pkgattachment.RemotePresignInput{
		StorageType:  req.StorageType,
		OwnerUserID:  req.OwnerUserID,
		Purpose:      req.Purpose,
		Filename:     req.Filename,
		OriginalName: req.OriginalName,
		MimeType:     req.MimeType,
		Size:         req.Size,
		AccessScope:  req.AccessScope,
		Checksum:     req.Checksum,
		CategoryIDs:  req.CategoryIDs,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, v1.AttachmentPresignResponse{
		Attachment: toAttachmentResponse(out.Attachment, nil),
		UploadURL:  out.Upload.URL,
		Method:     out.Upload.Method,
		Headers:    out.Upload.Headers,
		ExpiresAt:  out.Upload.ExpiresAt,
	})
}

// CompleteRemote handles POST /api/v1/attachments/:id/complete (admin).
// CompleteRemote 处理 POST /api/v1/attachments/:id/complete（管理员）。
func (h *AttachmentsController) CompleteRemote(ctx *gin.Context) {
	id, ok := parseIDParam(ctx)
	if !ok {
		return
	}
	var req v1.AttachmentCompleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.Fail(ctx, common.NewBadRequest("invalid request body", err))
		return
	}
	out, err := h.completeRemote(ctx.Request.Context(), actorFromClaims(ctx), id, pkgattachment.CompleteInput{
		ETag:     req.ETag,
		Checksum: req.Checksum,
		Size:     req.Size,
	})
	if err != nil {
		writeAttachmentError(ctx, err)
		return
	}
	common.Success(ctx, toAttachmentResponse(out, nil))
}

// uploadLocal creates a ready local attachment from a server-side upload.
// uploadLocal 基于服务端上传创建 ready 状态本地附件。
func (h *AttachmentsController) uploadLocal(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.LocalUploadInput) (model.Attachment, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.Attachment{}, err
	}
	if err := pkgattachment.ValidateCommon(h.attachmentOptions, in.OwnerUserID, in.Purpose, in.MimeType, in.Size, in.AccessScope); err != nil {
		return model.Attachment{}, err
	}
	if in.Reader == nil {
		return model.Attachment{}, fmt.Errorf("%w: upload reader is required", pkgattachment.ErrInvalid)
	}
	if strings.TrimSpace(in.Filename) == "" {
		return model.Attachment{}, fmt.Errorf("%w: filename is required", pkgattachment.ErrInvalid)
	}

	objectKey, filename, err := pkgattachment.ObjectKeyFor(in.Purpose, in.Filename)
	if err != nil {
		return model.Attachment{}, err
	}
	backend, err := h.storage.Backend(options.AttachmentStorageLocal)
	if err != nil {
		return model.Attachment{}, err
	}
	stored, err := backend.Save(ctx, storage.PutRequest{ObjectKey: objectKey, Reader: in.Reader, Size: in.Size})
	if err != nil {
		return model.Attachment{}, err
	}

	attachment := model.Attachment{
		OwnerUserID:  in.OwnerUserID,
		Purpose:      in.Purpose,
		Filename:     filename,
		OriginalName: pkgattachment.CleanOptional(in.OriginalName),
		MimeType:     in.MimeType,
		Size:         in.Size,
		StorageType:  options.AttachmentStorageLocal,
		LocalPath:    &stored.LocalPath,
		ETag:         pkgattachment.CleanOptional(&stored.ETag),
		Checksum:     pkgattachment.CleanOptional(&stored.Checksum),
		AccessScope:  in.AccessScope,
		UploadStatus: pkgattachment.UploadReady,
		Status:       pkgattachment.StatusActive,
	}
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&attachment).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		return replaceCategoriesTx(tx, attachment.ID, in.CategoryIDs)
	})
	if err != nil {
		return model.Attachment{}, err
	}
	return attachment, nil
}

// presignRemote creates a pending remote attachment and direct-upload instructions.
// presignRemote 创建 pending 远端附件并返回直传指令。
func (h *AttachmentsController) presignRemote(ctx context.Context, actor pkgattachment.Actor, in pkgattachment.RemotePresignInput) (pkgattachment.PresignOutput, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return pkgattachment.PresignOutput{}, err
	}
	if in.StorageType != options.AttachmentStorageS3 && in.StorageType != options.AttachmentStorageOSS {
		return pkgattachment.PresignOutput{}, fmt.Errorf("%w: storage_type must be s3 or oss", pkgattachment.ErrInvalid)
	}
	if err := pkgattachment.ValidateCommon(h.attachmentOptions, in.OwnerUserID, in.Purpose, in.MimeType, in.Size, in.AccessScope); err != nil {
		return pkgattachment.PresignOutput{}, err
	}
	objectKey, filename, err := pkgattachment.ObjectKeyFor(in.Purpose, in.Filename)
	if err != nil {
		return pkgattachment.PresignOutput{}, err
	}
	backend, err := h.storage.Backend(in.StorageType)
	if err != nil {
		return pkgattachment.PresignOutput{}, err
	}
	upload, err := backend.PresignUpload(ctx, storage.PresignRequest{
		ObjectKey: objectKey,
		MimeType:  in.MimeType,
		Checksum:  pkgattachment.DerefString(in.Checksum),
		Size:      in.Size,
		TTL:       h.attachmentOptions.PresignTTL,
	})
	if err != nil {
		return pkgattachment.PresignOutput{}, err
	}

	bucket := upload.Bucket
	attachment := model.Attachment{
		OwnerUserID:  in.OwnerUserID,
		Purpose:      in.Purpose,
		Filename:     filename,
		OriginalName: pkgattachment.CleanOptional(in.OriginalName),
		MimeType:     in.MimeType,
		Size:         in.Size,
		StorageType:  in.StorageType,
		Bucket:       &bucket,
		ObjectKey:    &upload.ObjectKey,
		Checksum:     pkgattachment.CleanOptional(in.Checksum),
		AccessScope:  in.AccessScope,
		UploadStatus: pkgattachment.UploadPending,
		Status:       pkgattachment.StatusActive,
	}
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&attachment).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		return replaceCategoriesTx(tx, attachment.ID, in.CategoryIDs)
	})
	if err != nil {
		return pkgattachment.PresignOutput{}, err
	}
	return pkgattachment.PresignOutput{Attachment: attachment, Upload: upload}, nil
}

// completeRemote marks a pending remote attachment ready.
// completeRemote 将 pending 远端附件标记为 ready。
func (h *AttachmentsController) completeRemote(ctx context.Context, actor pkgattachment.Actor, id int64, in pkgattachment.CompleteInput) (model.Attachment, error) {
	if err := pkgattachment.RequireAdmin(actor); err != nil {
		return model.Attachment{}, err
	}
	var out model.Attachment
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, "id = ?", id).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		if out.StorageType == options.AttachmentStorageLocal {
			return fmt.Errorf("%w: local attachments do not need completion", pkgattachment.ErrInvalid)
		}
		if out.UploadStatus != pkgattachment.UploadPending {
			return fmt.Errorf("%w: attachment is not pending", pkgattachment.ErrConflict)
		}
		if in.Size != nil && *in.Size != out.Size {
			return fmt.Errorf("%w: size mismatch", pkgattachment.ErrInvalid)
		}
		updates := map[string]interface{}{
			"upload_status": pkgattachment.UploadReady,
			"updated_at":    time.Now(),
		}
		if in.ETag != nil {
			updates["etag"] = strings.TrimSpace(*in.ETag)
		}
		if in.Checksum != nil {
			updates["checksum"] = strings.TrimSpace(*in.Checksum)
		}
		if err := tx.Model(&out).Updates(updates).Error; err != nil {
			return pkgattachment.MapDBError(err)
		}
		return tx.First(&out, "id = ?", id).Error
	})
	if err != nil {
		return model.Attachment{}, err
	}
	return out, nil
}
