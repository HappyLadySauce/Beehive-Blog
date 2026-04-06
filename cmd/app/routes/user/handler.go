package user

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/middlewares"
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

func parseIDParam(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

// HandleLogout godoc
//
//	@Summary		用户登出
//	@Description	使当前访问令牌立即失效
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LogoutRequest	false	"登出参数（可空）"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/user/logout [post]
func (s *UserService) HandleLogout(c *gin.Context) {
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

// HandleMe godoc
//
//	@Summary		当前用户信息
//	@Description	返回已登录用户的公开资料
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/me [get]
func (s *UserService) HandleMe(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.GetMe(ctx, userID)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// HandleUpdateProfile godoc
//
//	@Summary		更新用户资料
//	@Description	更新昵称和/或头像
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.UpdateProfileRequest	true	"资料"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/profile [put]
func (s *UserService) HandleUpdateProfile(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	req := v1.UpdateProfileRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.ErrorS(err, "Could not read update profile request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.UpdateProfile(ctx, userID, &req)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// HandleUpdatePassword godoc
//
//	@Summary		修改密码
//	@Description	校验当前密码后更新；成功后撤销 Redis 会话快照需重新登录
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.UpdatePasswordRequest	true	"密码"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/password [put]
func (s *UserService) HandleUpdatePassword(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	req := v1.UpdatePasswordRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.ErrorS(err, "Could not read update password request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.UpdatePassword(ctx, userID, &req)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// HandleListNotifications godoc
//
//	@Summary		通知列表
//	@Description	分页获取当前用户的站内通知
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码，默认 1"
//	@Param			pageSize	query		int		false	"每页条数，默认 20，最大 100"
//	@Param			isRead		query		bool	false	"按已读筛选"
//	@Success		200	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/notifications [get]
func (s *UserService) HandleListNotifications(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := v1.NotificationListQuery{}
	if err := c.ShouldBindQuery(&q); err != nil {
		klog.ErrorS(err, "Could not read notification list query")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	page := q.Page
	pageSize := q.PageSize

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.ListNotifications(ctx, userID, page, pageSize, q.IsRead)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// HandleMarkNotificationRead godoc
//
//	@Summary		标记通知已读
//	@Description	将单条通知标记为已读
//	@Tags			user
//	@Produce		json
//	@Param			id	path		int	true	"通知 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.MarkNotificationReadResponse}
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/notifications/{id}/read [put]
func (s *UserService) HandleMarkNotificationRead(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		common.FailMessage(c, http.StatusBadRequest, "invalid notification id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.MarkNotificationRead(ctx, userID, id)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// HandleDeleteNotification godoc
//
//	@Summary		删除通知
//	@Description	删除当前用户的一条通知
//	@Tags			user
//	@Produce		json
//	@Param			id	path		int	true	"通知 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.DeleteNotificationResponse}
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/user/notifications/{id} [delete]
func (s *UserService) HandleDeleteNotification(c *gin.Context) {
	userID, ok := middlewares.GetCurrentUserID(c)
	if !ok || userID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		common.FailMessage(c, http.StatusBadRequest, "invalid notification id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, statusCode, err := s.DeleteNotification(ctx, userID, id)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}

// Init registers authenticated routes under /api/v1/user.
func Init(svcCtx *svc.ServiceContext) {
	userSvc := NewUserService(svcCtx)
	ug := router.V1().Group("/user")
	ug.Use(middlewares.Auth(svcCtx))
	ug.POST("/logout", userSvc.HandleLogout)
	ug.GET("/me", userSvc.HandleMe)
	ug.PUT("/profile", userSvc.HandleUpdateProfile)
	ug.PUT("/password", userSvc.HandleUpdatePassword)
	ug.PUT("/notifications/:id/read", userSvc.HandleMarkNotificationRead)
	ug.DELETE("/notifications/:id", userSvc.HandleDeleteNotification)
	ug.GET("/notifications", userSvc.HandleListNotifications)
}
