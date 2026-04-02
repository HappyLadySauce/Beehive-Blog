package v1

import "time"

type RegisterRequest struct {
	// 用户名 3-20位
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	// 邮箱 50位
	Email string `json:"email" binding:"required,email,max=50"`
	// 密码 6-20位
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type RegisterResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// MeResponse is the authenticated user's public profile (no secrets).
type MeResponse struct {
	ID               int64      `json:"id"`
	Username         string     `json:"username"`
	Nickname         string     `json:"nickname"`
	Email            string     `json:"email"`
	Avatar           string     `json:"avatar"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	Level            int        `json:"level"`
	Experience       int        `json:"experience"`
	CommentCount     int        `json:"commentCount"`
	ArticleViewCount int        `json:"articleViewCount"`
	LastLoginAt      *time.Time `json:"lastLoginAt"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// UpdateProfileRequest updates nickname and/or avatar.
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Avatar   string `json:"avatar" binding:"omitempty,max=500"`
}

// UpdatePasswordRequest changes password after verifying the current one.
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=20"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}

// UpdatePasswordResponse confirms password change.
type UpdatePasswordResponse struct {
	Message string `json:"message"`
}

// NotificationListQuery binds query parameters for notification listing.
type NotificationListQuery struct {
	Page     int   `form:"page"`
	PageSize int   `form:"pageSize"`
	IsRead   *bool `form:"isRead"`
}

// NotificationItem is one row in the notification list.
type NotificationItem struct {
	ID         int64      `json:"id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	IsRead     bool       `json:"isRead"`
	SourceID   string     `json:"sourceId"`
	SourceType string     `json:"sourceType"`
	CreatedAt  time.Time  `json:"createdAt"`
	ReadAt     *time.Time `json:"readAt"`
}

// NotificationListResponse paginates notifications.
type NotificationListResponse struct {
	Items []NotificationItem `json:"items"`
	Total int64              `json:"total"`
}
