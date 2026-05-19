package model

import (
	"time"

	"gorm.io/gorm"
)

// Tag maps to content.tags — content categorization labels.
// Tag 映射到 content.tags 表——内容分类标签。
type Tag struct {
	ID          int64          `gorm:"primaryKey;column:id"`
	Name        string         `gorm:"column:name;size:64;not null"`
	Slug        string         `gorm:"column:slug;size:64;not null"`
	Description *string        `gorm:"column:description"`
	Color       *string        `gorm:"column:color;size:7"`
	Status      string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName returns the fully qualified table name for Tag.
// TableName 返回 Tag 的完全限定表名。
func (Tag) TableName() string {
	return "content.tags"
}

// ContentTag maps to content.content_tags — many-to-many junction.
// ContentTag 映射到 content.content_tags 表——多对多联结表。
type ContentTag struct {
	ContentID int64     `gorm:"primaryKey;column:content_id"`
	TagID     int64     `gorm:"primaryKey;column:tag_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName returns the fully qualified table name for ContentTag.
// TableName 返回 ContentTag 的完全限定表名。
func (ContentTag) TableName() string {
	return "content.content_tags"
}
