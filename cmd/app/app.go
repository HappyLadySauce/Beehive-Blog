package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/options"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
)

func NewAPICommand(ctx context.Context, basename string) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:   basename,
		Short: "Beehive-Blog is a web server for Beehive Blog",
		Long:  "Beehive-Blog is a web server for Beehive Blog",
		RunE: func(cmd *cobra.Command, args []string) error {
			// bind command line flags to viper (command line args override config file)
			// 从命令行标志中绑定到 viperiper
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("failed to bind command line flags: %w", err)
			}

			// unmarshal viper config to options struct
			// 从 viperiper 解析配置到选项结构体
			if err := viper.Unmarshal(opts); err != nil {
				return fmt.Errorf("failed to unmarshal viper config: %w", err)
			}

			// validate options after flags & config are fully populated
			// 验证选项
			if err := opts.Validate(); err != nil {
				klog.ErrorS(err, "Configuration validation failed")
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// create service context
			// 创建服务上下文
			serviceCtx, err := svc.NewServiceContext(*opts)
			if err != nil {
				klog.ErrorS(err, "Failed to create service context")
				return err
			}

			// ensure service context is closed on exit
			// 确保服务上下文在退出时关闭
			defer func() {
				if err := serviceCtx.Close(); err != nil {
					klog.ErrorS(err, "Failed to close service context")
				}
			}()

			return run(ctx, serviceCtx)
		},
	}

	// Add command line flags
	// 添加命令行标志
	nfs := opts.AddFlags(cmd.Flags(), basename)
	flag.SetUsageAndHelpFunc(cmd, *nfs, 80)

	return cmd
}

func run(ctx context.Context, serviceCtx *svc.ServiceContext) error {
	// 创建 HTTP 服务器
	srv := serve(serviceCtx)

	// 等待 context 取消信号（优雅关闭）
	<-ctx.Done()
	klog.InfoS("Shutting down server...")

	// 使用超时 context 进行优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		klog.ErrorS(err, "Server forced to shutdown")
		return err
	}

	klog.InfoS("Server exited gracefully")
	return nil
}

func serve(svcCtx *svc.ServiceContext) *http.Server {
	address := fmt.Sprintf("%s:%d", svcCtx.Config.ServerOptions.BindAddress, svcCtx.Config.ServerOptions.BindPort)
	klog.InfoS("Listening and serving on", "address", address)

	// 创建 gin 引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 注册路由
	router.GET("/livez", func(c *gin.Context) {
		c.String(http.StatusOK, "livez")
	})
	router.GET("/readyz", func(c *gin.Context) {
		c.String(http.StatusOK, "readyz")
	})

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// 在 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatal(err)
		}
	}()

	return srv
}
