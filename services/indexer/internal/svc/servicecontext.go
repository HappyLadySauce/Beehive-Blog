package svc

import (
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/config"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	Content contentrpc.Content
	Search  searchrpc.Search
	Store   *outboxStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.MustNewConn(c.DB)
	contentClient := zrpc.MustNewClient(c.ContentRpc)
	searchClient := zrpc.MustNewClient(c.SearchRpc)

	return &ServiceContext{
		Config:  c,
		Content: contentrpc.NewContent(contentClient),
		Search:  searchrpc.NewSearch(searchClient),
		Store:   newOutboxStore(conn),
	}
}
