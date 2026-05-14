package v1

import "time"

// RegisterRequest is the JSON body for user registration against identity.users writable columns plus password.
// RegisterRequest 为用户注册的 JSON 请求体：覆盖 identity.users 可由客户端写入的列，并包含密码（应用层凭证）。
type RegisterRequest struct {
	// Username is the unique login name among live rows (max 64 per DB).
	// Username 为活跃行内唯一的登录名（数据库最长 64）。
	Username string `json:"username" binding:"required,max=64"`
	// Password is the plaintext credential for hashing upstream of persistence (not a DB column yet).
	// Password 为明文凭证，供持久化前哈希使用（当前迁移尚无对应列）。
	Password string `json:"password" binding:"required,min=8,max=72"`
	// Email is optional; when set must be unique among live rows (max 320 per DB).
	// Email 可选；有值时在活跃行内唯一（数据库最长 320）。
	Email string `json:"email" binding:"omitempty,email,max=320"`
	// Nickname is an optional display name (max 128 per DB).
	// Nickname 为可选展示昵称（数据库最长 128）。
	Nickname string `json:"nickname" binding:"omitempty,max=128"`
	// Phone is an optional phone number (max 16 per DB).
	// Phone 为可选手机号（数据库最长 16）。
	Phone string `json:"phone" binding:"omitempty,max=16"`
}

// RegisterResponse is the safe public subset returned after a successful registration (no secrets except issued tokens).
// RegisterResponse 为注册成功后的安全公开字段子集（除签发的令牌外不含任何敏感数据）。
type RegisterResponse struct {
	// Token is the auth credential bundle granted on successful registration (auto-login).
	// Token 为注册成功后自动签发的鉴权凭证集合（自动登录）。
	Token AuthToken `json:"token"`
}

// ListUsersRequest carries pagination and filter query params for admin user listing.
// ListUsersRequest 承载管理员用户列表的分页与筛选查询参数。
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active disabled locked pending"`
	Role     string `form:"role" binding:"omitempty,oneof=member admin"`
	Search   string `form:"search" binding:"omitempty,max=64"`
}

// UserItem is the safe public subset for list responses.
// UserItem 为列表响应中的用户公开字段子集。
type UserItem struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       *string    `json:"email"`
	Nickname    *string    `json:"nickname"`
	Phone       *string    `json:"phone"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListUsersResponse wraps the paginated result set.
// ListUsersResponse 封装分页用户列表结果。
type ListUsersResponse struct {
	Items    []UserItem `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// UserDetailResponse extends UserItem with admin-only fields.
// UserDetailResponse 在 UserItem 基础上扩展管理员可见字段。
type UserDetailResponse struct {
	UserItem
	AvatarAttachmentID *int64 `json:"avatar_attachment_id"`
}

// CreateUserRequest is the admin-only payload for creating a user.
// Password is optional: when nil or empty the user is created without local credential (OAuth-only).
// CreateUserRequest 为管理员创建用户的请求体。密码可选：为空时不创建本地凭证（仅 OAuth 用户）。
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required,max=64"`
	Password *string `json:"password" binding:"omitempty,min=8,max=72"`
	Email    *string `json:"email" binding:"omitempty,max=320"`
	Nickname *string `json:"nickname" binding:"omitempty,max=128"`
	Phone    *string `json:"phone" binding:"omitempty,max=16"`
	Role     *string `json:"role" binding:"omitempty,oneof=member admin"`
	Status   *string `json:"status" binding:"omitempty,oneof=active disabled locked pending"`
}

// CreateUserResponse returns the new user's ID.
// CreateUserResponse 返回新建用户的 ID。
type CreateUserResponse struct {
	ID int64 `json:"id"`
}

// UpdateUserRequest is the PATCH payload for admin user updates.
// Pointer fields: nil = leave unchanged, pointer to empty string = clear to NULL, pointer to value = set.
// UpdateUserRequest 为管理员更新用户的 PATCH 请求体。指针字段：nil=不修改，指向空串=清空为 NULL，指向值=设置。
type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,max=64"`
	Email    *string `json:"email" binding:"omitempty,max=320"`
	Nickname *string `json:"nickname" binding:"omitempty,max=128"`
	Phone    *string `json:"phone" binding:"omitempty,max=16"`
	Role     *string `json:"role" binding:"omitempty,oneof=member admin"`
	Status   *string `json:"status" binding:"omitempty,oneof=active disabled locked pending"`
	Password *string `json:"password" binding:"omitempty,min=8,max=72"`
}

// UpdateUserResponse returns the full user detail after update.
// UpdateUserResponse 返回更新后的用户完整详情。
type UpdateUserResponse struct {
	UserDetailResponse
}
