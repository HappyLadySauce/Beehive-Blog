package v1

// TagListItem 标签列表项。
type TagListItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Color        string `json:"color"`
	Description  string `json:"description,omitempty"`
	ArticleCount int64  `json:"articleCount"`
	SortOrder    int    `json:"sortOrder"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

// TagListRequest 标签列表查询。
type TagListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=100"`
	Sort     string `form:"sort" binding:"omitempty,oneof=name count newest"`
}

// TagListResponse 标签分页。
type TagListResponse struct {
	List     []TagListItem `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// TagDetailRequest 标签详情内文章分页。
type TagDetailRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// RelatedTagItem 与当前标签常一起出现的标签。
type RelatedTagItem struct {
	TagListItem
	CoCount int64 `json:"coCount"`
}

// TagDetailResponse 标签详情。
type TagDetailResponse struct {
	TagListItem
	Articles    *ArticleListResponse `json:"articles,omitempty"`
	RelatedTags []RelatedTagItem     `json:"relatedTags,omitempty"`
}

// TagCloudRequest 标签云。
type TagCloudRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=200"`
}

// TagCloudResponse 标签云（按文章数降序截断）。
type TagCloudResponse struct {
	Tags []TagListItem `json:"tags"`
}

// AdminTagListRequest 管理员标签列表。
type AdminTagListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=200"`
	Keyword  string `form:"keyword" binding:"omitempty,max=100"`
	Sort     string `form:"sort" binding:"omitempty,oneof=name count newest"`
}

// CreateTagRequest 创建标签。
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Slug        string `json:"slug" binding:"omitempty,max=50"`
	Color       string `json:"color" binding:"omitempty,max=40"`
	Description string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int   `json:"sortOrder" binding:"omitempty"`
}

// UpdateTagRequest 更新标签。
type UpdateTagRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Slug        *string `json:"slug" binding:"omitempty,max=50"`
	Color       *string `json:"color" binding:"omitempty,max=40"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int    `json:"sortOrder" binding:"omitempty"`
}

// DeleteTagResponse 删除标签。
type DeleteTagResponse struct {
	ID int64 `json:"id"`
}
