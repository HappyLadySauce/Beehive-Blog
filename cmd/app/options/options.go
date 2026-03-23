package options


import (
	"encoding/json"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

type Options struct {
	Server          *options.ServerOptions          `mapstructure:"server"`
}

func NewOptions() *Options {
	return &Options{
		Server: options.NewServerOptions(),
	}
}

// AddFlags adds the flags to the specified FlagSet and returns the grouped flag sets.
func (o *Options) AddFlags(fs *pflag.FlagSet) *flag.NamedFlagSets {
	nfs := &flag.NamedFlagSets{}

	// add the flags to the NamedFlagSets
	configFS := nfs.FlagSet("Config")
	options.AddConfigFlag(configFS)

	serverFS := nfs.FlagSet("Server")
	o.Server.AddFlags(serverFS)

	// add the flags to the main Command
	for _, name := range nfs.Order {
		fs.AddFlagSet(nfs.FlagSets[name])
	}
	return nfs
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}