package user

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/middlewares"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/common"
	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 在管理员分组下注册用户管理路由。
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	userSvc := NewUserService(svcCtx)
	g.GET("/users", userSvc.HandleAdminListUsers)
	g.POST("/users", userSvc.HandleAdminCreateUser)
	g.GET("/users/:id", userSvc.HandleAdminGetUser)
	g.PUT("/users/:id", userSvc.HandleAdminUpdateUser)
	g.PUT("/users/:id/status", userSvc.HandleAdminUpdateUserStatus)
	g.PUT("/users/:id/password/reset", userSvc.HandleAdminResetUserPassword)
	g.DELETE("/users/:id", userSvc.HandleAdminDeleteUser)
}

// HandleAdminCreateUser godoc
//
//	@Summary		管理员新建用户
//	@Description	需管理员；支持设置用户基本信息、角色与状态
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.AdminCreateUserRequest	true	"新建用户"
//	@Success		200		{object}	common.BaseResponse{data=v1.AdminUserItem}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/users [post]
func (s *UserService) HandleAdminCreateUser(c *gin.Context) {
	var req v1.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminCreateUser(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminListUsers godoc
//
//	@Summary		管理员用户列表
//	@Description	需管理员；支持分页与关键词/角色/状态筛选
//	@Tags			admin
//	@Produce		json
//	@Param			page		query		int		false	"页码"
//	@Param			pageSize	query		int		false	"每页条数"
//	@Param			keyword		query		string	false	"用户名/昵称/邮箱关键词"
//	@Param			role		query		string	false	"角色 guest|user|admin"
//	@Param			status		query		string	false	"状态 active|inactive|disabled|deleted"
//	@Success		200			{object}	common.BaseResponse{data=v1.AdminUserListResponse}
//	@Failure		401			{object}	common.BaseResponse
//	@Failure		403			{object}	common.BaseResponse
//	@Failure		500			{object}	common.BaseResponse
//	@Router			/api/v1/admin/users [get]
func (s *UserService) HandleAdminListUsers(c *gin.Context) {
	var req v1.AdminUserListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.AdminListUsers(ctx, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminGetUser godoc
//
//	@Summary		管理员获取用户详情
//	@Description	需管理员
//	@Tags			admin
//	@Produce		json
//	@Param			id	path		int	true	"用户 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.AdminUserItem}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/users/{id} [get]
func (s *UserService) HandleAdminGetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.AdminGetUser(ctx, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminUpdateUser godoc
//
//	@Summary		管理员更新用户资料
//	@Description	需管理员；可更新昵称、邮箱、头像与角色
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"用户 ID"
//	@Param			request	body		v1.AdminUpdateUserRequest	true	"更新字段"
//	@Success		200		{object}	common.BaseResponse{data=v1.AdminUserItem}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		409		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/users/{id} [put]
func (s *UserService) HandleAdminUpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req v1.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminUpdateUser(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminUpdateUserStatus godoc
//
//	@Summary		管理员更新用户状态
//	@Description	需管理员；支持 active/inactive/disabled
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"用户 ID"
//	@Param			request	body		v1.AdminUpdateUserStatusRequest	true	"状态"
//	@Success		200		{object}	common.BaseResponse{data=v1.AdminUserItem}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/users/{id}/status [put]
func (s *UserService) HandleAdminUpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req v1.AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminUpdateUserStatus(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminResetUserPassword godoc
//
//	@Summary		管理员重置用户密码
//	@Description	需管理员；直接设置新密码并使目标用户会话失效
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int									true	"用户 ID"
//	@Param			request	body		v1.AdminResetUserPasswordRequest	true	"新密码"
//	@Success		200		{object}	common.BaseResponse{data=v1.AdminResetUserPasswordResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		404		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/admin/users/{id}/password/reset [put]
func (s *UserService) HandleAdminResetUserPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req v1.AdminResetUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminResetUserPassword(ctx, id, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleAdminDeleteUser godoc
//
//	@Summary		管理员删除用户
//	@Description	需管理员；执行软删除，禁止删除自己与最后一个管理员
//	@Tags			admin
//	@Produce		json
//	@Param			id	path		int	true	"用户 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.AdminDeleteUserResponse}
//	@Failure		400	{object}	common.BaseResponse
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		404	{object}	common.BaseResponse
//	@Failure		409	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/users/{id} [delete]
func (s *UserService) HandleAdminDeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	operatorID, ok := middlewares.GetCurrentUserID(c)
	if !ok || operatorID <= 0 {
		common.FailMessage(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.AdminDeleteUser(ctx, operatorID, id)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}
