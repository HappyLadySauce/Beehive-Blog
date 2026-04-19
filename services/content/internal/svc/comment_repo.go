package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (s *contentStore) ListComments(ctx context.Context, in *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	if in == nil || in.ContentId <= 0 {
		return nil, fmt.Errorf("content_id is required")
	}
	page, pageSize := normalizePage(in.Page, in.PageSize)
	offset := (page - 1) * pageSize
	query := `
SELECT id, content_id, COALESCE(parent_comment_id, 0) AS parent_comment_id, author_name, author_email, body_markdown, status, moderation_note
FROM comments
WHERE content_id = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3`
	var rows []commentRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, in.ContentId, pageSize, offset); err != nil {
		return nil, err
	}
	list := make([]*pb.Comment, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Comment{
			Id:              rows[i].ID,
			ContentId:       rows[i].ContentID,
			ParentCommentId: rows[i].ParentID,
			AuthorName:      rows[i].AuthorName,
			AuthorEmail:     rows[i].AuthorEmail,
			BodyMarkdown:    rows[i].BodyMarkdown,
			Status:          rows[i].Status,
			ModerationNote:  rows[i].ModerationNote,
		})
	}
	return &pb.ListCommentsResponse{List: list}, nil
}

func (s *contentStore) CreateComment(ctx context.Context, in *pb.CreateCommentRequest) (*pb.Comment, error) {
	if in == nil || in.ContentId <= 0 || strings.TrimSpace(in.BodyMarkdown) == "" {
		return nil, fmt.Errorf("content_id and body_markdown are required")
	}
	query := `
INSERT INTO comments (content_id, parent_comment_id, author_name, author_email, body_markdown, status)
VALUES ($1, NULLIF($2, 0), $3, $4, $5, 'visible')
RETURNING id, content_id, COALESCE(parent_comment_id, 0) AS parent_comment_id, author_name, author_email, body_markdown, status, moderation_note`
	var out commentRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, in.ContentId, in.ParentCommentId, strings.TrimSpace(in.AuthorName), strings.TrimSpace(in.AuthorEmail), in.BodyMarkdown); err != nil {
		return nil, err
	}
	return &pb.Comment{
		Id:              out.ID,
		ContentId:       out.ContentID,
		ParentCommentId: out.ParentID,
		AuthorName:      out.AuthorName,
		AuthorEmail:     out.AuthorEmail,
		BodyMarkdown:    out.BodyMarkdown,
		Status:          out.Status,
		ModerationNote:  out.ModerationNote,
	}, nil
}

func (s *contentStore) UpdateCommentStatus(ctx context.Context, in *pb.UpdateCommentStatusRequest) (*pb.Comment, error) {
	if in == nil || in.Id <= 0 || strings.TrimSpace(in.Status) == "" {
		return nil, fmt.Errorf("id and status are required")
	}
	status := strings.TrimSpace(in.Status)
	if status != "visible" && status != "hidden" && status != "deleted" {
		return nil, fmt.Errorf("invalid status")
	}
	query := `
UPDATE comments
SET status = $2, moderation_note = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, content_id, COALESCE(parent_comment_id, 0) AS parent_comment_id, author_name, author_email, body_markdown, status, moderation_note`
	var out commentRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, in.Id, status, in.ModerationNote); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, err
	}
	return &pb.Comment{
		Id:              out.ID,
		ContentId:       out.ContentID,
		ParentCommentId: out.ParentID,
		AuthorName:      out.AuthorName,
		AuthorEmail:     out.AuthorEmail,
		BodyMarkdown:    out.BodyMarkdown,
		Status:          out.Status,
		ModerationNote:  out.ModerationNote,
	}, nil
}
