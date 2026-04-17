package worker

import (
	"context"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/events"
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/svc"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
	"github.com/zeromicro/go-zero/core/logx"
)

type IndexerWorker struct {
	svcCtx       *svc.ServiceContext
	pollInterval time.Duration
	retryBackoff time.Duration
	batchSize    int
	maxAttempts  int
}

type BatchResult struct {
	Picked    int64
	Succeeded int64
	Failed    int64
}

func NewIndexerWorker(svcCtx *svc.ServiceContext) *IndexerWorker {
	poll := parseDurationOrDefault(svcCtx.Config.Worker.PollInterval, 2*time.Second)
	backoff := parseDurationOrDefault(svcCtx.Config.Worker.RetryBackoff, 5*time.Second)
	batchSize := svcCtx.Config.Worker.BatchSize
	if batchSize <= 0 {
		batchSize = 20
	}
	maxAttempts := svcCtx.Config.Worker.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 8
	}

	return &IndexerWorker{
		svcCtx:       svcCtx,
		pollInterval: poll,
		retryBackoff: backoff,
		batchSize:    batchSize,
		maxAttempts:  maxAttempts,
	}
}

func (w *IndexerWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.ProcessBatch(ctx, 0)
	for {
		select {
		case <-ctx.Done():
			logx.WithContext(ctx).Info("indexer worker stopped")
			return
		case <-ticker.C:
			w.ProcessBatch(ctx, 0)
		}
	}
}

func (w *IndexerWorker) ProcessBatch(ctx context.Context, overrideBatchSize int64) (*BatchResult, error) {
	batchSize := w.batchSize
	if overrideBatchSize > 0 {
		batchSize = int(overrideBatchSize)
	}

	rows, err := w.svcCtx.Store.ClaimPending(ctx, batchSize)
	if err != nil {
		return nil, err
	}
	result := &BatchResult{
		Picked: int64(len(rows)),
	}
	if len(rows) == 0 {
		return result, nil
	}

	for _, row := range rows {
		err := w.handleEvent(ctx, row.EventType, row.ResourceID)
		if err == nil {
			if markErr := w.svcCtx.Store.MarkDone(ctx, row.ID); markErr != nil {
				logx.WithContext(ctx).Errorf("mark done failed, event_id=%d: %v", row.ID, markErr)
				result.Failed++
				continue
			}
			result.Succeeded++
			continue
		}

		if markErr := w.svcCtx.Store.MarkRetry(ctx, row.ID, row.Attempts, w.maxAttempts, w.retryBackoff, err); markErr != nil {
			logx.WithContext(ctx).Errorf("mark retry failed, event_id=%d: %v", row.ID, markErr)
			result.Failed++
			continue
		}
		result.Failed++
		logx.WithContext(ctx).Errorf("process outbox failed, event_id=%d event_type=%s content_id=%d attempts=%d err=%v", row.ID, row.EventType, row.ResourceID, row.Attempts+1, err)
	}
	return result, nil
}

func (w *IndexerWorker) handleEvent(ctx context.Context, eventType string, contentID int64) error {
	switch strings.TrimSpace(eventType) {
	case events.TopicContentCreated, events.TopicContentUpdated, events.TopicContentStatusChanged, events.TopicContentRevisionCreated:
		return w.upsertDocument(ctx, contentID)
	case events.TopicContentDeleted:
		_, err := w.svcCtx.Search.DeleteDocument(ctx, &searchrpc.DeleteDocumentRequest{ContentId: contentID})
		return err
	default:
		logx.WithContext(ctx).Infof("ignore unsupported outbox event, type=%s, content_id=%d", eventType, contentID)
		return nil
	}
}

func (w *IndexerWorker) upsertDocument(ctx context.Context, contentID int64) error {
	contentDetail, err := w.svcCtx.Content.GetContent(ctx, &contentrpc.GetContentRequest{Id: contentID})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			_, delErr := w.svcCtx.Search.DeleteDocument(ctx, &searchrpc.DeleteDocumentRequest{ContentId: contentID})
			return delErr
		}
		return err
	}

	_, err = w.svcCtx.Search.UpsertDocument(ctx, &searchrpc.UpsertDocumentRequest{
		ContentId:    contentDetail.Id,
		Type:         contentDetail.Type,
		Title:        contentDetail.Title,
		Slug:         contentDetail.Slug,
		Summary:      contentDetail.Summary,
		BodyMarkdown: contentDetail.BodyMarkdown,
		Status:       contentDetail.Status,
		Visibility:   contentDetail.Visibility,
		AiAccess:     contentDetail.AiAccess,
	})
	return err
}

func parseDurationOrDefault(raw string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(raw)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
