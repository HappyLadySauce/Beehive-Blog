package svc

import (
	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Store  *searchStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.MustNewConn(c.DB)
	store, err := newSearchStore(conn)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		Store:  store,
	}
}
