package models

import (
	"time"
)

// ThemeConfig 主题配置
type ThemeConfig struct {
	PrimaryColor   string `json:"primaryColor,omitempty"`
	SecondaryColor string `json:"secondaryColor,omitempty"`
	CustomCSS      string `json:"customCss,omitempty"`
	CustomJS       string `json:"customJs,omitempty"`
	LogoURL        string `json:"logoUrl,omitempty"`
	FaviconURL     string `json:"faviconUrl,omitempty"`
}

// Theme 主题模型
type Theme struct {
	ID          int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string      `json:"name" gorm:"size:50;not null"`
	Slug        string      `json:"slug" gorm:"uniqueIndex;size:50;not null"`
	Description string      `json:"description" gorm:"size:255"`
	Author      string      `json:"author" gorm:"size:50"`
	Version     string      `json:"version" gorm:"size:20"`
	Path        string      `json:"path" gorm:"size:255;not null"`
	Screenshot  string      `json:"screenshot" gorm:"size:500"`
	IsActive    bool        `json:"isActive" gorm:"default:false"`
	IsSystem    bool        `json:"isSystem" gorm:"default:false"`
	Config      ThemeConfig `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time   `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Theme) TableName() string {
	return "themes"
}

// Menu 菜单模型
type Menu struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	Location  string    `json:"location" gorm:"size:50;not null"` // header/footer/sidebar
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	IsEnabled bool      `json:"isEnabled" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Items []MenuItem `json:"items" gorm:"foreignKey:MenuID"`
}

func (Menu) TableName() string {
	return "menus"
}

// MenuItem 菜单项模型
type MenuItem struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuID   int64  `json:"menuId" gorm:"not null;index"`
	ParentID *int64 `json:"parentId" gorm:"index"`
	Name     string `json:"name" gorm:"size:50;not null"`
	Type     string `json:"type" gorm:"size:20;not null"` // link/page/category/tag
	TargetID string `json:"targetId" gorm:"size:50"`      // 关联的资源ID
	URL      string `json:"url" gorm:"size:500"`
	Icon     string `json:"icon" gorm:"size:50"`
	SortOrder int   `json:"sortOrder" gorm:"default:0"`
	IsEnabled bool  `json:"isEnabled" gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// 关联关系
	Parent   *MenuItem  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []MenuItem `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (MenuItem) TableName() string {
	return "menu_items"
}

// Page 独立页面模型
type Page struct {
	ID          int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string        `json:"title" gorm:"size:200;not null"`
	Slug        string        `json:"slug" gorm:"uniqueIndex;size:100;not null"`
	Content     string        `json:"content" gorm:"type:text;not null"`
	Status      ArticleStatus `json:"status" gorm:"size:20;default:'published'"`
	IsInMenu    bool          `json:"isInMenu" gorm:"default:false"`
	SortOrder   int           `json:"sortOrder" gorm:"default:0"`
	ViewCount   int64         `json:"viewCount" gorm:"default:0"`
	CreatedAt   time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time    `json:"-" gorm:"index"`
}

func (Page) TableName() string {
	return "pages"
}
