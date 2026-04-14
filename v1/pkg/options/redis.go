package options

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

const (
	DefaultRedisHost                  = "localhost:6379"
	DefaultRedisPass                  = ""
	DefaultRedisDB                    = 0
	DefaultRedisDialTimeoutSeconds    = 5
	DefaultRedisReadTimeoutSeconds    = 3
	DefaultRedisWriteTimeoutSeconds   = 3
	DefaultRedisConnectTimeoutSeconds = 3
	DefaultRedisEnableTLS             = false
	DefaultRedisInsecureSkipVerify    = false
)

type RedisOptions struct {
	RedisHost             string `json:"redisHost" mapstructure:"redisHost"`
	RedisPass             string `json:"redisPass" mapstructure:"redisPass"`
	RedisDB               int    `json:"redisDB" mapstructure:"redisDB"`
	DialTimeoutSeconds    int    `json:"dialTimeoutSeconds" mapstructure:"dialTimeoutSeconds"`
	ReadTimeoutSeconds    int    `json:"readTimeoutSeconds" mapstructure:"readTimeoutSeconds"`
	WriteTimeoutSeconds   int    `json:"writeTimeoutSeconds" mapstructure:"writeTimeoutSeconds"`
	ConnectTimeoutSeconds int    `json:"connectTimeoutSeconds" mapstructure:"connectTimeoutSeconds"`
	EnableTLS             bool   `json:"enableTLS" mapstructure:"enableTLS"`
	InsecureSkipVerify    bool   `json:"insecureSkipVerify" mapstructure:"insecureSkipVerify"`
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		RedisHost:             DefaultRedisHost,
		RedisPass:             DefaultRedisPass,
		RedisDB:               DefaultRedisDB,
		DialTimeoutSeconds:    DefaultRedisDialTimeoutSeconds,
		ReadTimeoutSeconds:    DefaultRedisReadTimeoutSeconds,
		WriteTimeoutSeconds:   DefaultRedisWriteTimeoutSeconds,
		ConnectTimeoutSeconds: DefaultRedisConnectTimeoutSeconds,
		EnableTLS:             DefaultRedisEnableTLS,
		InsecureSkipVerify:    DefaultRedisInsecureSkipVerify,
	}
}

func (o *RedisOptions) Validate() error {
	var errs []error
	if o.RedisHost == "" {
		errs = append(errs, fmt.Errorf("redis host is empty"))
	}
	if o.RedisDB < 0 {
		errs = append(errs, fmt.Errorf("redis db is negative"))
	}
	if o.DialTimeoutSeconds <= 0 {
		errs = append(errs, fmt.Errorf("dialTimeoutSeconds must be greater than 0"))
	}
	if o.ReadTimeoutSeconds <= 0 {
		errs = append(errs, fmt.Errorf("readTimeoutSeconds must be greater than 0"))
	}
	if o.WriteTimeoutSeconds <= 0 {
		errs = append(errs, fmt.Errorf("writeTimeoutSeconds must be greater than 0"))
	}
	if o.ConnectTimeoutSeconds <= 0 {
		errs = append(errs, fmt.Errorf("connectTimeoutSeconds must be greater than 0"))
	}
	return errors.Join(errs...)
}

func (o *RedisOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.RedisHost, "redisHost", o.RedisHost, "redis host")
	fs.StringVar(&o.RedisPass, "redisPass", o.RedisPass, "redis password")
	fs.IntVar(&o.RedisDB, "redisDB", o.RedisDB, "redis db")
	fs.IntVar(&o.DialTimeoutSeconds, "redisDialTimeoutSeconds", o.DialTimeoutSeconds, "redis dial timeout in seconds")
	fs.IntVar(&o.ReadTimeoutSeconds, "redisReadTimeoutSeconds", o.ReadTimeoutSeconds, "redis read timeout in seconds")
	fs.IntVar(&o.WriteTimeoutSeconds, "redisWriteTimeoutSeconds", o.WriteTimeoutSeconds, "redis write timeout in seconds")
	fs.IntVar(&o.ConnectTimeoutSeconds, "redisConnectTimeoutSeconds", o.ConnectTimeoutSeconds, "redis connect check timeout in seconds")
	fs.BoolVar(&o.EnableTLS, "redisEnableTLS", o.EnableTLS, "enable TLS for redis connection")
	fs.BoolVar(&o.InsecureSkipVerify, "redisInsecureSkipVerify", o.InsecureSkipVerify, "skip redis TLS cert verification (unsafe, debug only)")
}
