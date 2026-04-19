package svc

import (
	"context"
	"fmt"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/events"
)

func (s *contentStore) publishContentEvent(ctx context.Context, eventType string, contentID int64) error {
	if strings.TrimSpace(eventType) == "" || contentID <= 0 {
		return fmt.Errorf("invalid outbox event")
	}
	query := `
INSERT INTO event_outbox (event_type, resource_type, resource_id, payload, status, attempts, next_retry_at, last_error)
VALUES ($1, 'content_item', $2, '{}'::jsonb, 'pending', 0, NOW(), '')`
	_, err := s.conn.ExecCtx(ctx, query, normalizeContentEventType(eventType), contentID)
	return err
}

func normalizeContentEventType(eventType string) string {
	eventType = strings.TrimSpace(eventType)
	switch eventType {
	case events.TopicContentCreated, events.TopicContentUpdated, events.TopicContentStatusChanged:
		return eventType
	default:
		return eventType
	}
}
