package options

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

const (
	DefaultDatabaseHost                  = "localhost"
	DefaultDatabasePort                  = 5432
	DefaultDatabaseUsername              = "Beehive-Blog"
	DefaultDatabasePassword              = "Beehive-Blog"
	DefaultDatabaseDatabase              = "Beehive-Blog"
	DefaultDatabaseTimeZone              = "Asia/Shanghai"
	DefaultDatabaseSSLMode               = "disable"
	DefaultDatabaseConnectTimeoutSeconds = 5
	DefaultDatabaseAutoMigrate           = true
)

type DatabaseOptions struct {
	Host                  string `json:"host" mapstructure:"host"`
	Port                  int    `json:"port" mapstructure:"port"`
	Username              string `json:"username" mapstructure:"username"`
	Password              string `json:"password" mapstructure:"password"`
	Database              string `json:"database" mapstructure:"database"`
	TimeZone              string `json:"timeZone" mapstructure:"timeZone"`
	SSLMode               string `json:"sslMode" mapstructure:"sslMode"`
	ConnectTimeoutSeconds int    `json:"connectTimeoutSeconds" mapstructure:"connectTimeoutSeconds"`
	AutoMigrate           bool   `json:"autoMigrate" mapstructure:"autoMigrate"`
}

func NewDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		Host:                  DefaultDatabaseHost,
		Port:                  DefaultDatabasePort,
		Username:              DefaultDatabaseUsername,
		Password:              DefaultDatabasePassword,
		Database:              DefaultDatabaseDatabase,
		TimeZone:              DefaultDatabaseTimeZone,
		SSLMode:               DefaultDatabaseSSLMode,
		ConnectTimeoutSeconds: DefaultDatabaseConnectTimeoutSeconds,
		AutoMigrate:           DefaultDatabaseAutoMigrate,
	}
}

func (o *DatabaseOptions) Validate() error {
	var errs []error
	if o.Host == "" {
		errs = append(errs, fmt.Errorf("host is empty"))
	}
	if o.Port <= 0 {
		errs = append(errs, fmt.Errorf("port is negative"))
	}
	if o.Username == "" {
		errs = append(errs, fmt.Errorf("username is empty"))
	}
	if o.Password == "" {
		errs = append(errs, fmt.Errorf("password is empty"))
	}
	if o.Database == "" {
		errs = append(errs, fmt.Errorf("database is empty"))
	}
	if o.TimeZone == "" {
		errs = append(errs, fmt.Errorf("timeZone is empty"))
	}
	if o.SSLMode == "" {
		errs = append(errs, fmt.Errorf("sslMode is empty"))
	}
	if o.ConnectTimeoutSeconds <= 0 {
		errs = append(errs, fmt.Errorf("connectTimeoutSeconds must be greater than 0"))
	}
	return errors.Join(errs...)
}

func (o *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "host", o.Host, "database host")
	fs.IntVar(&o.Port, "port", o.Port, "database port")
	fs.StringVar(&o.Username, "username", o.Username, "database username")
	fs.StringVar(&o.Password, "password", o.Password, "database password")
	fs.StringVar(&o.Database, "database", o.Database, "database name")
	fs.StringVar(&o.TimeZone, "timeZone", o.TimeZone, "database timeZone")
	fs.StringVar(&o.SSLMode, "sslMode", o.SSLMode, "database TLS mode, e.g. disable/require/verify-ca/verify-full")
	fs.IntVar(&o.ConnectTimeoutSeconds, "connectTimeoutSeconds", o.ConnectTimeoutSeconds, "database connect timeout in seconds")
	fs.BoolVar(&o.AutoMigrate, "autoMigrate", o.AutoMigrate, "automatically run database schema migration at startup")
}
