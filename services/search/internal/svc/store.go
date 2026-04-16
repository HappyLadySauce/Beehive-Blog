package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/HappyLadySauce/Beehive-Blog/services/search/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	defaultLang     = "zh"
	defaultPageSize = int64(20)
	maxPageSize     = int64(50)
	chunkMaxRunes   = 600
)

var markdownCleanupRe = regexp.MustCompile("[`*_>#\\[\\]()!~-]")

type searchStore struct {
	conn sqlx.SqlConn
}

type queryRecord struct {
	ContentID   int64   `db:"content_id"`
	ContentType string  `db:"content_type"`
	Title       string  `db:"title"`
	Slug        string  `db:"slug"`
	Summary     string  `db:"summary"`
	BodyPlain   string  `db:"body_plain"`
	Score       float64 `db:"score"`
}

type documentMeta struct {
	ContentType string `json:"type"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Visibility  string `json:"visibility"`
	AiAccess    string `json:"ai_access"`
}

type indexRecord struct {
	ContentID   int64     `db:"content_id"`
	ContentType string    `db:"content_type"`
	Title       string    `db:"title"`
	Slug        string    `db:"slug"`
	Status      string    `db:"status"`
	Visibility  string    `db:"visibility"`
	IndexedAt   time.Time `db:"indexed_at"`
}

func newSearchStore(conn sqlx.SqlConn) (*searchStore, error) {
	return &searchStore{conn: conn}, nil
}

func (s *searchStore) Query(ctx context.Context, in *pb.SearchRequest) (*pb.SearchResponse, error) {
	kw := strings.TrimSpace(in.GetQuery())
	if kw == "" {
		return &pb.SearchResponse{List: []*pb.SearchResultItem{}}, nil
	}

	page, pageSize := normalizePage(in.GetPage(), in.GetPageSize())
	offset := (page - 1) * pageSize
	scope := normalizeScope(in.GetScope())

	pattern := "%" + strings.ToLower(kw) + "%"
	args := []any{pattern}
	conds := []string{"1=1"}
	if t := strings.TrimSpace(in.GetType()); t != "" {
		args = append(args, t)
		conds = append(conds, fmt.Sprintf("metadata->>'type' = $%d", len(args)))
	}
	if scope == "public" {
		conds = append(conds, "metadata->>'status' = 'published'", "metadata->>'visibility' = 'public'")
	}
	args = append(args, pageSize, offset)

	query := `
SELECT
	content_id,
	metadata->>'type' AS content_type,
	title,
	metadata->>'slug' AS slug,
	summary,
	body_plain,
	(
		CASE WHEN LOWER(title) LIKE $1 THEN 3 ELSE 0 END +
		CASE WHEN LOWER(summary) LIKE $1 THEN 2 ELSE 0 END +
		CASE WHEN LOWER(body_plain) LIKE $1 THEN 1 ELSE 0 END
	)::float8 AS score
FROM search_documents
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY score DESC, indexed_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	var rows []queryRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	items := make([]*pb.SearchResultItem, 0, len(rows))
	for i := range rows {
		items = append(items, &pb.SearchResultItem{
			ContentId: rows[i].ContentID,
			Type:      rows[i].ContentType,
			Title:     rows[i].Title,
			Slug:      rows[i].Slug,
			Summary:   rows[i].Summary,
			Highlight: buildHighlight(rows[i].Summary, rows[i].BodyPlain, kw),
			Score:     rows[i].Score,
		})
	}
	return &pb.SearchResponse{List: items}, nil
}

func (s *searchStore) UpsertDocument(ctx context.Context, in *pb.UpsertDocumentRequest) (*pb.IndexDocument, error) {
	if in == nil || in.ContentId <= 0 {
		return nil, fmt.Errorf("content_id is required")
	}
	title := strings.TrimSpace(in.Title)
	slug := strings.TrimSpace(in.Slug)
	if title == "" || slug == "" {
		return nil, fmt.Errorf("title and slug are required")
	}

	meta := documentMeta{
		ContentType: strings.TrimSpace(in.Type),
		Slug:        slug,
		Status:      normalizeStatus(in.Status),
		Visibility:  normalizeVisibility(in.Visibility),
		AiAccess:    normalizeAiAccess(in.AiAccess),
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	bodyPlain := markdownToPlain(in.BodyMarkdown)

	var indexed indexRecord
	upsert := `
INSERT INTO search_documents(content_id, language, title, summary, body_plain, metadata, indexed_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, NOW(), NOW())
ON CONFLICT (content_id) DO UPDATE
SET
	language = EXCLUDED.language,
	title = EXCLUDED.title,
	summary = EXCLUDED.summary,
	body_plain = EXCLUDED.body_plain,
	metadata = EXCLUDED.metadata,
	indexed_at = NOW(),
	updated_at = NOW()
RETURNING
	content_id,
	metadata->>'type' AS content_type,
	title,
	metadata->>'slug' AS slug,
	metadata->>'status' AS status,
	metadata->>'visibility' AS visibility,
	indexed_at`
	if err := s.conn.QueryRowCtx(ctx, &indexed, upsert, in.ContentId, defaultLang, title, in.Summary, bodyPlain, string(metaBytes)); err != nil {
		return nil, err
	}

	if err := s.rebuildChunks(ctx, in.ContentId, bodyPlain); err != nil {
		return nil, err
	}

	return &pb.IndexDocument{
		ContentId:  indexed.ContentID,
		Type:       indexed.ContentType,
		Title:      indexed.Title,
		Slug:       indexed.Slug,
		Status:     indexed.Status,
		Visibility: indexed.Visibility,
		IndexedAt:  indexed.IndexedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *searchStore) DeleteDocument(ctx context.Context, contentID int64) error {
	if contentID <= 0 {
		return fmt.Errorf("content_id is required")
	}
	_, err := s.conn.ExecCtx(ctx, `DELETE FROM search_documents WHERE content_id = $1`, contentID)
	return err
}

func (s *searchStore) rebuildChunks(ctx context.Context, contentID int64, body string) error {
	if _, err := s.conn.ExecCtx(ctx, `DELETE FROM content_chunks WHERE content_id = $1`, contentID); err != nil {
		return err
	}
	chunks := splitIntoChunks(body, chunkMaxRunes)
	for idx, text := range chunks {
		if strings.TrimSpace(text) == "" {
			continue
		}
		tokenCount := estimateTokenCount(text)
		insert := `
INSERT INTO content_chunks (content_id, chunk_no, chunk_text, token_count, embedding_model)
VALUES ($1, $2, $3, $4, '')`
		if _, err := s.conn.ExecCtx(ctx, insert, contentID, idx+1, text, tokenCount); err != nil {
			return err
		}
	}
	return nil
}

func normalizePage(page, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func normalizeScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "owner":
		return "owner"
	default:
		return "public"
	}
}

func normalizeStatus(v string) string {
	switch strings.TrimSpace(v) {
	case "review", "published", "archived":
		return strings.TrimSpace(v)
	default:
		return "draft"
	}
}

func normalizeVisibility(v string) string {
	switch strings.TrimSpace(v) {
	case "public", "member":
		return strings.TrimSpace(v)
	default:
		return "private"
	}
}

func normalizeAiAccess(v string) string {
	if strings.TrimSpace(v) == "allowed" {
		return "allowed"
	}
	return "denied"
}

func markdownToPlain(in string) string {
	s := strings.ReplaceAll(in, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = markdownCleanupRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func buildHighlight(summary, body, kw string) string {
	kw = strings.ToLower(strings.TrimSpace(kw))
	if kw == "" {
		return ""
	}
	for _, src := range []string{summary, body} {
		clean := strings.TrimSpace(src)
		if clean == "" {
			continue
		}
		lower := strings.ToLower(clean)
		idx := strings.Index(lower, kw)
		if idx < 0 {
			continue
		}
		start := idx - 40
		if start < 0 {
			start = 0
		}
		end := idx + len(kw) + 80
		if end > len(clean) {
			end = len(clean)
		}
		return strings.TrimSpace(clean[start:end])
	}
	if len(summary) > 120 {
		return summary[:120]
	}
	return summary
}

func splitIntoChunks(text string, maxRunes int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}
	parts := strings.Split(text, "\n\n")
	chunks := make([]string, 0, len(parts))
	var builder strings.Builder

	flush := func() {
		if strings.TrimSpace(builder.String()) == "" {
			builder.Reset()
			return
		}
		chunks = append(chunks, strings.TrimSpace(builder.String()))
		builder.Reset()
	}

	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(p)
		if utf8.RuneCountInString(builder.String()) >= maxRunes {
			flush()
		}
	}
	flush()

	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		if utf8.RuneCountInString(c) <= maxRunes {
			out = append(out, c)
			continue
		}
		runes := []rune(c)
		for start := 0; start < len(runes); start += maxRunes {
			end := start + maxRunes
			if end > len(runes) {
				end = len(runes)
			}
			out = append(out, strings.TrimSpace(string(runes[start:end])))
		}
	}
	return out
}

func estimateTokenCount(s string) int {
	if s == "" {
		return 0
	}
	return utf8.RuneCountInString(s) / 2
}
