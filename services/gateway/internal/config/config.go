// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	zgateway "github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/zrpc"
)

type AuthConf struct {
	AccessSecret string `json:",optional"`
}

type RateLimitConf struct {
	Enabled bool    `json:",default=true"`
	RPS     float64 `json:",default=50"`
	Burst   int     `json:",default=100"`
}

type AccessLogConf struct {
	SlowRequestWarnThresholdMs int64 `json:",default=500"`
}

type Config struct {
	zgateway.GatewayConf
	UpstreamFiles []string `json:",optional"`
	IdentityRpc   zrpc.RpcClientConf
	ContentRpc    zrpc.RpcClientConf
	SearchRpc     zrpc.RpcClientConf
	Auth          AuthConf
	RateLimit     RateLimitConf
	AccessLog     AccessLogConf
}
