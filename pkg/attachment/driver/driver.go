package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	DriverLocal = "local"
	DriverS3    = "s3"
	DriverOSS   = "oss"
)

var (
	// ErrUnsupportedDriver is returned when a driver_name is not registered.
	// ErrUnsupportedDriver 在 driver_name 未注册时返回。
	ErrUnsupportedDriver = errors.New("storage driver is unsupported")

	// ErrDriverNotReady is returned when a driver fails health check or config validation.
	// ErrDriverNotReady 在驱动健康检查或配置校验失败时返回。
	ErrDriverNotReady = errors.New("storage driver is not ready")
)

// DriverBackend is the interface that all storage drivers implement.
// DriverBackend 是所有存储驱动必须实现的接口。
type DriverBackend interface {
	// DriverName returns the driver identifier (e.g. "local", "s3", "oss").
	// DriverName 返回驱动标识名（如 "local"、"s3"、"oss"）。
	DriverName() string

	Save(ctx context.Context, req PutRequest) (StoredObject, error)

	PresignUpload(ctx context.Context, req PresignRequest) (PresignResult, error)

	PresignDownload(ctx context.Context, objectKey string, ttl time.Duration) (PresignResult, error)

	// LocalFilePath resolves a relative path against the driver's root.
	// LocalFilePath 将相对路径解析为驱动根目录下的绝对路径。
	LocalFilePath(localPath string) (string, error)

	// Delete removes an object from storage. Best-effort; errors are logged, not fatal.
	// Delete 从存储中删除对象。尽力而为；错误仅记录，不会阻塞主流程。
	Delete(ctx context.Context, objectKey string) error

	// HealthCheck verifies the driver can access its backing storage.
	// HealthCheck 验证驱动是否可以访问其后端存储。
	HealthCheck(ctx context.Context) error
}

// DriverFactory creates a DriverBackend from a JSON config blob.
// DriverFactory 从 JSON 配置创建 DriverBackend。
type DriverFactory func(config json.RawMessage) (DriverBackend, error)

// DriverRegistry maps driver_name to its factory function.
// DriverRegistry 将 driver_name 映射到其工厂函数。
type DriverRegistry struct {
	factories map[string]DriverFactory
}

// NewDriverRegistry returns an empty DriverRegistry.
// NewDriverRegistry 返回空的 DriverRegistry。
func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{factories: make(map[string]DriverFactory)}
}

// Register adds a driver factory for the given name.
// Register 为给定名称注册驱动工厂。
func (r *DriverRegistry) Register(name string, factory DriverFactory) {
	r.factories[name] = factory
}

// CreateBackend instantiates a DriverBackend from driver_name and config.
// CreateBackend 根据 driver_name 和 config 实例化 DriverBackend。
func (r *DriverRegistry) CreateBackend(driverName string, config json.RawMessage) (DriverBackend, error) {
	if r == nil {
		return nil, fmt.Errorf("driver registry is nil")
	}
	factory, ok := r.factories[driverName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDriver, driverName)
	}
	return factory(config)
}

// Names returns the registered driver names.
// Names 返回已注册的驱动名称列表。
func (r *DriverRegistry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}
