package models

import "time"

// ArticleAttachment 文章正文引用的附件（多对多，解析 Markdown 维护）。
type ArticleAttachment struct {
	ArticleID    int64     `json:"articleId" gorm:"primaryKey;index"`
	AttachmentID int64     `json:"attachmentId" gorm:"primaryKey;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`

	Article    Article    `json:"-" gorm:"foreignKey:ArticleID"`
	Attachment Attachment `json:"-" gorm:"foreignKey:AttachmentID"`
}

func (ArticleAttachment) TableName() string {
	return "article_attachments"
}
