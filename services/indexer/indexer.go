package main

import (
	"context"
	"flag"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/config"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/server"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/svc"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/internal/worker"
	"github.com/HappyLadySauce/Beehive-Blog/services/indexer/pb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/indexer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)

	w := worker.NewIndexerWorker(svcCtx)
	logx.Infof("starting indexer worker, poll=%s, batch=%d", c.Worker.PollInterval, c.Worker.BatchSize)
	go w.Run(context.Background())

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterIndexerServer(grpcServer, server.NewIndexerServer(svcCtx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
