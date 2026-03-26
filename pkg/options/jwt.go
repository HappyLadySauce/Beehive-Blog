package options

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

const (
	DefaultJWTSecret = "default_jwt_secret"
	DefaultExpireDuration = 24 * 7 * time.Hour
	DefaultRefreshTokenExpireDuration = 30 * 24 * time.Hour
)

type JWTOptions struct {
	JWTSecret string `json:"jwtSecret" mapstructure:"jwtSecret"`
	ExpireDuration time.Duration `json:"expireDuration" mapstructure:"expireDuration"`
	RefreshTokenExpireDuration time.Duration `json:"refreshTokenExpireDuration" mapstructure:"refreshTokenExpireDuration"`
}

func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		JWTSecret: DefaultJWTSecret,
		ExpireDuration: DefaultExpireDuration,
		RefreshTokenExpireDuration: DefaultRefreshTokenExpireDuration,
	}
}

func (i *JWTOptions) Validate() error {
	var errs []error
	if i.JWTSecret == "" {
		errs = append(errs, fmt.Errorf("jwtSecret is empty"))
	}
	if i.ExpireDuration <= 0 {
		errs = append(errs, fmt.Errorf("expireDuration is out of range"))
	}
	if i.RefreshTokenExpireDuration <= 0 {
		errs = append(errs, fmt.Errorf("refreshTokenExpireDuration is out of range"))
	}
	if i.ExpireDuration > i.RefreshTokenExpireDuration {
		errs = append(errs, fmt.Errorf("expireDuration must be less than refreshTokenExpireDuration"))
	}
	return errors.Join(errs...)
}

func (i *JWTOptions) AddFlags(fs *pflag.FlagSet) {	
	fs.StringVarP(&i.JWTSecret, "jwtSecret", "s", i.JWTSecret, "JWT secret")
	fs.DurationVarP(&i.ExpireDuration, "expireDuration", "e", i.ExpireDuration, "JWT expire duration in hours")
	fs.DurationVarP(&i.RefreshTokenExpireDuration, "refreshTokenExpireDuration", "r", i.RefreshTokenExpireDuration, "JWT token expire duration in hours")
}
