package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StorageDriver maps to attachment.storage_drivers.
// StorageDriver 映射到 attachment.storage_drivers 表。
type StorageDriver struct {
	ID           int64           `gorm:"primaryKey;column:id"`
	Name         string          `gorm:"column:name;size:64;not null"`
	DisplayName  string          `gorm:"column:display_name;size:128;not null"`
	Description  *string         `gorm:"column:description"`
	ConfigSchema json.RawMessage `gorm:"column:config_schema;type:jsonb;not null;default:'{}'"`
	Capabilities json.RawMessage `gorm:"column:capabilities;type:jsonb;not null;default:'{}'"`
	Status       string          `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt    time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

func (StorageDriver) TableName() string {
	return "attachment.storage_drivers"
}
