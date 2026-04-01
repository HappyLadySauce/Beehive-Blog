package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
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

// Login godoc
//
//	@Summary		用户登录
//	@Description	使用用户名或邮箱登录并返回访问令牌
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LoginRequest	true	"登录参数"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/auth/login [post]
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
	response, statusCode, err := s.Login(ctx, &loginRequest, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, response)
}

// Logout godoc
//
//	@Summary		用户登出
//	@Description	使当前访问令牌立即失效
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LogoutRequest	false	"登出参数（可空）"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/auth/logout [post]
func (s *AuthService) handleLogout(c *gin.Context) {
	logoutRequest := v1.LogoutRequest{}
	if err := c.ShouldBindJSON(&logoutRequest); err != nil && !errors.Is(err, io.EOF) {
		klog.ErrorS(err, "Could not read logout request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, statusCode, err := s.Logout(ctx, &logoutRequest, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, response)
}

func Init(svc *svc.ServiceContext) {
	r := router.V1().Group("auth")
	authService := NewAuthService(svc)
	r.POST("/login", middlewares.LoginAttemptLimit(), authService.handleLogin)
	r.POST("/logout", authService.handleLogout)
}
