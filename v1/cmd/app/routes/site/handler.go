package site

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/types/common"
)

// HandleGetSettings godoc
//
//	@Summary		读取站点设置
//	@Description	需管理员；按 group 读取设置键值对（SMTP 密码脱敏）；group 可为 general/seo/smtp/comment/security/hexo（hexo 含只读 hexo.hexo_dir）
//	@Tags			admin
//	@Produce		json
//	@Param			group	path		string	true	"设置分组"
//	@Success		200		{object}	common.BaseResponse{data=v1.SettingsResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/settings/{group} [get]
func (s *Service) HandleGetSettings(c *gin.Context) {
	group := c.Param("group")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.GetSettingsValidated(ctx, group)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleUpdateSettings godoc
//
//	@Summary		更新站点设置
//	@Description	需管理员；批量 upsert 指定 group 的设置；SMTP 密码传 "***" 表示不修改；hexo 分组不可写入 hexo.hexo_dir
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			group	path		string						true	"设置分组"
//	@Param			request	body		v1.UpdateSettingsRequest	true	"设置键值对"
//	@Success		200		{object}	common.BaseResponse{data=v1.SettingsResponse}
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/settings/{group} [put]
func (s *Service) HandleUpdateSettings(c *gin.Context) {
	group := c.Param("group")
	var req v1.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.UpdateSettingsValidated(ctx, group, &req)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleTestSMTP godoc
//
//	@Summary		测试 SMTP 邮件发送
//	@Description	需管理员；使用当前数据库中的 SMTP 配置向指定邮箱发送测试邮件，不写库
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.TestSMTPRequest	true	"收件人邮箱"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		403		{object}	common.BaseResponse
//	@Failure		502		{object}	common.BaseResponse	"上游 SMTP 失败；message 为脱敏后的错误原因（非统一 internal server error）"
//	@Failure		500		{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/settings/smtp/test [post]
func (s *Service) HandleTestSMTP(c *gin.Context) {
	var req v1.TestSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	code, err := s.TestSMTP(ctx, &req)
	if err != nil {
		if code == http.StatusBadGateway {
			common.FailBadGateway(c, err)
			return
		}
		common.Fail(c, code, err)
		return
	}
	common.Success(c, gin.H{"message": "test email sent"})
}

// HandleGetStats godoc
//
//	@Summary		站点统计数据
//	@Description	需管理员；返回文章/用户/评论总数、今日访问量及热门文章 Top 10
//	@Tags			admin
//	@Produce		json
//	@Success		200	{object}	common.BaseResponse{data=v1.SiteStatsResponse}
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Security		BearerAuth
//	@Router			/api/v1/admin/stats [get]
func (s *Service) HandleGetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	resp, code, err := s.GetStats(ctx)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// RegisterAdminRoutes 在已挂载管理员鉴权的 RouterGroup 上注册站点设置与统计路由。
func RegisterAdminRoutes(g *gin.RouterGroup, svcCtx *svc.ServiceContext) {
	svc := NewService(svcCtx)
	// smtp/test 须在 :group 参数路由之前注册，避免路径冲突
	g.POST("/settings/smtp/test", svc.HandleTestSMTP)
	g.GET("/settings/:group", svc.HandleGetSettings)
	g.PUT("/settings/:group", svc.HandleUpdateSettings)
	g.GET("/stats", svc.HandleGetStats)
}
