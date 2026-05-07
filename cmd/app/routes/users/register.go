package users

import (
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"

	"github.com/gin-gonic/gin"
)

// Register handles the user registration business logic.
// Register 处理用户注册的业务逻辑。
func (u *UsersController) Register(ctx *gin.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return nil, nil
}
