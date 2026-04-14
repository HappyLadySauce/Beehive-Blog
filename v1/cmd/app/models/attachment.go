package models

import (
	"time"
)

// AttachmentType 附件类型
type AttachmentType string

const (
	AttachmentTypeImage    AttachmentType = "image"    // 图片
	AttachmentTypeDocument AttachmentType = "document" // 文档
	AttachmentTypeVideo    AttachmentType = "video"    // 视频
	AttachmentTypeAudio    AttachmentType = "audio"    // 音频
	AttachmentTypeOther    AttachmentType = "other"    // 其他
)

// StoragePolicyType 存储策略类型
type StoragePolicyType string

const (
	StoragePolicyLocal     StoragePolicyType = "local"      // 本地存储
	StoragePolicyAliyunOSS StoragePolicyType = "aliyun-oss" // 阿里云 OSS
	StoragePolicyAWSS3     StoragePolicyType = "aws-s3"     // AWS S3
	StoragePolicyMinIO     StoragePolicyType = "minio"      // MinIO
)

// Attachment 附件模型
type Attachment struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	OriginalName string         `json:"originalName" gorm:"size:255"`
	Path         string         `json:"path" gorm:"size:500;not null"`
	URL          string         `json:"url" gorm:"size:500;not null"`
	ThumbURL     string         `json:"thumbUrl" gorm:"size:500"`
	Type         AttachmentType `json:"type" gorm:"size:20;not null"`
	MimeType     string         `json:"mimeType" gorm:"size:100"`
	Size         int64          `json:"size" gorm:"not null"`
	Width        int            `json:"width"`  // 图片宽度
	Height       int            `json:"height"` // 图片高度
	PolicyID     int64          `json:"policyId" gorm:"not null;index"`
	GroupID      *int64         `json:"groupId" gorm:"index"`
	ParentID     *int64         `json:"parentId" gorm:"index"` // 非空表示由父附件派生（压缩/转换/复制）
	Variant      string         `json:"variant" gorm:"size:32"` // original|compressed|converted|copy 等，可选
	UploadedBy   int64          `json:"uploadedBy" gorm:"not null;index"`
	CreatedAt    time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Policy   StoragePolicy    `json:"policy" gorm:"foreignKey:PolicyID"`
	Group    *AttachmentGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	User     User             `json:"user" gorm:"foreignKey:UploadedBy"`
	Parent   *Attachment      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Attachment     `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (Attachment) TableName() string {
	return "attachments"
}

// IsImage 检查是否为图片
func (a *Attachment) IsImage() bool {
	return a.Type == AttachmentTypeImage
}

// StoragePolicyConfig 存储策略配置
type StoragePolicyConfig struct {
	Endpoint  string `json:"endpoint,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
	Region    string `json:"region,omitempty"`
}

// StoragePolicy 存储策略模型
type StoragePolicy struct {
	ID         int64               `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string              `json:"name" gorm:"size:50;not null"`
	Type       StoragePolicyType   `json:"type" gorm:"size:20;not null"`
	IsDefault  bool                `json:"isDefault" gorm:"default:false"`
	BaseURL    string              `json:"baseUrl" gorm:"size:500"`
	UploadPath string              `json:"uploadPath" gorm:"size:255"`
	Config     StoragePolicyConfig `json:"config" gorm:"type:jsonb;serializer:json"` // JSON 格式存储配置
	SortOrder  int                 `json:"sortOrder" gorm:"default:0"`
	CreatedAt  time.Time           `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time           `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (StoragePolicy) TableName() string {
	return "storage_policies"
}

// AttachmentGroup 附件分组模型
type AttachmentGroup struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:50;not null"`
	ParentID  *int64    `json:"parentId" gorm:"index"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// 关联关系
	Parent   *AttachmentGroup  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []AttachmentGroup `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (AttachmentGroup) TableName() string {
	return "attachment_groups"
}
