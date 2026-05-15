package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StorageMount maps to attachment.storage_mounts.
// StorageMount 映射到 attachment.storage_mounts 表。
type StorageMount struct {
	ID            int64           `gorm:"primaryKey;column:id"`
	DriverName    string          `gorm:"column:driver_name;size:64;not null"`
	MountPath     string          `gorm:"column:mount_path;size:512;not null"`
	Name          string          `gorm:"column:name;size:128;not null"`
	Remark        *string         `gorm:"column:remark"`
	Config        json.RawMessage `gorm:"column:config;type:jsonb;not null;default:'{}'"`
	OrderIndex    int             `gorm:"column:order_index;not null;default:0"`
	IsDefault     bool            `gorm:"column:is_default;not null;default:false"`
	Disabled      bool            `gorm:"column:disabled;not null;default:false"`
	Status        string          `gorm:"column:status;size:16;not null;default:unknown"`
	LastCheckedAt *time.Time      `gorm:"column:last_checked_at"`
	LastError     *string         `gorm:"column:last_error"`
	CreatedBy     *int64          `gorm:"column:created_by"`
	CreatedAt     time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;not null"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

func (StorageMount) TableName() string {
	return "attachment.storage_mounts"
}
