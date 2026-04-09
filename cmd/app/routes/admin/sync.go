package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/hexocfg"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/sync"
	"github.com/gin-gonic/gin"
)

func newHexoSyncServiceCtx(svcCtx *svc.ServiceContext, eff *hexocfg.EffectiveHexo) (*sync.SyncService, error) {
	return sync.NewSyncService(
		eff.PostsDirAbs,
		eff.GenerateWorkdirAbs,
		eff.CleanArgs,
		eff.GenerateArgs,
		svcCtx.DB,
		svcCtx.Redis,
	)
}

// HandleSyncPosts godoc
//
//	@Summary		Hexo 文章全量同步
//	@Description	将已发布文章写入 Hexo source/_posts；rebuild=true 且在后台 Hexo 设置中配置了 clean_args/generate_args 时顺序执行 hexo clean 与 generate
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		eff, err := hexocfg.LoadEffective(ctx, svcCtx.DB, svcCtx.Config.HexoOptions)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		syncSvc, err := newHexoSyncServiceCtx(svcCtx, eff)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		syncCtx := c.Request.Context()
		res, err := syncSvc.SyncAll(syncCtx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		if req.Rebuild {
			hasRebuild := len(eff.CleanArgs) > 0 || len(eff.GenerateArgs) > 0
			if hasRebuild {
				genCtx, genCancel := context.WithTimeout(syncCtx, 15*time.Minute)
				defer genCancel()
				if err := syncSvc.RunHexoRebuild(genCtx); err != nil {
					common.Fail(c, http.StatusInternalServerError, err)
					return
				}
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
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		eff, err := hexocfg.LoadEffective(ctx, svcCtx.DB, svcCtx.Config.HexoOptions)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}

		syncSvc, err := newHexoSyncServiceCtx(svcCtx, eff)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		syncCtx := c.Request.Context()

		total, err := syncSvc.PublishedArticleCount(syncCtx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		local, err := syncSvc.LocalBeehivePostCount()
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		pending, err := syncSvc.PendingSync(syncCtx)
		if err != nil {
			common.Fail(c, http.StatusInternalServerError, err)
			return
		}
		last, err := syncSvc.LastSyncTime(syncCtx)
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
