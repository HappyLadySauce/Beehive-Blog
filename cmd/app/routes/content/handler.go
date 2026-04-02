package content

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/router"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// HandleListArticles godoc
//
//	@Summary		文章列表
//	@Description	公开已发布文章，支持分类/标签 slug 筛选
//	@Tags			articles
//	@Produce		json
//	@Param			page		query	int		false	"页码"
//	@Param			pageSize	query	int		false	"每页条数"
//	@Param			keyword		query	string	false	"关键词"
//	@Param			category	query	string	false	"分类 slug"
//	@Param			tag			query	string	false	"标签 slug，逗号分隔多标签交集"
//	@Param			author		query	string	false	"作者用户名"
//	@Param			sort		query	string	false	"排序 newest|oldest|popular"
//	@Success		200	{object}	common.BaseResponse{data=v1.ArticleListResponse}
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/articles [get]
func (s *Service) HandleListArticles(c *gin.Context) {
	var req v1.ArticleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		klog.ErrorS(err, "ListArticles bind query")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.ListArticles(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleGetArticle godoc
//
//	@Summary		文章详情
//	@Description	公开已发布文章详情
//	@Tags			articles
//	@Produce		json
//	@Param			id	path		int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.ArticleDetailResponse}
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/articles/{id} [get]
func (s *Service) HandleGetArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	resp, code, err := s.GetArticle(ctx, id, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// HandleRecordArticleView godoc
//
//	@Summary		记录文章浏览
//	@Description	已发布文章浏览量 +1
//	@Tags			articles
//	@Produce		json
//	@Param			id	path		int	true	"文章 ID"
//	@Success		200	{object}	common.BaseResponse{data=v1.RecordArticleViewResponse}
//	@Failure		404	{object}	common.BaseResponse
//	@Router			/api/v1/articles/{id}/view [post]
func (s *Service) HandleRecordArticleView(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.FailMessage(c, http.StatusBadRequest, "invalid article id")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	resp, code, err := s.RecordArticleView(ctx, id, c.Request)
	if err != nil {
		common.Fail(c, code, err)
		return
	}
	common.Success(c, resp)
}

// Init registers public content routes on /api/v1.
func Init(svcCtx *svc.ServiceContext) {
	s := NewService(svcCtx)
	g := router.V1()
	g.GET("/articles", s.HandleListArticles)
	g.GET("/articles/:id", s.HandleGetArticle)
	g.POST("/articles/:id/view", s.HandleRecordArticleView)
}
