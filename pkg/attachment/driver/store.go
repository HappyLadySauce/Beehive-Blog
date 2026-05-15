package driver

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

// Store provides GORM-backed queries for storage drivers and mounts.
// Store 提供基于 GORM 的存储驱动与挂载项查询方法。
type Store struct {
	db *gorm.DB
}

// NewStore creates a Store backed by the given GORM handle.
// NewStore 基于给定 GORM 句柄创建 Store。
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// ListDrivers returns all active storage drivers.
// ListDrivers 返回所有活跃的存储驱动。
func (s *Store) ListDrivers(ctx context.Context) ([]model.StorageDriver, error) {
	var rows []model.StorageDriver
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list storage drivers: %w", err)
	}
	return rows, nil
}

// GetDriver returns a driver by name, ignoring soft-deleted rows.
// GetDriver 按名称返回驱动，忽略已软删的行。
func (s *Store) GetDriver(ctx context.Context, name string) (*model.StorageDriver, error) {
	var row model.StorageDriver
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&row).Error; err != nil {
		return nil, fmt.Errorf("get storage driver %s: %w", name, err)
	}
	return &row, nil
}

// ListMounts returns all mounts, ignoring soft-deleted rows.
// ListMounts 返回所有挂载项，忽略已软删的行。
func (s *Store) ListMounts(ctx context.Context) ([]model.StorageMount, error) {
	var rows []model.StorageMount
	if err := s.db.WithContext(ctx).Order("order_index ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list storage mounts: %w", err)
	}
	return rows, nil
}

// GetMountByID returns a single mount, ignoring soft-deleted rows.
// GetMountByID 返回单个挂载项，忽略已软删的行。
func (s *Store) GetMountByID(ctx context.Context, id int64) (*model.StorageMount, error) {
	var row model.StorageMount
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, fmt.Errorf("get storage mount %d: %w", id, err)
	}
	return &row, nil
}

// GetDefaultEnabledMount returns the default mount that can accept writes.
// GetDefaultEnabledMount 返回可写入的默认挂载项。
func (s *Store) GetDefaultEnabledMount(ctx context.Context) (*model.StorageMount, error) {
	var row model.StorageMount
	if err := s.db.WithContext(ctx).
		Where("is_default = ? AND disabled = ? AND status <> ?", true, false, "error").
		First(&row).Error; err != nil {
		return nil, fmt.Errorf("default storage mount: %w", err)
	}
	return &row, nil
}

// CountMounts returns the number of non-deleted storage mounts.
// CountMounts 返回未软删的存储挂载项数量。
func (s *Store) CountMounts(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.StorageMount{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count storage mounts: %w", err)
	}
	return count, nil
}

// CreateMount inserts a new mount row.
// CreateMount 插入新的挂载项。
func (s *Store) CreateMount(ctx context.Context, mount *model.StorageMount) error {
	if err := s.db.WithContext(ctx).Create(mount).Error; err != nil {
		return fmt.Errorf("create storage mount: %w", err)
	}
	return nil
}

// UpdateMount patches specified columns on an existing mount.
// UpdateMount 更新已有挂载项的指定列。
func (s *Store) UpdateMount(ctx context.Context, id int64, updates map[string]interface{}) error {
	res := s.db.WithContext(ctx).Model(&model.StorageMount{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update storage mount %d: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("storage mount %d: %w", id, gorm.ErrRecordNotFound)
	}
	return nil
}

// SoftDeleteMount sets deleted_at on a mount.
// SoftDeleteMount 对挂载项设置 deleted_at。
func (s *Store) SoftDeleteMount(ctx context.Context, id int64) error {
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.StorageMount{})
	if res.Error != nil {
		return fmt.Errorf("delete storage mount %d: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("storage mount %d: %w", id, gorm.ErrRecordNotFound)
	}
	return nil
}

// CountAttachmentsOnMount returns the number of non-deleted attachments
// referencing the given mount.
// CountAttachmentsOnMount 返回引用指定挂载项且未软删的附件数量。
func (s *Store) CountAttachmentsOnMount(ctx context.Context, mountID int64) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Attachment{}).
		Where("storage_mount_id = ?", mountID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count attachments on mount %d: %w", mountID, err)
	}
	return count, nil
}
