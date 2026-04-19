package svc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/events"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (s *contentStore) ListRevisions(ctx context.Context, in *pb.RevisionListRequest) (*pb.RevisionListResponse, error) {
	if in == nil || in.ContentId <= 0 {
		return nil, fmt.Errorf("content_id is required")
	}
	if _, err := s.getContentType(ctx, in.ContentId); err != nil {
		return nil, err
	}

	page, pageSize := normalizePage(in.Page, in.PageSize)
	offset := (page - 1) * pageSize

	var total int64
	if err := s.conn.QueryRowCtx(ctx, &total, `SELECT COUNT(1) FROM content_revisions WHERE content_id = $1`, in.ContentId); err != nil {
		return nil, err
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	var rows []revisionRecord
	query := `
SELECT id, content_id, version, title, summary, body_markdown, change_note, COALESCE(created_by, 0) AS created_by, created_at
FROM content_revisions
WHERE content_id = $1
ORDER BY version DESC
LIMIT $2 OFFSET $3`
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, in.ContentId, pageSize, offset); err != nil {
		return nil, err
	}

	list := make([]*pb.RevisionSummary, 0, len(rows))
	for _, row := range rows {
		list = append(list, &pb.RevisionSummary{
			Id:         row.ID,
			ContentId:  row.ContentID,
			Version:    row.Version,
			Title:      row.Title,
			Summary:    row.Summary,
			ChangeNote: row.ChangeNote,
			CreatedBy:  row.CreatedBy,
			CreatedAt:  row.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &pb.RevisionListResponse{
		List:       list,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *contentStore) GetRevision(ctx context.Context, in *pb.GetRevisionRequest) (*pb.RevisionDetail, error) {
	if in == nil || in.ContentId <= 0 || in.RevisionId <= 0 {
		return nil, fmt.Errorf("content_id and revision_id are required")
	}
	row, err := s.getRevisionRecord(ctx, in.ContentId, in.RevisionId)
	if err != nil {
		return nil, err
	}
	return &pb.RevisionDetail{
		Id:           row.ID,
		ContentId:    row.ContentID,
		Version:      row.Version,
		Title:        row.Title,
		Summary:      row.Summary,
		BodyMarkdown: row.BodyMarkdown,
		ChangeNote:   row.ChangeNote,
		CreatedBy:    row.CreatedBy,
		CreatedAt:    row.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *contentStore) RestoreRevision(ctx context.Context, in *pb.RestoreRevisionRequest) (*pb.ContentDetail, error) {
	if in == nil || in.ContentId <= 0 || in.RevisionId <= 0 {
		return nil, fmt.Errorf("content_id and revision_id are required")
	}
	row, err := s.getRevisionRecord(ctx, in.ContentId, in.RevisionId)
	if err != nil {
		return nil, err
	}

	if _, err := s.conn.ExecCtx(ctx, `
UPDATE content_items
SET title = $2, summary = $3, body_markdown = $4, updated_at = NOW()
WHERE id = $1`, in.ContentId, row.Title, row.Summary, row.BodyMarkdown); err != nil {
		return nil, err
	}

	latest, err := s.getContentRecord(ctx, in.ContentId)
	if err != nil {
		return nil, err
	}
	note := fmt.Sprintf("restored from revision #%d", row.Version)
	if err := s.createRevisionFromRecord(ctx, latest, note, 0); err != nil {
		return nil, err
	}
	if err := s.publishContentEvent(ctx, events.TopicContentUpdated, in.ContentId); err != nil {
		return nil, err
	}
	return s.Get(ctx, in.ContentId)
}

func (s *contentStore) getRevisionRecord(ctx context.Context, contentID, revisionID int64) (*revisionRecord, error) {
	var row revisionRecord
	query := `
SELECT id, content_id, version, title, summary, body_markdown, change_note, COALESCE(created_by, 0) AS created_by, created_at
FROM content_revisions
WHERE content_id = $1 AND id = $2
LIMIT 1`
	if err := s.conn.QueryRowCtx(ctx, &row, query, contentID, revisionID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("revision not found")
		}
		return nil, err
	}
	return &row, nil
}

func (s *contentStore) getContentRecord(ctx context.Context, contentID int64) (*contentRecord, error) {
	var out contentRecord
	query := `SELECT id, type, title, slug, summary, body_markdown, status, visibility, ai_access, published_at FROM content_items WHERE id = $1 LIMIT 1`
	if err := s.conn.QueryRowCtx(ctx, &out, query, contentID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("content not found")
		}
		return nil, err
	}
	return &out, nil
}

func (s *contentStore) createRevisionFromRecord(ctx context.Context, item *contentRecord, changeNote string, createdBy int64) error {
	if item == nil || item.ID <= 0 {
		return fmt.Errorf("invalid content record")
	}
	var version int64
	if err := s.conn.QueryRowCtx(ctx, &version, `SELECT COALESCE(MAX(version), 0) + 1 FROM content_revisions WHERE content_id = $1`, item.ID); err != nil {
		return err
	}
	changeNote = strings.TrimSpace(changeNote)
	if changeNote == "" {
		if version == 1 {
			changeNote = defaultCreateRevisionNote
		} else {
			changeNote = defaultUpdateRevisionNote
		}
	}

	var createdByAny any
	if createdBy > 0 {
		createdByAny = createdBy
	}
	_, err := s.conn.ExecCtx(ctx, `
INSERT INTO content_revisions(content_id, version, title, summary, body_markdown, change_note, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		item.ID, version, item.Title, item.Summary, item.BodyMarkdown, changeNote, createdByAny)
	return err
}

func (s *contentStore) getLatestRevisionID(ctx context.Context, session sqlx.Session, contentID int64) (int64, error) {
	var revisionID int64
	query := `SELECT COALESCE((SELECT id FROM content_revisions WHERE content_id = $1 ORDER BY version DESC LIMIT 1), 0)`
	if err := session.QueryRowCtx(ctx, &revisionID, query, contentID); err != nil {
		return 0, err
	}
	return revisionID, nil
}
