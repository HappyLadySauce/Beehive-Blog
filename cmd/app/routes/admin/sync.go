package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/sync"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/gin-gonic/gin"
)

func newHexoSyncService(svcCtx *svc.ServiceContext) (*sync.SyncService, error) {
	h := svcCtx.Config.HexoOptions
	return sync.NewSyncService(
		h.PostsDir,
		h.GenerateWorkdir,
		h.GenerateArgs,
		svcCtx.DB,
		svcCtx.Redis,
	)
}

// HandleSyncPosts godoc
//
//	@Summary		Hexo 文章全量同步
//	@Description	将已发布文章写入 Hexo source/_posts，可选执行 hexo 构建命令
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.SyncPostsRequest	false	"选项"
//	@Success		200	{object}	common.BaseResponse{data=v1.SyncResponse}
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/sync/posts [post]
func HandleSyncPosts(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req v1.SyncPostsRequest
		_ = c.ShouldBindJSON(&req)

		syncSvc, err := newHexoSyncService(svcCtx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		ctx := c.Request.Context()
		res, err := syncSvc.SyncAll(ctx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		if req.Rebuild && len(svcCtx.Config.HexoOptions.GenerateArgs) > 0 {
			genCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
			defer cancel()
			if err := syncSvc.RunHexoGenerate(genCtx); err != nil {
				common.Fail(c, http.StatusInternalServerError, err)
				return
			}
		}

		common.Success(c, v1.SyncResponse{
			Total:   res.Total,
			Created: res.Created,
			Updated: res.Updated,
			Deleted: res.Deleted,
			Files:   res.Files,
		})
	}
}

// HandleSyncStatus godoc
//
//	@Summary		Hexo 同步状态
//	@Description	查询上次同步时间与本地 beehive 文章文件数量等
//	@Tags			admin
//	@Produce		json
//	@Success		200	{object}	common.BaseResponse{data=v1.SyncStatusResponse}
//	@Failure		401	{object}	common.BaseResponse
//	@Failure		403	{object}	common.BaseResponse
//	@Failure		500	{object}	common.BaseResponse
//	@Router			/api/v1/admin/sync/status [get]
func HandleSyncStatus(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		syncSvc, err := newHexoSyncService(svcCtx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		ctx := c.Request.Context()

		total, err := syncSvc.PublishedArticleCount(ctx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		local, err := syncSvc.LocalBeehivePostCount()
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		pending, err := syncSvc.PendingSync(ctx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		last, err := syncSvc.LastSyncTime(ctx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		lastStr := ""
		if !last.IsZero() {
			lastStr = last.UTC().Format(time.RFC3339)
		}

		common.Success(c, v1.SyncStatusResponse{
			LastSyncTime: lastStr,
			TotalPosts:   total,
			LocalFiles:   local,
			PendingSync:  pending,
		})
	}
}
