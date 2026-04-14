package models

import (
	"time"
)

// CommentStatus 评论状态
type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"  // 待审核
	CommentStatusApproved CommentStatus = "approved" // 已通过
	CommentStatusRejected CommentStatus = "rejected" // 已拒绝
	CommentStatusSpam     CommentStatus = "spam"     // 垃圾评论
)

// Comment 评论模型
type Comment struct {
	ID          int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Content     string        `json:"content" gorm:"type:text;not null"`
	Status      CommentStatus `json:"status" gorm:"size:20;default:'pending'"`
	ArticleID   int64         `json:"articleId" gorm:"not null;index"`
	UserID      *int64        `json:"userId" gorm:"index"`
	ParentID    *int64        `json:"parentId" gorm:"index"`
	AuthorName  string        `json:"authorName,omitempty" gorm:"size:50"`   // 游客评论时使用
	AuthorEmail string        `json:"authorEmail,omitempty" gorm:"size:100"` // 游客评论时使用
	AuthorIP    string        `json:"-" gorm:"size:50"`
	LikeCount   int64         `json:"likeCount" gorm:"default:0"`
	CreatedAt   time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Article  Article   `json:"article,omitempty" gorm:"foreignKey:ArticleID"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Parent   *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Comment `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (Comment) TableName() string {
	return "comments"
}

// IsApproved 检查评论是否已通过审核
func (c *Comment) IsApproved() bool {
	return c.Status == CommentStatusApproved
}

// IsGuestComment 检查是否为游客评论
func (c *Comment) IsGuestComment() bool {
	return c.UserID == nil
}

// GetAuthorName 获取评论者名称
func (c *Comment) GetAuthorName() string {
	if c.User != nil && c.User.Nickname != "" {
		return c.User.Nickname
	}
	if c.User != nil {
		return c.User.Username
	}
	if c.AuthorName != "" {
		return c.AuthorName
	}
	return "匿名用户"
}

// CommentLike 评论点赞
type CommentLike struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CommentID int64     `json:"commentId" gorm:"uniqueIndex:idx_comment_user;not null"`
	UserID    int64     `json:"userId" gorm:"uniqueIndex:idx_comment_user;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

func (CommentLike) TableName() string {
	return "comment_likes"
}
