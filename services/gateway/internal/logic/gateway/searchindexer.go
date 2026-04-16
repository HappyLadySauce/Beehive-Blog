package gateway

import (
	"context"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
	"github.com/zeromicro/go-zero/core/logx"
)

func buildUpsertDocumentRequest(detail *contentrpc.ContentDetail) *searchrpc.UpsertDocumentRequest {
	if detail == nil {
		return nil
	}
	return &searchrpc.UpsertDocumentRequest{
		ContentId:    detail.Id,
		Type:         detail.Type,
		Title:        detail.Title,
		Slug:         detail.Slug,
		Summary:      detail.Summary,
		BodyMarkdown: detail.BodyMarkdown,
		Status:       detail.Status,
		Visibility:   detail.Visibility,
		AiAccess:     detail.AiAccess,
	}
}

func triggerAsyncIndexUpsert(ctx context.Context, searchCli searchrpc.Search, detail *contentrpc.ContentDetail) {
	req := buildUpsertDocumentRequest(detail)
	if req == nil {
		return
	}

	go func() {
		callCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := searchCli.UpsertDocument(callCtx, req); err != nil {
			logx.WithContext(ctx).Errorf("search index upsert failed, content_id=%d, err=%v", req.ContentId, err)
		}
	}()
}
