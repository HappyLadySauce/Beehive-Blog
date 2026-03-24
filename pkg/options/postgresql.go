package options

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

const (
	DefaultDatabaseHost      = "localhost"
	DefaultDatabasePort      = 5432
	DefaultDatabaseUsername  = "Beehive-Blog"
	DefaultDatabasePassword  = "Beehive-Blog"
	DefaultDatabaseDatabase  = "Beehive-Blog"
	DefaultDatabaseCharset   = "utf8mb4"
	DefaultDatabaseTimeZone  = "Asia/Shanghai"
	DefaultDatabaseParseTime = true
	DefaultDatabaseLoc       = "Local"
)

type DatabaseOptions struct {
	Host      string `json:"host" mapstructure:"host"`
	Port      int    `json:"port" mapstructure:"port"`
	Username  string `json:"username" mapstructure:"username"`
	Password  string `json:"password" mapstructure:"password"`
	Database  string `json:"database" mapstructure:"database"`
	Charset   string `json:"charset" mapstructure:"charset"`
	TimeZone  string `json:"timeZone" mapstructure:"timeZone"`
	ParseTime bool   `json:"parseTime" mapstructure:"parseTime"`
	Loc       string `json:"loc" mapstructure:"loc"`
}

func NewDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		Host:      DefaultDatabaseHost,
		Port:      DefaultDatabasePort,
		Username:  DefaultDatabaseUsername,
		Password:  DefaultDatabasePassword,
		Database:  DefaultDatabaseDatabase,
		Charset:   DefaultDatabaseCharset,
		TimeZone:  DefaultDatabaseTimeZone,
		ParseTime: DefaultDatabaseParseTime,
		Loc:       DefaultDatabaseLoc,
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
	if o.Charset == "" {
		errs = append(errs, fmt.Errorf("charset is empty"))
	}
	if o.Loc == "" {
		errs = append(errs, fmt.Errorf("loc is empty"))
	}
	return errors.Join(errs...)
}

func (o *DatabaseOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "host", o.Host, "database host")
	fs.IntVar(&o.Port, "port", o.Port, "database port")
	fs.StringVar(&o.Username, "username", o.Username, "database username")
	fs.StringVar(&o.Password, "password", o.Password, "database password")
	fs.StringVar(&o.Database, "database", o.Database, "database name")
	fs.StringVar(&o.Charset, "charset", o.Charset, "database charset")
	fs.StringVar(&o.TimeZone, "timeZone", o.TimeZone, "database timeZone")
	fs.BoolVar(&o.ParseTime, "parseTime", o.ParseTime, "database parseTime")
	fs.StringVar(&o.Loc, "loc", o.Loc, "database loc")
}
