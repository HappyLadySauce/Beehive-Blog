package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/events"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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
