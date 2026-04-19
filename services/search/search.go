package main

import (
	"flag"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	
	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/config"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/server"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/search/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/search.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSearchServer(grpcServer, server.NewSearchServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
