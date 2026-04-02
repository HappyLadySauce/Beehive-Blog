package models

import (
	"time"
)

// Category 分类模型（一级分类，无父子层级）。
type Category struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string `json:"name" gorm:"size:50;not null"`
	Slug         string `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Description  string `json:"description" gorm:"size:255"`
	ArticleCount int64  `json:"articleCount" gorm:"default:0"`
	SortOrder    int    `json:"sortOrder" gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Articles []Article `json:"articles,omitempty" gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "categories"
}
