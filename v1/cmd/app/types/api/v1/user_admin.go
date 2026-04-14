package v1

import "time"

// AdminUserListQuery 管理员用户列表查询。
type AdminUserListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=200"`
	Keyword  string `form:"keyword" binding:"omitempty,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=guest user admin"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive disabled deleted"`
}

// AdminUserItem 管理员用户列表项。
type AdminUserItem struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Email       string     `json:"email"`
	Avatar      string     `json:"avatar"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// AdminUserListResponse 管理员用户分页列表。
type AdminUserListResponse struct {
	Items    []AdminUserItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// AdminCreateUserRequest 管理员新建用户。
type AdminCreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Avatar   string `json:"avatar" binding:"omitempty,max=500"`
	Role     string `json:"role" binding:"omitempty,oneof=guest user admin"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive disabled"`
}

// AdminUpdateUserRequest 管理员更新用户资料。
type AdminUpdateUserRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=50"`
	Email    *string `json:"email" binding:"omitempty,email,max=100"`
	Avatar   *string `json:"avatar" binding:"omitempty,max=500"`
	Role     *string `json:"role" binding:"omitempty,oneof=guest user admin"`
}

// AdminUpdateUserStatusRequest 管理员更新用户状态。
type AdminUpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive disabled"`
}

// AdminResetUserPasswordRequest 管理员重置用户密码（直接设置新密码）。
type AdminResetUserPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}

// AdminResetUserPasswordResponse 密码重置结果。
type AdminResetUserPasswordResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// AdminDeleteUserResponse 删除结果。
type AdminDeleteUserResponse struct {
	ID int64 `json:"id"`
}
