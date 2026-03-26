package models

import (
	"time"
)

// ArticleStatus 文章状态
type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"     // 草稿
	ArticleStatusPublished ArticleStatus = "published" // 已发布
	ArticleStatusArchived  ArticleStatus = "archived"  // 已归档
	ArticleStatusPrivate   ArticleStatus = "private"   // 私密
	ArticleStatusScheduled ArticleStatus = "scheduled" // 定时发布
)

// Article 文章模型
type Article struct {
	ID          int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string        `json:"title" gorm:"size:200;not null"`
	Slug        string        `json:"slug" gorm:"uniqueIndex;size:100"`
	Content     string        `json:"content" gorm:"type:text;not null"`
	Summary     string        `json:"summary" gorm:"size:500"`
	CoverImage  string        `json:"coverImage" gorm:"size:500"`
	Status      ArticleStatus `json:"status" gorm:"size:20;default:'draft'"`
	Password    string        `json:"-" gorm:"size:100"`
	IsPinned    bool          `json:"isPinned" gorm:"default:false"`
	PinOrder    int           `json:"pinOrder" gorm:"default:0"`
	ViewCount   int64         `json:"viewCount" gorm:"default:0"`
	LikeCount   int64         `json:"likeCount" gorm:"default:0"`
	CommentCount int64        `json:"commentCount" gorm:"default:0"`
	AuthorID    int64         `json:"authorId" gorm:"not null;index"`
	CategoryID  *int64        `json:"categoryId" gorm:"index"`
	PublishedAt *time.Time    `json:"publishedAt"`
	CreatedAt   time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time    `json:"-" gorm:"index"`

	// 关联关系
	Author   User      `json:"author" gorm:"foreignKey:AuthorID"`
	Category *Category `json:"category" gorm:"foreignKey:CategoryID"`
	Tags     []Tag     `json:"tags" gorm:"many2many:article_tags;"`
}

func (Article) TableName() string {
	return "articles"
}

// IsPublished 检查文章是否已发布
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished
}

// IsProtected 检查文章是否受密码保护
func (a *Article) IsProtected() bool {
	return a.Password != ""
}

// ArticleVersion 文章版本历史
type ArticleVersion struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ArticleID int64     `json:"articleId" gorm:"not null;index"`
	Title     string    `json:"title" gorm:"size:200;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Version   int       `json:"version" gorm:"not null"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (ArticleVersion) TableName() string {
	return "article_versions"
}

// ArticleTag 文章标签关联
type ArticleTag struct {
	ArticleID int64     `json:"articleId" gorm:"primaryKey;autoIncrement:false"`
	TagID     int64     `json:"tagId" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (ArticleTag) TableName() string {
	return "article_tags"
}

// ArticleLike 文章点赞
type ArticleLike struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ArticleID int64     `json:"articleId" gorm:"uniqueIndex:idx_article_user;not null"`
	UserID    int64     `json:"userId" gorm:"uniqueIndex:idx_article_user;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (ArticleLike) TableName() string {
	return "article_likes"
}

// UserFavorite 用户收藏
type UserFavorite struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"userId" gorm:"uniqueIndex:idx_user_article;not null"`
	ArticleID int64     `json:"articleId" gorm:"uniqueIndex:idx_user_article;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (UserFavorite) TableName() string {
	return "user_favorites"
}

// ArticleViewLog 文章浏览记录
type ArticleViewLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ArticleID    int64     `json:"articleId" gorm:"not null;index"`
	UserID       *int64    `json:"userId" gorm:"index"`
	IP           string    `json:"ip" gorm:"size:50"`
	UserAgent    string    `json:"userAgent" gorm:"size:500"`
	ViewDuration int       `json:"viewDuration"` // 浏览时长（秒）
	IsValid      bool      `json:"isValid" gorm:"default:false"`
	ViewedAt     time.Time `json:"viewedAt" gorm:"autoCreateTime"`
}

func (ArticleViewLog) TableName() string {
	return "article_view_logs"
}
