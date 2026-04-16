package svc

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	return s.Get(ctx, in.Id)
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
	return s.Get(ctx, in.Id)
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
