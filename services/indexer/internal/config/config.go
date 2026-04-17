package config

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type WorkerConf struct {
	PollInterval string `json:",default=2s"`
	BatchSize    int    `json:",default=20"`
	MaxAttempts  int    `json:",default=8"`
	RetryBackoff string `json:",default=5s"`
}

type Config struct {
	zrpc.RpcServerConf
	DB         sqlx.SqlConf
	ContentRpc zrpc.RpcClientConf
	SearchRpc  zrpc.RpcClientConf
	Worker     WorkerConf
}
