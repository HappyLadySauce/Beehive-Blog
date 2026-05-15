package model

import (
	"time"

	"gorm.io/gorm"
)

// FileNode maps to attachment.file_nodes.
// FileNode 映射到 attachment.file_nodes 表。
type FileNode struct {
	ID             int64          `gorm:"primaryKey;column:id"`
	ParentID       *int64         `gorm:"column:parent_id"`
	StorageMountID int64          `gorm:"column:storage_mount_id;not null"`
	NodeType       string         `gorm:"column:node_type;size:16;not null"`
	Name           string         `gorm:"column:name;size:255;not null"`
	Path           string         `gorm:"column:path;not null"`
	FullPath       string         `gorm:"column:full_path;not null"`
	Depth          int            `gorm:"column:depth;not null;default:0"`
	SortOrder      int            `gorm:"column:sort_order;not null;default:0"`
	Status         string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (FileNode) TableName() string {
	return "attachment.file_nodes"
}
