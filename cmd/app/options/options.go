package options

import (
	"encoding/json"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
)

type Options struct {
	basename string
	InsecureServing *options.InsecureServingOptions `mapstructure:"insecure"`
}

func NewOptions(basename string) *Options {
	return &Options{
		basename: basename,
		InsecureServing: options.NewInsecureServingOptions(),
	}
}

// AddFlags adds the flags to the specified FlagSet and returns the grouped flag sets.
func (o *Options) AddFlags(fs *pflag.FlagSet) *flag.NamedFlagSets {
	nfs := &flag.NamedFlagSets{}

	// add the flags to the NamedFlagSets
	configFS := nfs.FlagSet("Config")
	options.AddConfigFlag(configFS, o.basename)

	insecureServingFS := nfs.FlagSet("Insecure Serving")
	o.InsecureServing.AddFlags(insecureServingFS)

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
