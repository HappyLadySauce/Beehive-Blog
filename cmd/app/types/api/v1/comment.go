package v1

import "time"

// CommentListQuery 公开评论列表查询。
type CommentListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// CommentAuthorItem 评论者展示（登录用户或游客占位）。
type CommentAuthorItem struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// CommentItem 单条评论（公开，仅已通过审核）。
type CommentItem struct {
	ID        int64             `json:"id"`
	Content   string            `json:"content"`
	ParentID  *int64            `json:"parentId,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	Author    CommentAuthorItem `json:"author"`
}

// CommentListResponse 文章下评论分页。
type CommentListResponse struct {
	Items    []CommentItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// CreateCommentRequest 登录用户发表评论。
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required,min=1,max=2000"`
	ParentID *int64 `json:"parentId"`
}

// CreateCommentResponse 发表评论结果。
type CreateCommentResponse struct {
	ID int64 `json:"id"`
}

// AdminCommentListQuery 管理员评论列表。
type AdminCommentListQuery struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	ArticleID int64  `form:"articleId"` // 0 表示不按文章筛选
	Status    string `form:"status" binding:"omitempty,oneof=pending approved rejected spam"`
	Keyword   string `form:"keyword" binding:"omitempty,max=200"`
}

// AdminCommentItem 管理员评论项。
type AdminCommentItem struct {
	ID        int64             `json:"id"`
	Content   string            `json:"content"`
	Status    string            `json:"status"`
	ArticleID int64             `json:"articleId"`
	UserID    *int64            `json:"userId,omitempty"`
	ParentID  *int64            `json:"parentId,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	Author    CommentAuthorItem `json:"author"`
}

// AdminCommentListResponse 管理员评论分页。
type AdminCommentListResponse struct {
	Items    []AdminCommentItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// UpdateCommentStatusRequest 审核评论。
type UpdateCommentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected spam"`
}

// UpdateCommentStatusResponse 审核结果。
type UpdateCommentStatusResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}
