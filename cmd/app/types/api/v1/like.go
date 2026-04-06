package v1

// LikeArticleResponse 点赞结果。
type LikeArticleResponse struct {
	ArticleID int64 `json:"articleId"`
	LikeCount int64 `json:"likeCount"`
}

// UnlikeArticleResponse 取消点赞结果。
type UnlikeArticleResponse struct {
	ArticleID int64 `json:"articleId"`
	LikeCount int64 `json:"likeCount"`
}
