package v1

type LoginRequest struct {
	// 账号 支持用户名或邮箱
	Account  string `json:"account" binding:"required,min=3,max=50"`
	// 密码 6-20位
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}
