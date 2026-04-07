package v1

// ArticleListRequest 文章列表查询（公开）。
type ArticleListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	Category string `form:"category" binding:"omitempty,max=50"`
	Tag      string `form:"tag" binding:"omitempty,max=500"`
	Author   string `form:"author" binding:"omitempty,max=50"`
	Sort     string `form:"sort" binding:"omitempty,oneof=newest oldest popular"`
}

// ArticleAuthorItem 作者摘要。
type ArticleAuthorItem struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ArticleCategoryItem 分类摘要。
type ArticleCategoryItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ArticleTagItem 标签摘要。
type ArticleTagItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Color string `json:"color"`
}

// ArticleNavItem 上一篇/下一篇。
type ArticleNavItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// ArticleListItem 列表项。
type ArticleListItem struct {
	ID           int64                `json:"id"`
	Title        string               `json:"title"`
	Slug         string               `json:"slug"`
	Summary      string               `json:"summary"`
	CoverImage   string               `json:"coverImage"`
	IsPinned     bool                 `json:"isPinned"`
	PinOrder     int                  `json:"pinOrder"`
	ViewCount    int64                `json:"viewCount"`
	LikeCount    int64                `json:"likeCount"`
	CommentCount int64                `json:"commentCount"`
	PublishedAt  string               `json:"publishedAt,omitempty"`
	UpdatedAt    string               `json:"updatedAt"`
	Author       ArticleAuthorItem    `json:"author"`
	Category     *ArticleCategoryItem `json:"category,omitempty"`
	Tags         []ArticleTagItem     `json:"tags"`
}

// ArticleListResponse 分页列表。
type ArticleListResponse struct {
	List     []ArticleListItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// AdminArticleListRequest 管理员文章列表查询（含草稿等）。
type AdminArticleListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	Category string `form:"category" binding:"omitempty,max=50"`
	Tag      string `form:"tag" binding:"omitempty,max=500"`
	Author   string `form:"author" binding:"omitempty,max=50"`
	Status   string `form:"status" binding:"omitempty,max=200"` // comma-separated: draft,published,...
	Sort     string `form:"sort" binding:"omitempty,oneof=newest oldest popular"`
}

// AdminArticleListItem 管理员列表项（含状态）。
type AdminArticleListItem struct {
	ArticleListItem
	Status string `json:"status"`
}

// AdminArticleListResponse 管理员分页列表。
type AdminArticleListResponse struct {
	List     []AdminArticleListItem `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// ArticleDetailResponse 详情。
type ArticleDetailResponse struct {
	ArticleListItem
	Content   string          `json:"content"`
	Protected bool            `json:"protected"`
	Previous  *ArticleNavItem `json:"previous,omitempty"`
	Next      *ArticleNavItem `json:"next,omitempty"`
}

// RecordArticleViewResponse 浏览记录响应。
type RecordArticleViewResponse struct {
	ViewCount int64 `json:"viewCount"`
}

// CreateArticleRequest 创建文章（管理员）。
type CreateArticleRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Slug        string  `json:"slug" binding:"omitempty,max=100"`
	Content     string  `json:"content" binding:"required,min=1"`
	Summary     string  `json:"summary" binding:"omitempty,max=500"`
	CoverImage  string  `json:"coverImage" binding:"omitempty,max=500"`
	CategoryID  *int64  `json:"categoryId"`
	TagIDs      []int64 `json:"tagIds"`
	Status      string  `json:"status" binding:"omitempty,oneof=draft published archived private scheduled"`
	PublishedAt *string `json:"publishedAt" binding:"omitempty"` // RFC3339
}

// UpdateArticleRequest 更新文章（管理员，字段均可选）。
type UpdateArticleRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Slug        *string `json:"slug" binding:"omitempty,max=100"`
	Content     *string `json:"content" binding:"omitempty,min=1"`
	Summary     *string `json:"summary" binding:"omitempty,max=500"`
	CoverImage  *string `json:"coverImage" binding:"omitempty,max=500"`
	CategoryID  *int64  `json:"categoryId"`
	TagIDs      []int64 `json:"tagIds"`
	Status      *string `json:"status" binding:"omitempty,oneof=draft published archived private scheduled"`
	PublishedAt *string `json:"publishedAt" binding:"omitempty"`
	// AutoSave 为 true 时，内容变更写入「自动保存」单槽（覆盖同文唯一一条），不递增手动版本号。
	AutoSave *bool `json:"autoSave"`
}

// UpdateArticleStatusRequest 状态变更。
type UpdateArticleStatusRequest struct {
	Status      string  `json:"status" binding:"required,oneof=draft published archived private scheduled"`
	PublishedAt *string `json:"publishedAt" binding:"omitempty"`
}

// UpdateArticleSlugRequest 别名。
type UpdateArticleSlugRequest struct {
	Slug string `json:"slug" binding:"required,min=1,max=100"`
}

// UpdateArticlePasswordRequest 密码（空字符串表示取消保护）。
type UpdateArticlePasswordRequest struct {
	Password string `json:"password" binding:"max=100"`
}

// UpdateArticlePinRequest 置顶。
type UpdateArticlePinRequest struct {
	IsPinned bool `json:"isPinned"`
	PinOrder int  `json:"pinOrder" binding:"min=0,max=1000000"`
}

// DeleteArticleResponse 删除结果。
type DeleteArticleResponse struct {
	ID int64 `json:"id"`
}

// ArticleSecurityResponse 密码更新结果（不落密文）。
type ArticleSecurityResponse struct {
	Protected bool `json:"protected"`
}

// ArticleVersionItem 单个版本摘要。
type ArticleVersionItem struct {
	ID         int64  `json:"id"`
	ArticleID  int64  `json:"articleId"`
	Title      string `json:"title"`
	Version    int    `json:"version"`
	IsAutosave bool   `json:"isAutosave"`
	CreatedBy  int64  `json:"createdBy"`
	CreatedAt  string `json:"createdAt"`
}

// ArticleVersionListResponse 版本列表。
type ArticleVersionListResponse struct {
	Items []ArticleVersionItem `json:"items"`
	Total int                  `json:"total"`
}

// UpdateArticleVersionRequest 修改版本快照显示名称（仅更新 article_versions.title）。
type UpdateArticleVersionRequest struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
}

// DeleteArticleVersionResponse 删除版本记录结果。
type DeleteArticleVersionResponse struct {
	ID int64 `json:"id"`
}

// BatchArticlePayload 批量操作的附加参数。
type BatchArticlePayload struct {
	Status     string  `json:"status"`
	CategoryID *int64  `json:"categoryId"`
	TagIDs     []int64 `json:"tagIds"`
}

// BatchArticleRequest 批量操作请求。
type BatchArticleRequest struct {
	IDs     []int64             `json:"ids" binding:"required,min=1"`
	Action  string              `json:"action" binding:"required,oneof=delete set_status set_category set_tags"`
	Payload BatchArticlePayload `json:"payload"`
}

// BatchArticleResponse 批量操作结果。
type BatchArticleResponse struct {
	Affected int64 `json:"affected"`
}
