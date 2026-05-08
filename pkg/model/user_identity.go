package model

import (
	"time"

	"gorm.io/gorm"
)

// UserIdentity maps external provider subjects to local users.
// UserIdentity 将外部身份提供方主体映射到本地用户。
type UserIdentity struct {
	ID              int64          `gorm:"primaryKey;column:id"`
	UserID          int64          `gorm:"column:user_id;not null;index"`
	Provider        string         `gorm:"column:provider;size:32;not null;uniqueIndex:ux_user_identities_provider_subject,where:deleted_at IS NULL"`
	ProviderSubject string         `gorm:"column:provider_subject;size:128;not null;uniqueIndex:ux_user_identities_provider_subject,where:deleted_at IS NULL"`
	EmailAtBind     *string        `gorm:"column:email_at_bind;size:320"`
	CreatedAt       time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserIdentity) TableName() string {
	return "identity.user_identities"
}
