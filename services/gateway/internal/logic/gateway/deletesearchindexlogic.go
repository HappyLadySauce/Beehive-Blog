package gateway

import (
	"context"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteSearchIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSearchIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSearchIndexLogic {
	return &DeleteSearchIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSearchIndexLogic) DeleteSearchIndex(req *types.SearchIndexPathRequest) (err error) {
	if req == nil || req.Id <= 0 {
		return status.Error(codes.InvalidArgument, "id is required")
	}

	userID, _ := parseAccessTokenUserIDFromContext(l.ctx)
	fields := map[string]any{
		"user_id":    userID,
		"content_id": req.Id,
	}
	startedAt := time.Now()
	defer func() {
		auditRPC(l.ctx, "search.delete_index", "search.DeleteDocument", startedAt, err, fields)
	}()

	_, err = l.svcCtx.Search.DeleteDocument(l.ctx, &searchrpc.DeleteDocumentRequest{ContentId: req.Id})
	return err
}
