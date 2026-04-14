package v1

// SyncPostsRequest 手动全量同步请求体。
type SyncPostsRequest struct {
	// Rebuild 为 true 且在后台 Hexo 设置中配置了 generate_args（及可选 clean_args）时，在同步后执行生成。
	Rebuild bool `json:"rebuild"`
}

// SyncResponse 同步结果（含文章 _posts 与独立页面 beehive-pages）。
type SyncResponse struct {
	Total   int      `json:"total"`
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Deleted int      `json:"deleted"`
	Files   []string `json:"files"`

	PagesTotal   int      `json:"pages_total"`
	PagesCreated int      `json:"pages_created"`
	PagesUpdated int      `json:"pages_updated"`
	PagesDeleted int      `json:"pages_deleted"`
	PageFiles    []string `json:"page_files"`
}

// SyncStatusResponse 同步状态。
type SyncStatusResponse struct {
	LastSyncTime string `json:"last_sync_time"`
	TotalPosts   int64  `json:"total_posts"`
	LocalFiles   int    `json:"local_files"`
	PendingSync  bool   `json:"pending_sync"`

	TotalPages int64 `json:"total_pages"`
	LocalPages int   `json:"local_pages"`
}
