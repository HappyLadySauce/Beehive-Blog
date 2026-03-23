package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

type ServerOptions struct {
	BindAddress string `json:"bind-address" mapstructure:"bind-address"`
	BindPort    int    `json:"bind-port"    mapstructure:"bind-port"`
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}


func (i *ServerOptions) Validate() []error {
	var errors []error
	if i.BindAddress == "" {
		errors = append(errors, fmt.Errorf("bind-address is required"))
	}
	if i.BindPort == 0 {
		errors = append(errors, fmt.Errorf("bind-port is required"))
	}
	return errors
}	

func (i *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&i.BindAddress, "bind-address", "b", "127.0.0.1", "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	fs.IntVarP(&i.BindPort, "bind-port", "p", 8081, "port to listen to for incoming requests")
}