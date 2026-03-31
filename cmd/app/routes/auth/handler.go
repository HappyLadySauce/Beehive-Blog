package auth

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

type AuthService struct {
	svc *svc.ServiceContext
}

func NewAuthService(svc *svc.ServiceContext) *AuthService {
	return &AuthService{
		svc: svc,
	}
}

func (s *AuthService) handleLogin(c *gin.Context) {
	loginRequest := v1.LoginRequest{}
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		klog.ErrorS(err, "Could not read login request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	// 创建带超时的 context，5秒超时
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 调用登录服务
	response, _, err := s.Login(ctx, &loginRequest, c.Request)
	if err != nil {
		common.Fail(c, http.StatusUnauthorized, err)
		return
	}
	common.Success(c, response)
}

func Init(svc *svc.ServiceContext) {
	r := router.V1().Group("auth")
	authService := NewAuthService(svc)
	r.POST("/login", authService.handleLogin)
}