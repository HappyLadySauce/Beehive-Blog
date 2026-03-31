package user

import (
	"context"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

type UserService struct {
	svc *svc.ServiceContext
}

func NewUserService(svc *svc.ServiceContext) *UserService {
	return &UserService{
		svc: svc,
	}
}

func (s *UserService) handleRegister(c *gin.Context) {
	registerRequest := v1.RegisterRequest{}
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		klog.ErrorS(err, "Could not read register request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}

	// 创建带超时的 context，5秒超时
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, statusCode, err := s.Register(ctx, &registerRequest, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, response)
}

func Init(svc *svc.ServiceContext) {
	r := router.V1().Group("user")
	userService := NewUserService(svc)
	r.POST("/register", userService.handleRegister)
}
