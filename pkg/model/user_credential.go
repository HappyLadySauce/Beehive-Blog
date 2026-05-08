package model

import (
	"time"

	"gorm.io/gorm"
)

// UserCredential maps to identity.user_credentials; holds bcrypt password hash.
// UserCredential 映射到 identity.user_credentials；存储 bcrypt 密码哈希。
type UserCredential struct {
	ID           int64          `gorm:"primaryKey;column:id"`
	UserID       int64          `gorm:"column:user_id;not null;uniqueIndex:ux_user_credentials_user_id,where:deleted_at IS NULL"`
	PasswordHash string         `gorm:"column:password_hash;size:255;not null;check:password_hash <> ''"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserCredential) TableName() string {
	return "identity.user_credentials"
}
