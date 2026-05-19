package v1

import "time"

// ListTagsRequest carries pagination and filter query params for tag listing.
// ListTagsRequest 承载标签列表的分页与筛选查询参数。
type ListTagsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active archived"`
	Search   string `form:"search" binding:"omitempty,max=64"`
}

// CreateTagRequest is the admin-only payload for creating a tag.
// CreateTagRequest 为管理员创建标签的请求体。
type CreateTagRequest struct {
	Name        string  `json:"name" binding:"required,max=64"`
	Slug        string  `json:"slug" binding:"required,max=64"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
}

// UpdateTagRequest is the PATCH payload for admin tag updates.
// Pointer fields: nil = leave unchanged; pointer to value = set.
// UpdateTagRequest 为管理员更新标签的 PATCH 请求体。指针=nil 不修改，指针=值则设置。
type UpdateTagRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=64"`
	Slug        *string `json:"slug,omitempty" binding:"omitempty,max=64"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=active archived"`
}

// TagItem is the response item for tag listings.
// TagItem 为标签列表项响应。
type TagItem struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  *string    `json:"description,omitempty"`
	Color        *string    `json:"color,omitempty"`
	Status       string     `json:"status"`
	ContentCount int64      `json:"content_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListTagsResponse wraps the paginated tag result set.
// ListTagsResponse 封装分页的标签列表结果。
type ListTagsResponse struct {
	Items    []TagItem `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// CreateTagResponse returns the new tag ID.
// CreateTagResponse 返回新建标签的 ID。
type CreateTagResponse struct {
	ID int64 `json:"id"`
}

// TagDetailResponse is the full tag detail with content count.
// TagDetailResponse 为含内容计数的完整标签详情。
type TagDetailResponse struct {
	TagItem
}
