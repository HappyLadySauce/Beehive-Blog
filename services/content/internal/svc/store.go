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

type contentRecord struct {
	ID           int64      `db:"id"`
	Type         string     `db:"type"`
	Title        string     `db:"title"`
	Slug         string     `db:"slug"`
	Summary      string     `db:"summary"`
	BodyMarkdown string     `db:"body_markdown"`
	Status       string     `db:"status"`
	Visibility   string     `db:"visibility"`
	AiAccess     string     `db:"ai_access"`
	PublishedAt  *time.Time `db:"published_at"`
}

type tagRecord struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Slug        string `db:"slug"`
	Color       string `db:"color"`
	Description string `db:"description"`
}

type relationRecord struct {
	ID              int64  `db:"id"`
	SourceContentID int64  `db:"source_content_id"`
	TargetContentID int64  `db:"target_content_id"`
	RelationType    string `db:"relation_type"`
	Weight          int32  `db:"weight"`
	Note            string `db:"note"`
}

type attachmentRecord struct {
	ID              int64  `db:"id"`
	ContentID       int64  `db:"content_id"`
	StorageProvider string `db:"storage_provider"`
	Bucket          string `db:"bucket"`
	ObjectKey       string `db:"object_key"`
	FileName        string `db:"file_name"`
	MimeType        string `db:"mime_type"`
	Ext             string `db:"ext"`
	SizeBytes       int64  `db:"size_bytes"`
	UsageType       string `db:"usage_type"`
}

type commentRecord struct {
	ID             int64  `db:"id"`
	ContentID      int64  `db:"content_id"`
	ParentID       int64  `db:"parent_comment_id"`
	AuthorName     string `db:"author_name"`
	AuthorEmail    string `db:"author_email"`
	BodyMarkdown   string `db:"body_markdown"`
	Status         string `db:"status"`
	ModerationNote string `db:"moderation_note"`
}

type contentStore struct {
	conn sqlx.SqlConn
}

func newContentStore(conn sqlx.SqlConn) (*contentStore, error) {
	s := &contentStore{conn: conn}
	if err := s.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *contentStore) Create(ctx context.Context, in *pb.CreateContentRequest) (*pb.ContentDetail, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	typ := strings.TrimSpace(in.Type)
	title := strings.TrimSpace(in.Title)
	slug := strings.TrimSpace(in.Slug)
	if typ == "" || title == "" || slug == "" {
		return nil, fmt.Errorf("type/title/slug are required")
	}

	visibility := defaultIfEmpty(in.Visibility, "private")
	aiAccess := defaultIfEmpty(in.AiAccess, "denied")
	if !isAllowedVisibility(visibility) {
		return nil, fmt.Errorf("invalid visibility")
	}
	if !isAllowedAiAccess(aiAccess) {
		return nil, fmt.Errorf("invalid ai_access")
	}

	var out contentRecord
	query := `
INSERT INTO content_items (type, title, slug, summary, body_markdown, status, visibility, ai_access)
VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7)
RETURNING id, type, title, slug, summary, body_markdown, status, visibility, ai_access, published_at`
	if err := s.conn.QueryRowCtx(ctx, &out, query, typ, title, slug, in.Summary, in.BodyMarkdown, visibility, aiAccess); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("slug already exists")
		}
		return nil, err
	}
	if err := s.publishContentEvent(ctx, events.TopicContentCreated, out.ID); err != nil {
		return nil, err
	}
	return toDetail(&out), nil
}

func (s *contentStore) Get(ctx context.Context, id int64) (*pb.ContentDetail, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	var out contentRecord
	query := `SELECT id, type, title, slug, summary, body_markdown, status, visibility, ai_access, published_at FROM content_items WHERE id = $1 LIMIT 1`
	if err := s.conn.QueryRowCtx(ctx, &out, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("content not found")
		}
		return nil, err
	}
	return toDetail(&out), nil
}

func (s *contentStore) Update(ctx context.Context, in *pb.UpdateContentRequest) (*pb.ContentDetail, error) {
	if in == nil || in.Id <= 0 {
		return nil, fmt.Errorf("invalid request")
	}
	visibility := strings.TrimSpace(in.Visibility)
	aiAccess := strings.TrimSpace(in.AiAccess)
	if visibility != "" && !isAllowedVisibility(visibility) {
		return nil, fmt.Errorf("invalid visibility")
	}
	if aiAccess != "" && !isAllowedAiAccess(aiAccess) {
		return nil, fmt.Errorf("invalid ai_access")
	}

	query := `
UPDATE content_items
SET
	title = CASE WHEN $2 <> '' THEN $2 ELSE title END,
	summary = CASE WHEN $3 <> '' THEN $3 ELSE summary END,
	body_markdown = CASE WHEN $4 <> '' THEN $4 ELSE body_markdown END,
	visibility = CASE WHEN $5 <> '' THEN $5 ELSE visibility END,
	ai_access = CASE WHEN $6 <> '' THEN $6 ELSE ai_access END,
	updated_at = NOW()
WHERE id = $1`
	if _, err := s.conn.ExecCtx(ctx, query, in.Id, strings.TrimSpace(in.Title), in.Summary, in.BodyMarkdown, visibility, aiAccess); err != nil {
		return nil, err
	}
	out, err := s.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if err := s.publishContentEvent(ctx, events.TopicContentUpdated, in.Id); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *contentStore) UpdateStatus(ctx context.Context, in *pb.UpdateStatusRequest) (*pb.ContentDetail, error) {
	if in == nil || in.Id <= 0 || strings.TrimSpace(in.Status) == "" {
		return nil, fmt.Errorf("invalid request")
	}
	status := strings.TrimSpace(in.Status)
	if !isAllowedStatus(status) {
		return nil, fmt.Errorf("invalid status")
	}
	query := `
UPDATE content_items
SET
	status = $2,
	published_at = CASE WHEN $2 = 'published' AND published_at IS NULL THEN NOW() ELSE published_at END,
	updated_at = NOW()
WHERE id = $1`
	if _, err := s.conn.ExecCtx(ctx, query, in.Id, status); err != nil {
		return nil, err
	}
	out, err := s.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if err := s.publishContentEvent(ctx, events.TopicContentStatusChanged, in.Id); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *contentStore) List(ctx context.Context, in *pb.ListContentsRequest, publicOnly bool) (*pb.ListContentsResponse, error) {
	page, pageSize := normalizePage(in.GetPage(), in.GetPageSize())
	offset := (page - 1) * pageSize

	args := []any{}
	conds := []string{"1=1"}
	if publicOnly {
		conds = append(conds, "status = 'published'", "visibility = 'public'")
	}
	if v := strings.TrimSpace(in.GetType()); v != "" {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf("type = $%d", len(args)))
	}
	if v := strings.TrimSpace(in.GetStatus()); v != "" {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if kw := strings.TrimSpace(in.GetKeyword()); kw != "" {
		args = append(args, "%"+strings.ToLower(kw)+"%")
		conds = append(conds, fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(summary) LIKE $%d)", len(args), len(args)))
	}

	args = append(args, pageSize, offset)
	query := `
SELECT id, type, title, slug, summary, body_markdown, status, visibility, ai_access, published_at
FROM content_items
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	var rows []contentRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	list := make([]*pb.ContentSummary, 0, len(rows))
	for i := range rows {
		list = append(list, toSummary(&rows[i]))
	}
	return &pb.ListContentsResponse{List: list}, nil
}

func (s *contentStore) ensureSchema(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS content_items (
	id BIGSERIAL PRIMARY KEY,
	type VARCHAR(32) NOT NULL,
	title VARCHAR(255) NOT NULL,
	slug VARCHAR(255) NOT NULL UNIQUE,
	summary TEXT NOT NULL DEFAULT '',
	body_markdown TEXT NOT NULL DEFAULT '',
	status VARCHAR(32) NOT NULL DEFAULT 'draft',
	visibility VARCHAR(32) NOT NULL DEFAULT 'private',
	ai_access VARCHAR(32) NOT NULL DEFAULT 'denied',
	published_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`
	_, err := s.conn.ExecCtx(ctx, query)
	return err
}

func (s *contentStore) ListTags(ctx context.Context) (*pb.ListTagsResponse, error) {
	query := `SELECT id, name, slug, color, description FROM tags ORDER BY id DESC`
	var rows []tagRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	list := make([]*pb.Tag, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Tag{
			Id:          rows[i].ID,
			Name:        rows[i].Name,
			Slug:        rows[i].Slug,
			Color:       rows[i].Color,
			Description: rows[i].Description,
		})
	}
	return &pb.ListTagsResponse{List: list}, nil
}

func (s *contentStore) CreateTag(ctx context.Context, in *pb.CreateTagRequest) (*pb.Tag, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	name := strings.TrimSpace(in.Name)
	slug := strings.TrimSpace(in.Slug)
	if name == "" || slug == "" {
		return nil, fmt.Errorf("name and slug are required")
	}
	query := `
INSERT INTO tags(name, slug, color, description)
VALUES ($1, $2, $3, $4)
RETURNING id, name, slug, color, description`
	var out tagRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, name, slug, strings.TrimSpace(in.Color), in.Description); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("tag already exists")
		}
		return nil, err
	}
	return &pb.Tag{
		Id:          out.ID,
		Name:        out.Name,
		Slug:        out.Slug,
		Color:       out.Color,
		Description: out.Description,
	}, nil
}

func (s *contentStore) DeleteTag(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM tags WHERE id = $1`, id)
	return err
}

func (s *contentStore) ListRelations(ctx context.Context, contentID int64) (*pb.ListRelationsResponse, error) {
	if contentID <= 0 {
		return nil, fmt.Errorf("invalid content_id")
	}
	query := `
SELECT id, source_content_id, target_content_id, relation_type, weight, note
FROM content_relations
WHERE source_content_id = $1
ORDER BY id DESC`
	var rows []relationRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, contentID); err != nil {
		return nil, err
	}
	list := make([]*pb.Relation, 0, len(rows))
	for i := range rows {
		list = append(list, &pb.Relation{
			Id:              rows[i].ID,
			SourceContentId: rows[i].SourceContentID,
			TargetContentId: rows[i].TargetContentID,
			RelationType:    rows[i].RelationType,
			Weight:          rows[i].Weight,
			Note:            rows[i].Note,
		})
	}
	return &pb.ListRelationsResponse{List: list}, nil
}

func (s *contentStore) CreateRelation(ctx context.Context, in *pb.CreateRelationRequest) (*pb.Relation, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	if in.SourceContentId <= 0 || in.TargetContentId <= 0 || strings.TrimSpace(in.RelationType) == "" {
		return nil, fmt.Errorf("source_content_id,target_content_id,relation_type are required")
	}
	weight := in.Weight
	if weight <= 0 {
		weight = 1
	}
	query := `
INSERT INTO content_relations (source_content_id, target_content_id, relation_type, weight, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, source_content_id, target_content_id, relation_type, weight, note`
	var out relationRecord
	if err := s.conn.QueryRowCtx(ctx, &out, query, in.SourceContentId, in.TargetContentId, strings.TrimSpace(in.RelationType), weight, in.Note); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("relation already exists")
		}
		return nil, err
	}
	return &pb.Relation{
		Id:              out.ID,
		SourceContentId: out.SourceContentID,
		TargetContentId: out.TargetContentID,
		RelationType:    out.RelationType,
		Weight:          out.Weight,
		Note:            out.Note,
	}, nil
}

func (s *contentStore) DeleteRelation(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM content_relations WHERE id = $1`, id)
	return err
}

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

func toSummary(in *contentRecord) *pb.ContentSummary {
	return &pb.ContentSummary{
		Id:          in.ID,
		Type:        in.Type,
		Title:       in.Title,
		Slug:        in.Slug,
		Summary:     in.Summary,
		Status:      in.Status,
		Visibility:  in.Visibility,
		AiAccess:    in.AiAccess,
		PublishedAt: formatTime(in.PublishedAt),
	}
}

func toDetail(in *contentRecord) *pb.ContentDetail {
	return &pb.ContentDetail{
		Id:           in.ID,
		Type:         in.Type,
		Title:        in.Title,
		Slug:         in.Slug,
		Summary:      in.Summary,
		BodyMarkdown: in.BodyMarkdown,
		Status:       in.Status,
		Visibility:   in.Visibility,
		AiAccess:     in.AiAccess,
	}
}

func defaultIfEmpty(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func normalizePage(page, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

func isAllowedVisibility(v string) bool {
	switch v {
	case "public", "member", "private":
		return true
	default:
		return false
	}
}

func isAllowedAiAccess(v string) bool {
	switch v {
	case "allowed", "denied":
		return true
	default:
		return false
	}
}

func isAllowedStatus(v string) bool {
	switch v {
	case "draft", "review", "published", "archived":
		return true
	default:
		return false
	}
}

func (s *contentStore) publishContentEvent(ctx context.Context, eventType string, contentID int64) error {
	if strings.TrimSpace(eventType) == "" || contentID <= 0 {
		return fmt.Errorf("invalid outbox event")
	}
	query := `
INSERT INTO event_outbox (event_type, resource_type, resource_id, payload, status, attempts, next_retry_at, last_error)
VALUES ($1, 'content_item', $2, '{}'::jsonb, 'pending', 0, NOW(), '')`
	_, err := s.conn.ExecCtx(ctx, query, eventType, contentID)
	return err
}
