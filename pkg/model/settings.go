package model

import (
	"time"

	"gorm.io/gorm"
)

// ApplicationSetting maps the singleton row setting.application_settings.
// ApplicationSetting 映射单行表 setting.application_settings。
type ApplicationSetting struct {
	ID        int16          `gorm:"primaryKey;column:id"`
	Revision  int64          `gorm:"column:revision;not null"`
	Payload   []byte         `gorm:"column:payload;type:jsonb;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName returns the qualified table name for GORM.
// TableName 返回 GORM 使用的带 schema 表名。
func (ApplicationSetting) TableName() string {
	return "setting.application_settings"
}
