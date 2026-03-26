package models

import (
	"time"
)

// Category 分类模型
type Category struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string `json:"name" gorm:"size:50;not null"`
	Slug         string `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Description  string `json:"description" gorm:"size:255"`
	ParentID     *int64 `json:"parentId" gorm:"index"`
	ArticleCount int64  `json:"articleCount" gorm:"default:0"`
	SortOrder    int    `json:"sortOrder" gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// 关联关系
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Articles []Article  `json:"articles,omitempty" gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "categories"
}

// IsTopLevel 检查是否为顶级分类
func (c *Category) IsTopLevel() bool {
	return c.ParentID == nil
}
