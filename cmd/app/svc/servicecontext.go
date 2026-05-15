package svc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/auth/jwt"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/options"
	pkgsettings "github.com/HappyLadySauce/Beehive-Blog/pkg/settings"
)

// ServiceContext wires shared infrastructure for HTTP handlers and background work.
// ServiceContext 为 HTTP 处理器与后台任务提供共享的基础设施连接。
type ServiceContext struct {
	Config      *config.Config
	DB          *gorm.DB
	Cache       *redis.Client
	Token       *jwt.Issuer
	PostgresDSN string

	SettingsStore    *pkgsettings.Store
	SettingsProvider *pkgsettings.Provider

	// DriverStore provides GORM-backed queries for storage drivers and mounts.
	// DriverStore 提供基于 GORM 的存储驱动与挂载项查询。
	DriverStore *driver.Store
	// DriverRegistry maps driver_name to DriverFactory for creating DriverBackend instances.
	// DriverRegistry 将 driver_name 映射到 DriverFactory，用于创建 DriverBackend 实例。
	DriverRegistry *driver.DriverRegistry
}

// NewServiceContext opens PostgreSQL (GORM) and Redis, applies pool settings, and verifies connectivity.
// NewServiceContext 打开 PostgreSQL（GORM）与 Redis，应用连接池参数并通过 Ping 校验连通性。
func NewServiceContext(ctx context.Context, cfg *config.Config) (*ServiceContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Database == nil {
		return nil, fmt.Errorf("database config is nil")
	}
	if cfg.Cache == nil {
		return nil, fmt.Errorf("cache config is nil")
	}
	if cfg.JWT == nil {
		return nil, fmt.Errorf("jwt config is nil")
	}
	if cfg.Email == nil {
		return nil, fmt.Errorf("email config is nil")
	}

	dsn, err := postgreDSN(cfg.Database)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	klog.InfoS("PostgreSQL connection established")

	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Cache.Host, strconv.Itoa(cfg.Cache.Port)),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	klog.InfoS("Redis connection established")

	issuer, err := jwt.NewIssuer(cfg.JWT)
	if err != nil {
		_ = rdb.Close()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("init token issuer: %w", err)
	}
	klog.InfoS("JWT issuer initialized")

	return &ServiceContext{
		Config:      cfg,
		DB:          db,
		Cache:       rdb,
		Token:       issuer,
		PostgresDSN: dsn,
	}, nil
}

// Close releases database and Redis resources (SQL first, then Redis).
// Close 释放数据库与 Redis 资源（先 SQL，后 Redis）。
func (s *ServiceContext) Close() error {
	var err error
	if s.DB != nil {
		sqlDB, e := s.DB.DB()
		if e != nil {
			err = errors.Join(err, e)
		} else {
			err = errors.Join(err, sqlDB.Close())
		}
	}
	if s.Cache != nil {
		err = errors.Join(err, s.Cache.Close())
	}
	return err
}

// postgreDSN builds a libpq-compatible URL DSN for the GORM postgres driver.
// postgreDSN 构造适用于 GORM postgres 驱动的 URL 形式 DSN，并对特殊字符做转义。
func postgreDSN(p *options.PostgreOptions) (string, error) {
	if p == nil {
		return "", fmt.Errorf("postgres options is nil")
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.User, p.Password),
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:   "/" + url.PathEscape(p.DB),
	}
	q := url.Values{}
	q.Set("sslmode", p.SSLMode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
