package v1

type RegisterRequest struct {
	// 用户名 3-20位
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	// 邮箱 50位
	Email    string `json:"email" binding:"required,email,max=50"`
	// 密码 6-20位
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type RegisterResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}
