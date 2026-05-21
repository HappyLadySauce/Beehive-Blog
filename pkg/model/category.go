package model

import (
	"time"

	"gorm.io/gorm"
)

// Category maps to content.categories — content taxonomy categories.
// Category 映射到 content.categories 表——内容分类。
type Category struct {
	ID          int64          `gorm:"primaryKey;column:id"`
	Name        string         `gorm:"column:name;size:64;not null"`
	Slug        string         `gorm:"column:slug;size:64;not null"`
	Description *string        `gorm:"column:description"`
	ParentID    *int64         `gorm:"column:parent_id"`
	SortOrder   int            `gorm:"column:sort_order;not null;default:0"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName returns the fully qualified table name for Category.
// TableName 返回 Category 的完全限定表名。
func (Category) TableName() string {
	return "content.categories"
}

// ContentCategory maps to content.content_categories — many-to-many junction.
// ContentCategory 映射到 content.content_categories 表——多对多联结表。
type ContentCategory struct {
	ContentID  int64     `gorm:"primaryKey;column:content_id"`
	CategoryID int64     `gorm:"primaryKey;column:category_id"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
}

// TableName returns the fully qualified table name for ContentCategory.
// TableName 返回 ContentCategory 的完全限定表名。
func (ContentCategory) TableName() string {
	return "content.content_categories"
}
