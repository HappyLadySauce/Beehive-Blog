// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	contentrpc "github.com/HappyLadySauce/Beehive-Blog/services/content/content"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	identityrpc "github.com/HappyLadySauce/Beehive-Blog/services/identity/identity"
	searchrpc "github.com/HappyLadySauce/Beehive-Blog/services/search/search"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config   config.Config
	Identity identityrpc.Identity
	Content  contentrpc.Content
	Search   searchrpc.Search
}

func NewServiceContext(c config.Config) *ServiceContext {
	identityClient := zrpc.MustNewClient(c.IdentityRpc)
	contentClient := zrpc.MustNewClient(c.ContentRpc)
	searchClient := zrpc.MustNewClient(c.SearchRpc)

	return &ServiceContext{
		Config:   c,
		Identity: identityrpc.NewIdentity(identityClient),
		Content:  contentrpc.NewContent(contentClient),
		Search:   searchrpc.NewSearch(searchClient),
	}
}
