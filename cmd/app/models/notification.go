package models

import (
	"time"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem  NotificationType = "system"  // 系统公告
	NotificationTypeComment NotificationType = "comment" // 评论回复
	NotificationTypeArticle NotificationType = "article" // 文章相关
	NotificationTypeUser    NotificationType = "user"    // 用户相关
	NotificationTypeLike    NotificationType = "like"    // 点赞通知
)

// Notification 站内通知模型
type Notification struct {
	ID         int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     int64            `json:"userId" gorm:"not null;index"`
	Type       NotificationType `json:"type" gorm:"size:20;not null"`
	Title      string           `json:"title" gorm:"size:200;not null"`
	Content    string           `json:"content" gorm:"type:text"`
	IsRead     bool             `json:"isRead" gorm:"default:false"`
	SourceID   string           `json:"sourceId" gorm:"size:50"` // 关联资源ID
	SourceType string           `json:"sourceType" gorm:"size:50"`
	CreatedAt  time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	ReadAt     *time.Time       `json:"readAt"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationSetting 用户通知设置
type NotificationSetting struct {
	ID             int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         int64 `json:"userId" gorm:"uniqueIndex;not null"`
	EmailOnComment bool  `json:"emailOnComment" gorm:"default:true"`
	EmailOnLike    bool  `json:"emailOnLike" gorm:"default:true"`
	EmailOnFollow  bool  `json:"emailOnFollow" gorm:"default:true"`
	EmailOnSystem  bool  `json:"emailOnSystem" gorm:"default:true"`
	SiteOnComment  bool  `json:"siteOnComment" gorm:"default:true"`
	SiteOnLike     bool  `json:"siteOnLike" gorm:"default:true"`
	SiteOnFollow   bool  `json:"siteOnFollow" gorm:"default:true"`
	SiteOnSystem   bool  `json:"siteOnSystem" gorm:"default:true"`
	UpdatedAt      time.Time
}

func (NotificationSetting) TableName() string {
	return "notification_settings"
}

// Subscription 邮件订阅
type Subscription struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Email       string `json:"email" gorm:"size:100;not null;index"`
	UserID      *int64 `json:"userId" gorm:"index"`
	Type        string `json:"type" gorm:"size:20;not null"`                // all/category/tag/author
	TargetID    string `json:"targetId" gorm:"size:50"`                     // 分类ID/标签ID/作者ID
	Frequency   string `json:"frequency" gorm:"size:20;default:'realtime'"` // realtime/daily/weekly
	IsActive    bool   `json:"isActive" gorm:"default:true"`
	VerifyToken string `json:"-" gorm:"size:100"`
	VerifiedAt  *time.Time
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

// Webhook Webhook配置
type Webhook struct {
	ID              int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string   `json:"name" gorm:"size:50;not null"`
	URL             string   `json:"url" gorm:"size:500;not null"`
	Secret          string   `json:"-" gorm:"size:255"`
	Events          []string `json:"events" gorm:"type:jsonb"` // ["article.created", "comment.created"]
	Method          string   `json:"method" gorm:"size:10;default:'POST'"`
	Headers         []byte   `json:"headers" gorm:"type:jsonb"`
	IsEnabled       bool     `json:"isEnabled" gorm:"default:true"`
	LastTriggeredAt *time.Time
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

// WebhookLog Webhook调用日志
type WebhookLog struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	WebhookID  int64     `json:"webhookId" gorm:"not null;index"`
	Event      string    `json:"event" gorm:"size:50;not null"`
	Payload    []byte    `json:"payload" gorm:"type:jsonb"`
	Response   string    `json:"response" gorm:"type:text"`
	StatusCode int       `json:"statusCode"`
	IsSuccess  bool      `json:"isSuccess"`
	RetryCount int       `json:"retryCount" gorm:"default:0"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (WebhookLog) TableName() string {
	return "webhook_logs"
}
