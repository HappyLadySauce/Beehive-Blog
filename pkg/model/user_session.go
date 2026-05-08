package model

import "time"

// UserSession maps to identity.user_sessions and stores revocable refresh state.
// UserSession 映射到 identity.user_sessions，保存可撤销的 refresh 状态。
type UserSession struct {
	ID               int64      `gorm:"primaryKey;column:id"`
	UserID           int64      `gorm:"column:user_id;not null;index"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash;size:64;not null"`
	RefreshJTI       string     `gorm:"column:refresh_jti;size:64;not null;uniqueIndex"`
	ExpiresAt        time.Time  `gorm:"column:expires_at;not null;index"`
	RevokedAt        *time.Time `gorm:"column:revoked_at;index"`
	RevokedReason    *string    `gorm:"column:revoked_reason;size:64"`
	RotatedAt        *time.Time `gorm:"column:rotated_at;index"`
	CreatedIP        string     `gorm:"column:created_ip;size:64"`
	UserAgent        string     `gorm:"column:user_agent;size:512"`
	LastUsedAt       *time.Time `gorm:"column:last_used_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
}

func (UserSession) TableName() string {
	return "identity.user_sessions"
}
