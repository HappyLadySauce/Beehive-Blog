package model

import (
	"time"

	"gorm.io/gorm"
)

// User maps to the identity.users table.
// User 映射到 identity.users 表。
type User struct {
	ID                 int64          `gorm:"primaryKey;column:id"`
	Username           string         `gorm:"column:username;size:64;not null"`
	Email              *string        `gorm:"column:email;size:320"`
	Nickname           *string        `gorm:"column:nickname;size:128"`
	Phone              *string        `gorm:"column:phone;size:16"`
	AvatarAttachmentID *int64         `gorm:"column:avatar_attachment_id"`
	Role               string         `gorm:"column:role;size:16;not null;default:member"`
	Status             string         `gorm:"column:status;size:16;not null;default:active"`
	LastLoginAt        *time.Time     `gorm:"column:last_login_at"`
	CreatedAt          time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (User) TableName() string {
	return "identity.users"
}
