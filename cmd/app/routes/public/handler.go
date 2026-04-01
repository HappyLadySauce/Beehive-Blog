package public

import (
	"context"
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

// HandleLogin godoc
//
//	@Summary		用户登录
//	@Description	使用用户名或邮箱登录并返回访问令牌
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LoginRequest	true	"登录参数"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/public/login [post]
func (s *PublicService) HandleLogin(c *gin.Context) {
	loginRequest := v1.LoginRequest{}
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		klog.ErrorS(err, "Could not read login request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, statusCode, err := s.Login(ctx, &loginRequest, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, response)
}

// HandleRegister godoc
//
//	@Summary		用户注册
//	@Description	创建新用户并返回访问令牌
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.RegisterRequest	true	"注册参数"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/public/register [post]
func (s *PublicService) HandleRegister(c *gin.Context) {
	registerRequest := v1.RegisterRequest{}
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		klog.ErrorS(err, "Could not read register request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, statusCode, err := s.Register(ctx, &registerRequest, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, response)
}

// Init registers unauthenticated routes under /api/v1/public.
func Init(svcCtx *svc.ServiceContext) {
	g := router.V1().Group("/public")
	pub := NewPublicService(svcCtx)
	g.POST("/login", middlewares.LoginAttemptLimit(), pub.HandleLogin)
	// 注册请求体无 account 字段，LoginAttemptLimit 不会生效；限流已由 v1 全局限流覆盖。
	g.POST("/register", pub.HandleRegister)
}
