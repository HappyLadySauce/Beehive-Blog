package model

import (
	"time"

	"gorm.io/gorm"
)

// Attachment maps to attachment.attachments.
// Attachment 映射到 attachment.attachments 表。
type Attachment struct {
	ID           int64          `gorm:"primaryKey;column:id"`
	OwnerUserID  *int64         `gorm:"column:owner_user_id"`
	Purpose      string         `gorm:"column:purpose;size:32;not null;default:content"`
	Filename     string         `gorm:"column:filename;size:255;not null"`
	OriginalName *string        `gorm:"column:original_name;size:255"`
	MimeType     string         `gorm:"column:mime_type;size:127;not null"`
	Size         int64          `gorm:"column:size;not null"`
	StorageType  string         `gorm:"column:storage_type;size:16;not null;default:local"`
	Bucket       *string        `gorm:"column:bucket;size:63"`
	ObjectKey    *string        `gorm:"column:object_key;size:1024"`
	LocalPath    *string        `gorm:"column:local_path;size:1024"`
	ETag         *string        `gorm:"column:etag;size:80"`
	Checksum     *string        `gorm:"column:checksum;size:128"`
	AccessScope  string         `gorm:"column:access_scope;size:16;not null;default:private"`
	UploadStatus string         `gorm:"column:upload_status;size:16;not null;default:ready"`
	Status       string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Attachment) TableName() string {
	return "attachment.attachments"
}

// AttachmentCategory maps to attachment.categories.
// AttachmentCategory 映射到 attachment.categories 表。
type AttachmentCategory struct {
	ID          int64          `gorm:"primaryKey;column:id"`
	ParentID    *int64         `gorm:"column:parent_id"`
	Name        string         `gorm:"column:name;size:64;not null"`
	Slug        string         `gorm:"column:slug;size:64;not null"`
	Description *string        `gorm:"column:description"`
	Icon        *string        `gorm:"column:icon;size:64"`
	Path        string         `gorm:"column:path;not null"`
	Depth       int            `gorm:"column:depth;not null;default:0"`
	SortOrder   int            `gorm:"column:sort_order;not null;default:0"`
	Status      string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (AttachmentCategory) TableName() string {
	return "attachment.categories"
}

// AttachmentCategoryBinding maps to attachment.attachment_categories.
// AttachmentCategoryBinding 映射到 attachment.attachment_categories 联结表。
type AttachmentCategoryBinding struct {
	AttachmentID int64     `gorm:"primaryKey;column:attachment_id"`
	CategoryID   int64     `gorm:"primaryKey;column:category_id"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (AttachmentCategoryBinding) TableName() string {
	return "attachment.attachment_categories"
}
