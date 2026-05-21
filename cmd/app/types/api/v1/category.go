package v1

import "time"

// ListCategoriesRequest carries pagination and filter query params for category listing.
// ListCategoriesRequest 承载分类列表的分页与筛选查询参数。
type ListCategoriesRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=64"`
}

// CreateCategoryRequest is the admin-only payload for creating a category.
// CreateCategoryRequest 为管理员创建分类的请求体。
type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,max=64"`
	Slug        string  `json:"slug" binding:"required,max=64"`
	Description *string `json:"description,omitempty"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest is the PATCH payload for admin category updates.
// Pointer fields: nil = leave unchanged; pointer to value = set.
// UpdateCategoryRequest 为管理员更新分类的 PATCH 请求体。指针=nil 不修改，指针=值则设置。
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=64"`
	Slug        *string `json:"slug,omitempty" binding:"omitempty,max=64"`
	Description *string `json:"description,omitempty"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// CategoryItem is the response item for category listings.
// CategoryItem 为分类列表项响应。
type CategoryItem struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	SortOrder    int       `json:"sort_order"`
	ContentCount int64     `json:"content_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListCategoriesResponse wraps the paginated category result set.
// ListCategoriesResponse 封装分页的分类列表结果。
type ListCategoriesResponse struct {
	Items    []CategoryItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// CreateCategoryResponse returns the new category ID.
// CreateCategoryResponse 返回新建分类的 ID。
type CreateCategoryResponse struct {
	ID int64 `json:"id"`
}

// CategoryDetailResponse is the full category detail with content count.
// CategoryDetailResponse 为含内容计数的完整分类详情。
type CategoryDetailResponse struct {
	CategoryItem
}
