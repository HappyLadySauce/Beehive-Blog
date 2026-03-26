package models

import (
	"time"
)

// Tag 标签模型
type Tag struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Slug         string    `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Color        string    `json:"color" gorm:"size:10;default:'#3B82F6'"`
	Description  string    `json:"description" gorm:"size:255"`
	ArticleCount int64     `json:"articleCount" gorm:"default:0"`
	SortOrder    int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Articles []Article `json:"articles,omitempty" gorm:"many2many:article_tags;"`
}

func (Tag) TableName() string {
	return "tags"
}
