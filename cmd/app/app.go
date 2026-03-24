package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/options"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
)

func NewAPICommand(ctx context.Context, basename string) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:   basename,
		Short: "Beehive-Blog is a web server for Beehive Blog",
		Long:  "Beehive-Blog is a web server for Beehive Blog",
		RunE: func(cmd *cobra.Command, args []string) error {
			// bind command line flags to viper (command line args override config file)
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if err := viper.Unmarshal(opts); err != nil {
				return err
			}

			// validate options after flags & config are fully populated
			if errs := opts.Validate(); len(errs) != 0 {
				for _, err := range errs {
					fmt.Fprintln(os.Stderr, "Error:", err)
				}
				os.Exit(1)
			}
			return run(ctx, opts)
		},
	}

	nfs := opts.AddFlags(cmd.Flags(), basename)
	flag.SetUsageAndHelpFunc(cmd, *nfs, 80)
	
	return cmd
}

func run(ctx context.Context, opts *options.Options) error {
	serve(opts)

	<-ctx.Done()
	os.Exit(0)
	return nil
}


func serve(opt *options.Options) {
	address := fmt.Sprintf("%s:%d", opt.Server.BindAddress, opt.Server.BindPort)
	klog.V(1).InfoS("Listening and serving on", "address", address)
	go func() {
		klog.Fatal(router.Router().Run(address))
	}()
}