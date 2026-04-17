package svc

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type outboxEvent struct {
	ID         int64  `db:"id"`
	EventType  string `db:"event_type"`
	ResourceID int64  `db:"resource_id"`
	Attempts   int64  `db:"attempts"`
}

type outboxStore struct {
	conn sqlx.SqlConn
}

func newOutboxStore(conn sqlx.SqlConn) *outboxStore {
	return &outboxStore{conn: conn}
}

func (s *outboxStore) ClaimPending(ctx context.Context, batchSize int) ([]outboxEvent, error) {
	if batchSize <= 0 {
		batchSize = 20
	}
	query := `
WITH picked AS (
	SELECT id
	FROM event_outbox
	WHERE status = 'pending' AND next_retry_at <= NOW()
	ORDER BY id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
UPDATE event_outbox AS e
SET status = 'processing', updated_at = NOW()
FROM picked
WHERE e.id = picked.id
RETURNING e.id, e.event_type, e.resource_id, e.attempts`
	var rows []outboxEvent
	if err := s.conn.QueryRowsCtx(ctx, &rows, query, batchSize); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *outboxStore) MarkDone(ctx context.Context, id int64) error {
	_, err := s.conn.ExecCtx(ctx, `
UPDATE event_outbox
SET status = 'done', last_error = '', updated_at = NOW()
WHERE id = $1`, id)
	return err
}

func (s *outboxStore) MarkRetry(ctx context.Context, id int64, currentAttempts int64, maxAttempts int, backoff time.Duration, cause error) error {
	nextAttempts := currentAttempts + 1
	message := strings.TrimSpace(cause.Error())
	if maxAttempts <= 0 {
		maxAttempts = 8
	}
	if len(message) > 1000 {
		message = message[:1000]
	}

	status := "pending"
	nextRetryAt := time.Now().UTC().Add(backoff * time.Duration(1<<minInt64(nextAttempts-1, 6)))
	if int(nextAttempts) >= maxAttempts {
		status = "failed"
		nextRetryAt = time.Now().UTC()
	}

	_, err := s.conn.ExecCtx(ctx, `
UPDATE event_outbox
SET status = $2, attempts = $3, next_retry_at = $4, last_error = $5, updated_at = NOW()
WHERE id = $1`, id, status, nextAttempts, nextRetryAt, message)
	return err
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
