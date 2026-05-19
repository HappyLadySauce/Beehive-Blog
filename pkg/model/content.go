package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Content maps to content.contents — the unified content table.
// Content 映射到 content.contents 统一内容表。
type Content struct {
	ID                 int64           `gorm:"primaryKey;column:id"`
	Type               string          `gorm:"column:type;size:32;not null"`
	Title              string          `gorm:"column:title;size:512;not null"`
	Slug               string          `gorm:"column:slug;size:512;not null"`
	Excerpt            *string         `gorm:"column:excerpt"`
	Body               *string         `gorm:"column:body"`
	CoverAttachmentID  *int64          `gorm:"column:cover_attachment_id"`
	AuthorID           int64           `gorm:"column:author_id;not null"`
	Status             string          `gorm:"column:status;size:16;not null;default:draft"`
	Visibility         string          `gorm:"column:visibility;size:16;not null;default:public"`
	AIAccess           string          `gorm:"column:ai_access;size:16;not null;default:allowed"`
	PublishedAt        *time.Time      `gorm:"column:published_at"`
	WordCount          int             `gorm:"column:word_count;not null;default:0"`
	ReadingTimeMinutes int             `gorm:"column:reading_time_minutes;not null;default:0"`
	Metadata           json.RawMessage `gorm:"column:metadata;type:jsonb;not null;default:'{}'"`
	ViewCount          int64           `gorm:"column:view_count;not null;default:0"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null"`
	DeletedAt          gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

// TableName returns the fully qualified table name for Content.
// TableName 返回 Content 的完全限定表名。
func (Content) TableName() string {
	return "content.contents"
}

// ContentVersion maps to content.content_versions.
// ContentVersion 映射到 content.content_versions 表。
type ContentVersion struct {
	ID            int64     `gorm:"primaryKey;column:id"`
	ContentID     int64     `gorm:"column:content_id;not null"`
	VersionNumber int       `gorm:"column:version_number;not null"`
	Title         string    `gorm:"column:title;size:512;not null"`
	Body          *string   `gorm:"column:body"`
	Excerpt       *string   `gorm:"column:excerpt"`
	ChangeSummary *string   `gorm:"column:change_summary;size:512"`
	CreatedBy     int64     `gorm:"column:created_by;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
}

// TableName returns the fully qualified table name for ContentVersion.
// TableName 返回 ContentVersion 的完全限定表名。
func (ContentVersion) TableName() string {
	return "content.content_versions"
}

// ContentRelation maps to content.content_relations.
// ContentRelation 映射到 content.content_relations 表。
type ContentRelation struct {
	ID              int64     `gorm:"primaryKey;column:id"`
	SourceContentID int64     `gorm:"column:source_content_id;not null"`
	TargetContentID int64     `gorm:"column:target_content_id;not null"`
	RelationType    string    `gorm:"column:relation_type;size:32;not null"`
	Label           *string   `gorm:"column:label;size:128"`
	SortOrder       int       `gorm:"column:sort_order;not null;default:0"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
}

// TableName returns the fully qualified table name for ContentRelation.
// TableName 返回 ContentRelation 的完全限定表名。
func (ContentRelation) TableName() string {
	return "content.content_relations"
}
