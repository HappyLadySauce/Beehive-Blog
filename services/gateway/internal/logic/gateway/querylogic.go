// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package gateway

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/types"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLogic {
	return &QueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryLogic) Query(req *types.SearchRequest) (resp *types.SearchResponse, err error) {
	if req == nil || strings.TrimSpace(req.Query) == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	out, err := l.svcCtx.Search.Query(l.ctx, &searchrpc.SearchRequest{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     req.ContentType,
	})
	if err != nil {
		return nil, err
	}
	return toSearchResponse(out), nil
}
