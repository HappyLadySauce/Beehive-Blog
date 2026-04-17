package svc

import (
	"github.com/HappyLadySauce/Beehive-Blog/services/identity/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Store  *identityStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.MustNewConn(c.DB)
	redisClient := c.Redis.NewRedis()
	store, err := newIdentityStore(conn, redisClient, c.IdentityAuth)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		Store:  store,
	}
}
