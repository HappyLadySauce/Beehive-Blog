package driver

import (
	"context"
	"fmt"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// ResolveMountForWrite resolves a storage mount for new uploads.
// ResolveMountForWrite 解析用于新上传的存储挂载项。
func ResolveMountForWrite(
	ctx context.Context,
	store *Store,
	registry *DriverRegistry,
	mountID *int64,
) (*model.StorageMount, DriverBackend, error) {
	var mount *model.StorageMount
	var err error
	if mountID != nil {
		mount, err = store.GetMountByID(ctx, *mountID)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve storage mount: %w", err)
		}
	} else {
		mount, err = store.GetDefaultEnabledMount(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("default storage mount: %w", err)
		}
	}

	if mount.DeletedAt.Valid {
		return nil, nil, fmt.Errorf("storage mount %d is deleted", mount.ID)
	}
	if mount.Disabled {
		return nil, nil, fmt.Errorf("storage mount %d is disabled", mount.ID)
	}
	if mount.Status == "error" {
		return nil, nil, fmt.Errorf("storage mount %d is not ready", mount.ID)
	}

	backend, err := registry.CreateBackend(mount.DriverName, mount.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("create backend for %s: %w", mount.DriverName, err)
	}
	return mount, backend, nil
}

// ResolveMountForRead resolves a storage mount for existing file reads/deletes.
// ResolveMountForRead 解析用于已有文件读取或删除的存储挂载项。
func ResolveMountForRead(
	ctx context.Context,
	store *Store,
	registry *DriverRegistry,
	mountID int64,
) (*model.StorageMount, DriverBackend, error) {
	mount, err := store.GetMountByID(ctx, mountID)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve storage mount: %w", err)
	}
	if mount.DeletedAt.Valid {
		return nil, nil, fmt.Errorf("storage mount %d is deleted", mount.ID)
	}
	backend, err := registry.CreateBackend(mount.DriverName, mount.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("create backend for %s: %w", mount.DriverName, err)
	}
	return mount, backend, nil
}
