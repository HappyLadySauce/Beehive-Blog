package logic

import (
	"context"
	"strings"

	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLogic {
	return &QueryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryLogic) Query(in *pb.SearchRequest) (*pb.SearchResponse, error) {
	query := strings.TrimSpace(in.GetQuery())
	if query == "" {
		return &pb.SearchResponse{List: []*pb.SearchResultItem{}}, nil
	}

	// Temporary bootstrap behavior before indexer + search engine integration.
	return &pb.SearchResponse{
		List: []*pb.SearchResultItem{
			{
				ContentId:  0,
				Type:       "note",
				Title:      "Search bootstrap result",
				Slug:       "search-bootstrap-result",
				Summary:    "Search service is online and waiting for indexed content.",
				Highlight:  query,
				Score:      1,
			},
		},
	}, nil
}
