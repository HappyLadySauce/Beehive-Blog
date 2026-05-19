package v1

import (
	"encoding/json"
	"time"
)

// ListContentsRequest carries pagination and filter query params for content listing.
// ListContentsRequest 承载内容列表的分页与筛选查询参数。
type ListContentsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type       string `form:"type" binding:"omitempty,oneof=article note project experience reflection portfolio"`
	Status     string `form:"status" binding:"omitempty,oneof=draft review published archived"`
	Visibility string `form:"visibility" binding:"omitempty,oneof=public member private"`
	TagID      int64  `form:"tag_id" binding:"omitempty,min=1"`
	Search     string `form:"search" binding:"omitempty,max=256"`
}

// CreateContentRequest is the admin-only payload for creating content.
// CreateContentRequest 为管理员创建内容的请求体。
type CreateContentRequest struct {
	Type               string          `json:"type" binding:"required,oneof=article note project experience reflection portfolio"`
	Title              string          `json:"title" binding:"required,max=512"`
	Slug               string          `json:"slug" binding:"required,max=512"`
	Excerpt            *string         `json:"excerpt,omitempty"`
	Body               *string         `json:"body,omitempty"`
	CoverAttachmentID  *int64          `json:"cover_attachment_id,omitempty"`
	Status             *string         `json:"status,omitempty" binding:"omitempty,oneof=draft review published archived"`
	Visibility         *string         `json:"visibility,omitempty" binding:"omitempty,oneof=public member private"`
	AIAccess           *string         `json:"ai_access,omitempty" binding:"omitempty,oneof=allowed denied"`
	WordCount          *int            `json:"word_count,omitempty"`
	ReadingTimeMinutes *int            `json:"reading_time_minutes,omitempty"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
}

// UpdateContentRequest is the PATCH payload for admin content updates.
// Pointer fields: nil = leave unchanged; pointer to value = set.
// UpdateContentRequest 为管理员更新内容的 PATCH 请求体。指针=nil 不修改，指针=值则设置。
type UpdateContentRequest struct {
	Type               *string          `json:"type,omitempty" binding:"omitempty,oneof=article note project experience reflection portfolio"`
	Title              *string          `json:"title,omitempty" binding:"omitempty,max=512"`
	Slug               *string          `json:"slug,omitempty" binding:"omitempty,max=512"`
	Excerpt            *string          `json:"excerpt,omitempty"`
	Body               *string          `json:"body,omitempty"`
	CoverAttachmentID  *int64           `json:"cover_attachment_id,omitempty"`
	Visibility         *string          `json:"visibility,omitempty" binding:"omitempty,oneof=public member private"`
	AIAccess           *string          `json:"ai_access,omitempty" binding:"omitempty,oneof=allowed denied"`
	WordCount          *int             `json:"word_count,omitempty"`
	ReadingTimeMinutes *int             `json:"reading_time_minutes,omitempty"`
	Metadata           *json.RawMessage `json:"metadata,omitempty"`
}

// TransitionStatusRequest is the payload for changing content status.
// TransitionStatusRequest 为变更内容状态的请求体。
type TransitionStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft review published archived"`
}

// CreateVersionRequest is the payload for creating a content version snapshot.
// CreateVersionRequest 为创建内容版本快照的请求体。
type CreateVersionRequest struct {
	ChangeSummary *string `json:"change_summary,omitempty" binding:"omitempty,max=512"`
}

// AddRelationRequest is the payload for adding a content relation.
// AddRelationRequest 为添加内容关系的请求体。
type AddRelationRequest struct {
	TargetContentID int64   `json:"target_content_id" binding:"required,min=1"`
	RelationType    string  `json:"relation_type" binding:"required,oneof=references part_of derived_from follows"`
	Label           *string `json:"label,omitempty" binding:"omitempty,max=128"`
	SortOrder       *int    `json:"sort_order,omitempty"`
}

// SetContentTagsRequest is the payload for replacing all tags on a content.
// SetContentTagsRequest 为替换内容全部标签的请求体。
type SetContentTagsRequest struct {
	TagIDs []int64 `json:"tag_ids" binding:"required"`
}

// ContentItem is the admin-safe response item for content listings.
// ContentItem 为管理员内容列表项响应。
type ContentItem struct {
	ID                 int64           `json:"id"`
	Type               string          `json:"type"`
	Title              string          `json:"title"`
	Slug               string          `json:"slug"`
	Excerpt            *string         `json:"excerpt,omitempty"`
	CoverAttachmentID  *int64          `json:"cover_attachment_id,omitempty"`
	AuthorID           int64           `json:"author_id"`
	AuthorUsername     string          `json:"author_username,omitempty"`
	Status             string          `json:"status"`
	Visibility         string          `json:"visibility"`
	AIAccess           string          `json:"ai_access"`
	PublishedAt        *time.Time      `json:"published_at,omitempty"`
	WordCount          int             `json:"word_count"`
	ReadingTimeMinutes int             `json:"reading_time_minutes"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
	ViewCount          int64           `json:"view_count"`
	Tags               []TagItem       `json:"tags,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// PublicContentItem is the public-safe subset for reader-facing listings.
// PublicContentItem 为面向读者的公开字段子集。
type PublicContentItem struct {
	ID                 int64           `json:"id"`
	Type               string          `json:"type"`
	Title              string          `json:"title"`
	Slug               string          `json:"slug"`
	Excerpt            *string         `json:"excerpt,omitempty"`
	CoverAttachmentID  *int64          `json:"cover_attachment_id,omitempty"`
	AuthorID           int64           `json:"author_id"`
	AuthorUsername     string          `json:"author_username,omitempty"`
	PublishedAt        *time.Time      `json:"published_at,omitempty"`
	WordCount          int             `json:"word_count"`
	ReadingTimeMinutes int             `json:"reading_time_minutes"`
	Metadata           json.RawMessage `json:"metadata"`
	Tags               []TagItem       `json:"tags,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// ListContentsResponse wraps the paginated admin content result set.
// ListContentsResponse 封装分页的管理员内容列表结果。
type ListContentsResponse struct {
	Items    []ContentItem `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// PublicListContentsResponse wraps the paginated public content result set.
// PublicListContentsResponse 封装分页的公开内容列表结果。
type PublicListContentsResponse struct {
	Items    []PublicContentItem `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// CreateContentResponse returns the new content ID.
// CreateContentResponse 返回新建内容的 ID。
type CreateContentResponse struct {
	ID int64 `json:"id"`
}

// ContentDetailResponse is the full content detail (admin view).
// ContentDetailResponse 为完整内容详情（管理员视图）。
type ContentDetailResponse struct {
	ContentItem
	Body *string `json:"body,omitempty"`
}

// PublicContentDetailResponse is the public-safe content detail.
// PublicContentDetailResponse 为面向读者的内容详情。
type PublicContentDetailResponse struct {
	PublicContentItem
	Body *string `json:"body,omitempty"`
}

// VersionItem is a single content version entry.
// VersionItem 为单个内容版本条目。
type VersionItem struct {
	ID            int64     `json:"id"`
	ContentID     int64     `json:"content_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `json:"title"`
	Body          *string   `json:"body,omitempty"`
	Excerpt       *string   `json:"excerpt,omitempty"`
	ChangeSummary *string   `json:"change_summary,omitempty"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// ListVersionsResponse wraps the version list for a content item.
// ListVersionsResponse 封装某个内容的版本列表。
type ListVersionsResponse struct {
	Items []VersionItem `json:"items"`
}

// CreateVersionResponse returns the new version entry.
// CreateVersionResponse 返回新建的版本条目。
type CreateVersionResponse struct {
	VersionItem
}

// ContentRelationItem is a single content relation entry with target summary.
// ContentRelationItem 为单个内容关系条目，含目标摘要。
type ContentRelationItem struct {
	ID              int64     `json:"id"`
	SourceContentID int64     `json:"source_content_id"`
	TargetContentID int64     `json:"target_content_id"`
	RelationType    string    `json:"relation_type"`
	Label           *string   `json:"label,omitempty"`
	SortOrder       int       `json:"sort_order"`
	TargetTitle     string    `json:"target_title,omitempty"`
	TargetType      string    `json:"target_type,omitempty"`
	TargetSlug      string    `json:"target_slug,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListContentRelationsResponse wraps the relations list for a content item.
// ListContentRelationsResponse 封装某个内容的关系列表。
type ListContentRelationsResponse struct {
	Items []ContentRelationItem `json:"items"`
}

// AddRelationResponse returns the new relation entry.
// AddRelationResponse 返回新建的关系条目。
type AddRelationResponse struct {
	ContentRelationItem
}
