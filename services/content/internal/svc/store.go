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

type projectProfileRecord struct {
	ContentID   int64      `db:"content_id"`
	ProjectName string     `db:"project_name"`
	Stack       string     `db:"stack"`
	RepoURL     string     `db:"repo_url"`
	DemoURL     string     `db:"demo_url"`
	StartedAt   *time.Time `db:"started_at"`
	EndedAt     *time.Time `db:"ended_at"`
}

type experienceProfileRecord struct {
	ContentID int64      `db:"content_id"`
	OrgName   string     `db:"org_name"`
	RoleName  string     `db:"role_name"`
	Location  string     `db:"location"`
	StartedAt *time.Time `db:"started_at"`
	EndedAt   *time.Time `db:"ended_at"`
}

type timelineEventProfileRecord struct {
	ContentID     int64      `db:"content_id"`
	EventTime     *time.Time `db:"event_time"`
	EventCategory string     `db:"event_category"`
}

type portfolioProfileRecord struct {
	ContentID    int64  `db:"content_id"`
	ArtifactType string `db:"artifact_type"`
	ExternalLink string `db:"external_link"`
}

type revisionRecord struct {
	ID           int64     `db:"id"`
	ContentID    int64     `db:"content_id"`
	Version      int64     `db:"version"`
	Title        string    `db:"title"`
	Summary      string    `db:"summary"`
	BodyMarkdown string    `db:"body_markdown"`
	ChangeNote   string    `db:"change_note"`
	CreatedBy    int64     `db:"created_by"`
	CreatedAt    time.Time `db:"created_at"`
}

type reviewTaskRecord struct {
	ID              int64      `db:"id"`
	ContentID       int64      `db:"content_id"`
	RevisionID      int64      `db:"revision_id"`
	SubmitterUserID int64      `db:"submitter_user_id"`
	ReviewerUserID  int64      `db:"reviewer_user_id"`
	SourceType      string     `db:"source_type"`
	Status          string     `db:"status"`
	Priority        int32      `db:"priority"`
	Note            string     `db:"note"`
	DecidedAt       *time.Time `db:"decided_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
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

const (
	contentTypeArticle        = "article"
	contentTypeNote           = "note"
	contentTypeProject        = "project"
	contentTypeExperience     = "experience"
	contentTypeTimeline       = "timeline_event"
	contentTypePortfolio      = "portfolio"
	contentTypePage           = "page"
	contentTypeInsight        = "insight"
	defaultCreateRevisionNote = "initial version"
	defaultUpdateRevisionNote = "content updated"
	reviewStatusPending       = "pending"
	reviewStatusApproved      = "approved"
	reviewStatusRejected      = "rejected"
	reviewStatusCancelled     = "cancelled"
	reviewSourceHuman         = "human"
	reviewSourceSystem        = "system"
)

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
	typ := normalizeContentType(in.Type)
	title := strings.TrimSpace(in.Title)
	slug := strings.TrimSpace(in.Slug)
	if typ == "" || title == "" || slug == "" {
		return nil, fmt.Errorf("type/title/slug are required")
	}
	if !isAllowedContentType(typ) {
		return nil, fmt.Errorf("invalid content type")
	}
	if err := validateProfilePayloadByType(typ, in.ProjectProfile, in.ExperienceProfile, in.TimelineEventProfile, in.PortfolioProfile); err != nil {
		return nil, err
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
	if err := s.upsertProfileByType(ctx, out.ID, typ, in.ProjectProfile, in.ExperienceProfile, in.TimelineEventProfile, in.PortfolioProfile); err != nil {
		return nil, err
	}
	if err := s.createRevisionFromRecord(ctx, &out, strings.TrimSpace(in.ChangeNote), 0); err != nil {
		return nil, err
	}
	if err := s.publishContentEvent(ctx, events.TopicContentCreated, out.ID); err != nil {
		return nil, err
	}
	return s.Get(ctx, out.ID)
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
	detail := toDetail(&out)
	if err := s.fillProfileDetail(ctx, detail); err != nil {
		return nil, err
	}
	return detail, nil
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
	contentType, err := s.getContentType(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if err := validateProfilePayloadByType(contentType, in.ProjectProfile, in.ExperienceProfile, in.TimelineEventProfile, in.PortfolioProfile); err != nil {
		return nil, err
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
	if hasAnyProfile(in.ProjectProfile, in.ExperienceProfile, in.TimelineEventProfile, in.PortfolioProfile) {
		if err := s.upsertProfileByType(ctx, in.Id, contentType, in.ProjectProfile, in.ExperienceProfile, in.TimelineEventProfile, in.PortfolioProfile); err != nil {
			return nil, err
		}
	}
	latest, err := s.getContentRecord(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if err := s.createRevisionFromRecord(ctx, latest, strings.TrimSpace(in.ChangeNote), 0); err != nil {
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
	if err := s.conn.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		query := `
UPDATE content_items
SET
	status = $2,
	published_at = CASE WHEN $2 = 'published' AND published_at IS NULL THEN NOW() ELSE published_at END,
	updated_at = NOW()
WHERE id = $1`
		if _, err := session.ExecCtx(txCtx, query, in.Id, status); err != nil {
			return err
		}
		if status == "review" {
			if _, err := s.submitReviewWithSession(txCtx, session, in.Id, "", 0, reviewSourceSystem); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
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

func (s *contentStore) ListReviews(ctx context.Context, in *pb.ReviewListRequest) (*pb.ReviewListResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("empty request")
	}
	page, pageSize := normalizePage(in.Page, in.PageSize)
	offset := (page - 1) * pageSize

	args := []any{}
	conds := []string{"1=1"}
	if status := strings.TrimSpace(in.Status); status != "" {
		if !isAllowedReviewStatus(status) {
			return nil, fmt.Errorf("invalid review status")
		}
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}

	countQuery := `SELECT COUNT(1) FROM review_tasks WHERE ` + strings.Join(conds, " AND ")
	var total int64
	if err := s.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	args = append(args, pageSize, offset)
	query := `
SELECT
	id,
	COALESCE(content_id, 0) AS content_id,
	COALESCE(revision_id, 0) AS revision_id,
	COALESCE(submitter_user_id, 0) AS submitter_user_id,
	COALESCE(reviewer_user_id, 0) AS reviewer_user_id,
	source_type,
	status,
	priority,
	note,
	decided_at,
	created_at,
	updated_at
FROM review_tasks
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY created_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	var rows []reviewTaskRecord
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	list := make([]*pb.ReviewTask, 0, len(rows))
	for i := range rows {
		list = append(list, toReviewTask(&rows[i]))
	}
	return &pb.ReviewListResponse{
		List:       list,
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *contentStore) SubmitReview(ctx context.Context, in *pb.SubmitReviewRequest) (*pb.ReviewTask, error) {
	if in == nil || in.ContentId <= 0 {
		return nil, fmt.Errorf("content_id is required")
	}
	note := strings.TrimSpace(in.Note)
	priority := normalizeReviewPriority(in.Priority)
	submitterUserID := in.SubmitterUserId

	var out *reviewTaskRecord
	if err := s.conn.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		task, err := s.submitReviewWithSession(txCtx, session, in.ContentId, note, priority, reviewSourceHuman)
		if err != nil {
			return err
		}
		if task.SubmitterUserID == 0 && submitterUserID > 0 {
			if _, err := session.ExecCtx(txCtx, `
UPDATE review_tasks
SET submitter_user_id = $2, updated_at = NOW()
WHERE id = $1`, task.ID, submitterUserID); err != nil {
				return err
			}
			task.SubmitterUserID = submitterUserID
		}
		out = task
		return nil
	}); err != nil {
		return nil, err
	}
	return toReviewTask(out), nil
}

func (s *contentStore) ApproveReview(ctx context.Context, in *pb.ApproveReviewRequest) (*pb.ReviewTask, error) {
	return s.decideReview(ctx, in.GetId(), in.GetReviewerUserId(), in.GetReason(), reviewStatusApproved)
}

func (s *contentStore) RejectReview(ctx context.Context, in *pb.RejectReviewRequest) (*pb.ReviewTask, error) {
	return s.decideReview(ctx, in.GetId(), in.GetReviewerUserId(), in.GetReason(), reviewStatusRejected)
}

func (s *contentStore) decideReview(ctx context.Context, reviewID, reviewerUserID int64, reason, decision string) (*pb.ReviewTask, error) {
	if reviewID <= 0 || reviewerUserID <= 0 {
		return nil, fmt.Errorf("review id and reviewer_user_id are required")
	}
	if decision != reviewStatusApproved && decision != reviewStatusRejected {
		return nil, fmt.Errorf("invalid review decision")
	}

	var out *reviewTaskRecord
	if err := s.conn.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		task, err := s.getReviewTaskByID(txCtx, session, reviewID, true)
		if err != nil {
			return err
		}
		if task.Status != reviewStatusPending {
			return fmt.Errorf("review task is not pending")
		}
		if task.ReviewerUserID > 0 && task.ReviewerUserID != reviewerUserID {
			return fmt.Errorf("review task already claimed")
		}
		if task.ReviewerUserID == 0 {
			if _, err := session.ExecCtx(txCtx, `
UPDATE review_tasks
SET reviewer_user_id = $2, updated_at = NOW()
WHERE id = $1`, task.ID, reviewerUserID); err != nil {
				return err
			}
			task.ReviewerUserID = reviewerUserID
		}

		reason = strings.TrimSpace(reason)
		if _, err := session.ExecCtx(txCtx, `
INSERT INTO review_decisions (review_task_id, decision, reason, decided_by)
VALUES ($1, $2, $3, $4)`, task.ID, decision, reason, reviewerUserID); err != nil {
			return err
		}
		if _, err := session.ExecCtx(txCtx, `
UPDATE review_tasks
SET status = $2, decided_at = NOW(), updated_at = NOW()
WHERE id = $1`, task.ID, decision); err != nil {
			return err
		}
		if task.ContentID <= 0 {
			return fmt.Errorf("review task has no content")
		}

		targetContentStatus := "published"
		if decision == reviewStatusRejected {
			targetContentStatus = "draft"
		}
		if _, err := session.ExecCtx(txCtx, `
UPDATE content_items
SET
	status = $2,
	published_at = CASE WHEN $2 = 'published' AND published_at IS NULL THEN NOW() ELSE published_at END,
	updated_at = NOW()
WHERE id = $1`, task.ContentID, targetContentStatus); err != nil {
			return err
		}

		latest, err := s.getReviewTaskByID(txCtx, session, task.ID, false)
		if err != nil {
			return err
		}
		out = latest
		return nil
	}); err != nil {
		return nil, err
	}

	if err := s.publishContentEvent(ctx, events.TopicContentStatusChanged, out.ContentID); err != nil {
		return nil, err
	}
	return toReviewTask(out), nil
}

func (s *contentStore) submitReviewWithSession(ctx context.Context, session sqlx.Session, contentID int64, note string, priority int32, sourceType string) (*reviewTaskRecord, error) {
	if contentID <= 0 {
		return nil, fmt.Errorf("content_id is required")
	}
	if sourceType == "" {
		sourceType = reviewSourceHuman
	}
	if sourceType != reviewSourceHuman && sourceType != reviewSourceSystem {
		return nil, fmt.Errorf("invalid review source type")
	}
	priority = normalizeReviewPriority(priority)
	if _, err := s.getContentTypeWithSession(ctx, session, contentID); err != nil {
		return nil, err
	}

	existing, err := s.findPendingReviewTaskByContent(ctx, session, contentID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	revisionID, err := s.getLatestRevisionID(ctx, session, contentID)
	if err != nil {
		return nil, err
	}
	note = strings.TrimSpace(note)

	var out reviewTaskRecord
	query := `
INSERT INTO review_tasks (content_id, revision_id, source_type, status, priority, note)
VALUES ($1, NULLIF($2, 0), $3, $4, $5, $6)
RETURNING
	id,
	COALESCE(content_id, 0) AS content_id,
	COALESCE(revision_id, 0) AS revision_id,
	COALESCE(submitter_user_id, 0) AS submitter_user_id,
	COALESCE(reviewer_user_id, 0) AS reviewer_user_id,
	source_type,
	status,
	priority,
	note,
	decided_at,
	created_at,
	updated_at`
	if err := session.QueryRowCtx(ctx, &out, query, contentID, revisionID, sourceType, reviewStatusPending, priority, note); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *contentStore) findPendingReviewTaskByContent(ctx context.Context, session sqlx.Session, contentID int64) (*reviewTaskRecord, error) {
	var out reviewTaskRecord
	query := `
SELECT
	id,
	COALESCE(content_id, 0) AS content_id,
	COALESCE(revision_id, 0) AS revision_id,
	COALESCE(submitter_user_id, 0) AS submitter_user_id,
	COALESCE(reviewer_user_id, 0) AS reviewer_user_id,
	source_type,
	status,
	priority,
	note,
	decided_at,
	created_at,
	updated_at
FROM review_tasks
WHERE content_id = $1 AND status = $2
ORDER BY id DESC
LIMIT 1`
	if err := session.QueryRowCtx(ctx, &out, query, contentID, reviewStatusPending); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (s *contentStore) getReviewTaskByID(ctx context.Context, session sqlx.Session, reviewID int64, forUpdate bool) (*reviewTaskRecord, error) {
	if reviewID <= 0 {
		return nil, fmt.Errorf("review id is required")
	}
	query := `
SELECT
	id,
	COALESCE(content_id, 0) AS content_id,
	COALESCE(revision_id, 0) AS revision_id,
	COALESCE(submitter_user_id, 0) AS submitter_user_id,
	COALESCE(reviewer_user_id, 0) AS reviewer_user_id,
	source_type,
	status,
	priority,
	note,
	decided_at,
	created_at,
	updated_at
FROM review_tasks
WHERE id = $1
LIMIT 1`
	if forUpdate {
		query += " FOR UPDATE"
	}
	var out reviewTaskRecord
	if err := session.QueryRowCtx(ctx, &out, query, reviewID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, fmt.Errorf("review task not found")
		}
		return nil, err
	}
	return &out, nil
}

func (s *contentStore) getLatestRevisionID(ctx context.Context, session sqlx.Session, contentID int64) (int64, error) {
	var revisionID int64
	query := `SELECT COALESCE((SELECT id FROM content_revisions WHERE content_id = $1 ORDER BY version DESC LIMIT 1), 0)`
	if err := session.QueryRowCtx(ctx, &revisionID, query, contentID); err != nil {
		return 0, err
	}
	return revisionID, nil
}

func (s *contentStore) getContentTypeWithSession(ctx context.Context, session sqlx.Session, contentID int64) (string, error) {
	var contentType string
	if err := session.QueryRowCtx(ctx, &contentType, `SELECT type FROM content_items WHERE id = $1 LIMIT 1`, contentID); err != nil {
		if err == sqlx.ErrNotFound {
			return "", fmt.Errorf("content not found")
		}
		return "", err
	}
	contentType = normalizeContentType(contentType)
	if !isAllowedContentType(contentType) {
		return "", fmt.Errorf("invalid content type")
	}
	return contentType, nil
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

func (s *contentStore) getContentType(ctx context.Context, contentID int64) (string, error) {
	var contentType string
	if err := s.conn.QueryRowCtx(ctx, &contentType, `SELECT type FROM content_items WHERE id = $1 LIMIT 1`, contentID); err != nil {
		if err == sqlx.ErrNotFound {
			return "", fmt.Errorf("content not found")
		}
		return "", err
	}
	contentType = normalizeContentType(contentType)
	if !isAllowedContentType(contentType) {
		return "", fmt.Errorf("invalid content type")
	}
	return contentType, nil
}

func (s *contentStore) fillProfileDetail(ctx context.Context, detail *pb.ContentDetail) error {
	if detail == nil {
		return nil
	}
	switch normalizeContentType(detail.Type) {
	case contentTypeProject:
		var row projectProfileRecord
		if err := s.conn.QueryRowCtx(ctx, &row, `SELECT content_id, project_name, stack, repo_url, demo_url, started_at, ended_at FROM project_profiles WHERE content_id = $1 LIMIT 1`, detail.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return nil
			}
			return err
		}
		detail.ProjectProfile = &pb.ProjectProfile{
			ProjectName: row.ProjectName,
			Stack:       row.Stack,
			RepoUrl:     row.RepoURL,
			DemoUrl:     row.DemoURL,
			StartedAt:   formatDate(row.StartedAt),
			EndedAt:     formatDate(row.EndedAt),
		}
	case contentTypeExperience:
		var row experienceProfileRecord
		if err := s.conn.QueryRowCtx(ctx, &row, `SELECT content_id, org_name, role_name, location, started_at, ended_at FROM experience_profiles WHERE content_id = $1 LIMIT 1`, detail.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return nil
			}
			return err
		}
		detail.ExperienceProfile = &pb.ExperienceProfile{
			OrgName:   row.OrgName,
			RoleName:  row.RoleName,
			Location:  row.Location,
			StartedAt: formatDate(row.StartedAt),
			EndedAt:   formatDate(row.EndedAt),
		}
	case contentTypeTimeline:
		var row timelineEventProfileRecord
		if err := s.conn.QueryRowCtx(ctx, &row, `SELECT content_id, event_time, event_category FROM timeline_event_profiles WHERE content_id = $1 LIMIT 1`, detail.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return nil
			}
			return err
		}
		detail.TimelineEventProfile = &pb.TimelineEventProfile{
			EventTime:     formatTime(row.EventTime),
			EventCategory: row.EventCategory,
		}
	case contentTypePortfolio:
		var row portfolioProfileRecord
		if err := s.conn.QueryRowCtx(ctx, &row, `SELECT content_id, artifact_type, external_link FROM portfolio_profiles WHERE content_id = $1 LIMIT 1`, detail.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return nil
			}
			return err
		}
		detail.PortfolioProfile = &pb.PortfolioProfile{
			ArtifactType: row.ArtifactType,
			ExternalLink: row.ExternalLink,
		}
	}
	return nil
}

func (s *contentStore) upsertProfileByType(ctx context.Context, contentID int64, contentType string, project *pb.ProjectProfile, experience *pb.ExperienceProfile, timeline *pb.TimelineEventProfile, portfolio *pb.PortfolioProfile) error {
	switch normalizeContentType(contentType) {
	case contentTypeProject:
		if project == nil {
			_, err := s.conn.ExecCtx(ctx, `INSERT INTO project_profiles(content_id) VALUES ($1) ON CONFLICT (content_id) DO NOTHING`, contentID)
			return err
		}
		_, err := s.conn.ExecCtx(ctx, `
INSERT INTO project_profiles(content_id, project_name, stack, repo_url, demo_url, started_at, ended_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::date, NULLIF($7, '')::date, NOW())
ON CONFLICT (content_id) DO UPDATE
SET
	project_name = EXCLUDED.project_name,
	stack = EXCLUDED.stack,
	repo_url = EXCLUDED.repo_url,
	demo_url = EXCLUDED.demo_url,
	started_at = EXCLUDED.started_at,
	ended_at = EXCLUDED.ended_at,
	updated_at = NOW()`,
			contentID,
			strings.TrimSpace(project.ProjectName),
			strings.TrimSpace(project.Stack),
			strings.TrimSpace(project.RepoUrl),
			strings.TrimSpace(project.DemoUrl),
			strings.TrimSpace(project.StartedAt),
			strings.TrimSpace(project.EndedAt),
		)
		return err
	case contentTypeExperience:
		if experience == nil {
			_, err := s.conn.ExecCtx(ctx, `INSERT INTO experience_profiles(content_id) VALUES ($1) ON CONFLICT (content_id) DO NOTHING`, contentID)
			return err
		}
		_, err := s.conn.ExecCtx(ctx, `
INSERT INTO experience_profiles(content_id, org_name, role_name, location, started_at, ended_at, updated_at)
VALUES ($1, $2, $3, $4, NULLIF($5, '')::date, NULLIF($6, '')::date, NOW())
ON CONFLICT (content_id) DO UPDATE
SET
	org_name = EXCLUDED.org_name,
	role_name = EXCLUDED.role_name,
	location = EXCLUDED.location,
	started_at = EXCLUDED.started_at,
	ended_at = EXCLUDED.ended_at,
	updated_at = NOW()`,
			contentID,
			strings.TrimSpace(experience.OrgName),
			strings.TrimSpace(experience.RoleName),
			strings.TrimSpace(experience.Location),
			strings.TrimSpace(experience.StartedAt),
			strings.TrimSpace(experience.EndedAt),
		)
		return err
	case contentTypeTimeline:
		if timeline == nil {
			_, err := s.conn.ExecCtx(ctx, `INSERT INTO timeline_event_profiles(content_id) VALUES ($1) ON CONFLICT (content_id) DO NOTHING`, contentID)
			return err
		}
		_, err := s.conn.ExecCtx(ctx, `
INSERT INTO timeline_event_profiles(content_id, event_time, event_category, updated_at)
VALUES ($1, NULLIF($2, '')::timestamptz, $3, NOW())
ON CONFLICT (content_id) DO UPDATE
SET
	event_time = EXCLUDED.event_time,
	event_category = EXCLUDED.event_category,
	updated_at = NOW()`,
			contentID,
			strings.TrimSpace(timeline.EventTime),
			strings.TrimSpace(timeline.EventCategory),
		)
		return err
	case contentTypePortfolio:
		if portfolio == nil {
			_, err := s.conn.ExecCtx(ctx, `INSERT INTO portfolio_profiles(content_id) VALUES ($1) ON CONFLICT (content_id) DO NOTHING`, contentID)
			return err
		}
		_, err := s.conn.ExecCtx(ctx, `
INSERT INTO portfolio_profiles(content_id, artifact_type, external_link, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (content_id) DO UPDATE
SET
	artifact_type = EXCLUDED.artifact_type,
	external_link = EXCLUDED.external_link,
	updated_at = NOW()`,
			contentID,
			strings.TrimSpace(portfolio.ArtifactType),
			strings.TrimSpace(portfolio.ExternalLink),
		)
		return err
	default:
		return nil
	}
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

func toReviewTask(in *reviewTaskRecord) *pb.ReviewTask {
	if in == nil {
		return &pb.ReviewTask{}
	}
	return &pb.ReviewTask{
		Id:              in.ID,
		ContentId:       in.ContentID,
		RevisionId:      in.RevisionID,
		SubmitterUserId: in.SubmitterUserID,
		ReviewerUserId:  in.ReviewerUserID,
		SourceType:      in.SourceType,
		Status:          in.Status,
		Priority:        in.Priority,
		Note:            in.Note,
		DecidedAt:       formatTime(in.DecidedAt),
		CreatedAt:       in.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       in.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func defaultIfEmpty(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func normalizeContentType(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
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

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02")
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

func isAllowedReviewStatus(v string) bool {
	switch v {
	case reviewStatusPending, reviewStatusApproved, reviewStatusRejected, reviewStatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeReviewPriority(priority int32) int32 {
	if priority <= 0 {
		return 3
	}
	if priority > 9 {
		return 9
	}
	return priority
}

func isAllowedContentType(v string) bool {
	switch normalizeContentType(v) {
	case contentTypeArticle, contentTypeNote, contentTypeProject, contentTypeExperience, contentTypeTimeline, contentTypePortfolio, contentTypePage, contentTypeInsight:
		return true
	default:
		return false
	}
}

func hasAnyProfile(project *pb.ProjectProfile, experience *pb.ExperienceProfile, timeline *pb.TimelineEventProfile, portfolio *pb.PortfolioProfile) bool {
	return project != nil || experience != nil || timeline != nil || portfolio != nil
}

func validateProfilePayloadByType(contentType string, project *pb.ProjectProfile, experience *pb.ExperienceProfile, timeline *pb.TimelineEventProfile, portfolio *pb.PortfolioProfile) error {
	contentType = normalizeContentType(contentType)
	switch contentType {
	case contentTypeProject:
		if experience != nil || timeline != nil || portfolio != nil {
			return fmt.Errorf("project only accepts project_profile")
		}
	case contentTypeExperience:
		if project != nil || timeline != nil || portfolio != nil {
			return fmt.Errorf("experience only accepts experience_profile")
		}
	case contentTypeTimeline:
		if project != nil || experience != nil || portfolio != nil {
			return fmt.Errorf("timeline_event only accepts timeline_event_profile")
		}
	case contentTypePortfolio:
		if project != nil || experience != nil || timeline != nil {
			return fmt.Errorf("portfolio only accepts portfolio_profile")
		}
	case contentTypePage:
		if hasAnyProfile(project, experience, timeline, portfolio) {
			return fmt.Errorf("page does not accept profile payload")
		}
	default:
		if hasAnyProfile(project, experience, timeline, portfolio) {
			return fmt.Errorf("content type does not accept profile payload")
		}
	}
	return nil
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
