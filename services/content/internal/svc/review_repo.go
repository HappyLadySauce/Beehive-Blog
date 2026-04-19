package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/events"
	"github.com/HappyLadySauce/Beehive-Blog/services/content/pb"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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
