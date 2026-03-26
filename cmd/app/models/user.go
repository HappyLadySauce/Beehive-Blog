package models

import (
	"time"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 正常
	UserStatusInactive UserStatus = "inactive" // 未激活
	UserStatusDisabled UserStatus = "disabled" // 禁用
	UserStatusDeleted  UserStatus = "deleted"  // 已删除
)

// UserRole 用户角色
type UserRole string

const (
	UserRoleGuest UserRole = "guest" // 访客
	UserRoleUser  UserRole = "user"  // 普通用户
	UserRoleAdmin UserRole = "admin" // 管理员
)

// User 用户模型
type User struct {
	ID               int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username         string     `json:"username" gorm:"uniqueIndex;size:20;not null"`
	Nickname         string     `json:"nickname" gorm:"size:50"`
	Email            string     `json:"email" gorm:"uniqueIndex;size:100"`
	Password         string     `json:"-" gorm:"size:255;not null"`
	Avatar           string     `json:"avatar" gorm:"size:500"`
	Role             UserRole   `json:"role" gorm:"size:20;default:'user'"`
	Status           UserStatus `json:"status" gorm:"size:20;default:'inactive'"`
	Level            int        `json:"level" gorm:"default:1"`
	Experience       int        `json:"experience" gorm:"default:0"`
	CommentCount     int        `json:"commentCount" gorm:"default:0"`
	ArticleViewCount int        `json:"articleViewCount" gorm:"default:0"`
	LastLoginAt      *time.Time `json:"lastLoginAt"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// GetLevelName 获取等级名称
func (u *User) GetLevelName() string {
	levelNames := map[int]string{
		1: "初来乍到",
		2: "博客新手",
		3: "活跃读者",
		4: "资深访客",
		5: "博客达人",
		6: "资深博主",
	}
	if name, ok := levelNames[u.Level]; ok {
		return name
	}
	return "未知等级"
}

// UserLevel 用户等级配置模型
type UserLevel struct {
	ID               int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Level            int    `json:"level" gorm:"uniqueIndex;not null"`
	Name             string `json:"name" gorm:"size:50;not null"`
	RequiredExp      int    `json:"requiredExp" gorm:"not null"`
	RequiredDays     int    `json:"requiredDays"`
	RequiredArticles int    `json:"requiredArticles"`
	Description      string `json:"description" gorm:"size:255"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (UserLevel) TableName() string {
	return "user_levels"
}
