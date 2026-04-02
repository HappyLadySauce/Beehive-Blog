package v1

// CategoryBrief 分类摘要（公开列表项与管理员列表项）。
type CategoryBrief struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description,omitempty"`
	ArticleCount int64  `json:"articleCount"`
	SortOrder    int    `json:"sortOrder"`
}

// CategoryListResponse 公开分类一级列表。
type CategoryListResponse struct {
	List []CategoryBrief `json:"list"`
}

// CategoryDetailRequest 分类详情查询（文章分页）。
type CategoryDetailRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// CategoryDetailResponse 公开分类详情。
type CategoryDetailResponse struct {
	CategoryBrief
	Articles *ArticleListResponse `json:"articles,omitempty"`
}

// AdminCategoryListRequest 管理员分类列表。
type AdminCategoryListRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=200"`
}

// AdminCategoryListResponse 管理员分类扁平列表。
type AdminCategoryListResponse struct {
	List     []CategoryBrief `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// CreateCategoryRequest 创建分类。
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Slug        string `json:"slug" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int   `json:"sortOrder" binding:"omitempty"`
}

// UpdateCategoryRequest 更新分类。
type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Slug        *string `json:"slug" binding:"omitempty,max=50"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int    `json:"sortOrder" binding:"omitempty"`
}

// DeleteCategoryResponse 删除分类。
type DeleteCategoryResponse struct {
	ID int64 `json:"id"`
}
