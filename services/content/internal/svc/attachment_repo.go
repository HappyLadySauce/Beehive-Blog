package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
)

func (s *contentStore) ListAttachments(ctx context.Context, in *pb.ListAttachmentsRequest) (*pb.ListAttachmentsResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	page, pageSize := normalizePage(in.Page, in.PageSize)
	offset := (page - 1) * pageSize

	args := []any{}
	conds := []string{"1=1"}
	if in.ContentId > 0 {
		args = append(args, in.ContentId)
		conds = append(conds, fmt.Sprintf("ca.content_id = $%d", len(args)))
	}
	args = append(args, pageSize, offset)
	query := `
SELECT
	a.id, ca.content_id, a.storage_provider, a.bucket, a.object_key, a.file_name, a.mime_type, a.ext, a.size_bytes, ca.usage_type
FROM content_attachments ca
INNER JOIN attachments a ON a.id = ca.attachment_id
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY ca.id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	var rows []attachmentRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	list := make([]*pb.Attachment, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Attachment{
			Id:              rows[i].ID,
			ContentId:       rows[i].ContentID,
			StorageProvider: rows[i].StorageProvider,
			Bucket:          rows[i].Bucket,
			ObjectKey:       rows[i].ObjectKey,
			FileName:        rows[i].FileName,
			MimeType:        rows[i].MimeType,
			Ext:             rows[i].Ext,
			SizeBytes:       rows[i].SizeBytes,
			UsageType:       rows[i].UsageType,
		})
	}
	return &pb.ListAttachmentsResponse{List: list}, nil
}

func (s *contentStore) CreateAttachment(ctx context.Context, in *pb.CreateAttachmentRequest) (*pb.Attachment, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	if in.ContentId <= 0 || strings.TrimSpace(in.ObjectKey) == "" || strings.TrimSpace(in.FileName) == "" {
		return nil, fmt.Errorf("content_id, object_key and file_name are required")
	}
	provider := defaultIfEmpty(in.StorageProvider, "local")
	usageType := defaultIfEmpty(in.UsageType, "inline")

	var attachmentID int64
	insertAttachment := `
INSERT INTO attachments (storage_provider, bucket, object_key, file_name, mime_type, ext, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id`
	if err := s.conn.QueryRowCtx(ctx, &attachmentID, insertAttachment, provider, in.Bucket, strings.TrimSpace(in.ObjectKey), strings.TrimSpace(in.FileName), strings.TrimSpace(in.MimeType), strings.TrimSpace(in.Ext), in.SizeBytes); err != nil {
		return nil, err
	}

	insertMapping := `
INSERT INTO content_attachments (content_id, attachment_id, usage_type)
VALUES ($1, $2, $3)`
	if _, err := s.conn.ExecCtx(ctx, insertMapping, in.ContentId, attachmentID, usageType); err != nil {
		return nil, err
	}

	return &pb.Attachment{
		Id:              attachmentID,
		ContentId:       in.ContentId,
		StorageProvider: provider,
		Bucket:          in.Bucket,
		ObjectKey:       strings.TrimSpace(in.ObjectKey),
		FileName:        strings.TrimSpace(in.FileName),
		MimeType:        strings.TrimSpace(in.MimeType),
		Ext:             strings.TrimSpace(in.Ext),
		SizeBytes:       in.SizeBytes,
		UsageType:       usageType,
	}, nil
}

func (s *contentStore) DeleteAttachment(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	deleteMapping := `DELETE FROM content_attachments WHERE attachment_id = $1`
	if _, err := s.conn.ExecCtx(ctx, deleteMapping, id); err != nil {
		return err
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	return err
}
