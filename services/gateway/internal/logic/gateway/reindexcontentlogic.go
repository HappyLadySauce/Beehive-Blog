package gateway

import (
	"context"
	"time"

	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReindexContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReindexContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReindexContentLogic {
	return &ReindexContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReindexContentLogic) ReindexContent(req *types.SearchIndexPathRequest) (resp *types.SearchIndexDocument, err error) {
	if req == nil || req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":    userID,
		"content_id": req.Id,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "search.reindex_content", "search.UpsertDocument", startedAt, err, fields)
	}()

	contentDetail, err := l.svcCtx.Content.GetContent(l.ctx, &contentrpc.GetContentRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	out, err := l.svcCtx.Search.UpsertDocument(l.ctx, buildUpsertDocumentRequest(contentDetail))
	if err != nil {
		return nil, err
	}
	return toSearchIndexDocument(out), nil
}
