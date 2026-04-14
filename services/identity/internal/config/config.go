package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type AuthConf struct {
	AccessSecret string        `json:",optional"`
	Issuer       string        `json:",default=beehive-blog"`
	AccessTTL    time.Duration `json:",default=2h"`
	RefreshTTL   time.Duration `json:",default=720h"`
}

type Config struct {
	zrpc.RpcServerConf
	DB    sqlx.SqlConf
	Redis redis.RedisConf
	Auth  AuthConf
}
