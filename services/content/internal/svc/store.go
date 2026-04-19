package svc

import (
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
	return &contentStore{conn: conn}, nil
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
