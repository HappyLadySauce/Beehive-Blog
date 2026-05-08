package v1

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
	// AvatarAttachmentID links to attachment.attachments when the client pre-uploaded an avatar; nil means omit.
	// AvatarAttachmentID 在客户端已预上传头像时指向 attachment.attachments；指针为 nil 表示不传该字段。
	AvatarAttachmentID *int64 `json:"avatar_attachment_id" binding:"omitempty"`
}

// RegisterResponse is the safe public subset returned after a successful registration (no secrets except issued tokens).
// RegisterResponse 为注册成功后的安全公开字段子集（除签发的令牌外不含任何敏感数据）。
type RegisterResponse struct {
	// Token is the auth credential bundle granted on successful registration (auto-login).
	// Token 为注册成功后自动签发的鉴权凭证集合（自动登录）。
	Token AuthToken `json:"token"`
}
