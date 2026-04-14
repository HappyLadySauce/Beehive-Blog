package v1

// AdminPageListRequest 管理员页面列表。
type AdminPageListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	Status   string `form:"status" binding:"omitempty,max=200"` // comma-separated draft,published,archived,private
	Sort     string `form:"sort" binding:"omitempty,oneof=newest oldest popular"`
}

// AdminPageListItem 管理员页面列表项。
type AdminPageListItem struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Status     string `json:"status"`
	ViewCount  int64  `json:"viewCount"`
	IsInMenu   bool   `json:"isInMenu"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// AdminPageListResponse 分页列表。
type AdminPageListResponse struct {
	List     []AdminPageListItem `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

// PageDetailResponse 页面详情（管理员）。
type PageDetailResponse struct {
	AdminPageListItem
	Content string `json:"content"`
}

// CreatePageRequest 创建页面。
type CreatePageRequest struct {
	Title     string `json:"title" binding:"required,min=1,max=200"`
	Slug      string `json:"slug" binding:"omitempty,max=100"`
	Content   string `json:"content" binding:"required"`
	Status    string `json:"status" binding:"omitempty,oneof=draft published archived private"`
	IsInMenu  *bool  `json:"isInMenu"`
	SortOrder *int   `json:"sortOrder"`
}

// UpdatePageRequest 更新页面（出现的字段才更新）。
type UpdatePageRequest struct {
	Title     *string `json:"title" binding:"omitempty,min=1,max=200"`
	Slug      *string `json:"slug" binding:"omitempty,max=100"`
	Content   *string `json:"content"`
	Status    *string `json:"status" binding:"omitempty,oneof=draft published archived private"`
	IsInMenu  *bool   `json:"isInMenu"`
	SortOrder *int    `json:"sortOrder"`
}

// UpdatePageStatusRequest 更新页面状态。
type UpdatePageStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft published archived private"`
}

// DeletePageResponse 删除/恢复占位响应。
type DeletePageResponse struct {
	ID int64 `json:"id"`
}
