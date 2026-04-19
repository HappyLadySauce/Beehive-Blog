package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	zgateway "github.com/zeromicro/go-zero/gateway"

	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/config"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/handler"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/middleware"
	"github.com/HappyLadySauce/Beehive-Blog/services/gateway/internal/svc"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	upstreams, err := config.LoadUpstreams(filepath.Dir(*configFile), c.UpstreamFiles)
	if err != nil {
		panic(err)
	}
	c.GatewayConf.Upstreams = append(c.GatewayConf.Upstreams, upstreams...)

	ctx := svc.NewServiceContext(c)
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(c.RateLimit)
	accessLogMiddleware := middleware.NewAccessLogMiddleware(c.AccessLog)

	server := zgateway.MustNewServer(
		c.GatewayConf,
		zgateway.WithMiddleware(requestIDMiddleware.Handle, rateLimitMiddleware.Handle, accessLogMiddleware.Handle),
	)
	defer server.Stop()

	handler.RegisterHandlers(server.Server, ctx)

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
