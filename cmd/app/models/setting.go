package models

import (
	"time"
)

// Setting 系统设置模型
type Setting struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Key       string    `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	Group     string    `json:"group" gorm:"size:50;default:'general'"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Setting) TableName() string {
	return "settings"
}

// SettingGroup 设置分组常量
const (
	SettingGroupGeneral  = "general"  // 常规设置
	SettingGroupSEO      = "seo"      // SEO设置
	SettingGroupSMTP     = "smtp"     // 邮件设置
	SettingGroupComment  = "comment"  // 评论设置
	SettingGroupSecurity = "security" // 安全设置
	SettingGroupHexo     = "hexo"     // Hexo 同步行为（路径仍来自 YAML hexo_dir）
)

// Link 友情链接模型
type Link struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	URL         string    `json:"url" gorm:"size:500;not null"`
	Description string    `json:"description" gorm:"size:255"`
	Logo        string    `json:"logo" gorm:"size:500"`
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`
	IsEnabled   bool      `json:"isEnabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Link) TableName() string {
	return "links"
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     int64     `json:"userId" gorm:"index"`
	Username   string    `json:"username" gorm:"size:50"`
	Action     string    `json:"action" gorm:"size:50;not null"`
	ObjectType string    `json:"objectType" gorm:"size:50"`
	ObjectID   string    `json:"objectId" gorm:"size:50"`
	Detail     string    `json:"detail" gorm:"type:text"`
	IP         string    `json:"ip" gorm:"size:50"`
	UserAgent  string    `json:"userAgent" gorm:"size:500"`
	Status     string    `json:"status" gorm:"size:20;default:'success'"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

// Backup 数据备份记录
type Backup struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	FilePath  string    `json:"filePath" gorm:"size:500;not null"`
	FileSize  int64     `json:"fileSize"`
	Type      string    `json:"type" gorm:"size:20;default:'manual'"` // manual/auto
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (Backup) TableName() string {
	return "backups"
}
