package v1

// SettingsResponse 设置组读取结果。
type SettingsResponse struct {
	Group    string            `json:"group"`
	Settings map[string]string `json:"settings"`
}

// UpdateSettingsRequest 批量更新设置。
type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// TestSMTPRequest 测试 SMTP 发送。
type TestSMTPRequest struct {
	To string `json:"to" binding:"required,email"`
}

// ArticleStatItem 热门文章统计项。
type ArticleStatItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ViewCount int64  `json:"viewCount"`
	LikeCount int64  `json:"likeCount"`
}

// SiteStatsResponse 站点统计数据。
type SiteStatsResponse struct {
	ArticleCount int64             `json:"articleCount"`
	UserCount    int64             `json:"userCount"`
	CommentCount int64             `json:"commentCount"`
	TodayViews   int64             `json:"todayViews"`
	TopArticles  []ArticleStatItem `json:"topArticles"`
}
