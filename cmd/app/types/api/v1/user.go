package v1

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" binding:"required,email,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type RegisterResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}
