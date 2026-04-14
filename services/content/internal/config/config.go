package config

import "github.com/zeromicro/go-zero/core/stores/sqlx"
import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DB sqlx.SqlConf
}
