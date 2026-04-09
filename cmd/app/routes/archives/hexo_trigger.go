package archives

import (
	"context"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/hexocfg"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/sync"
	"k8s.io/klog/v2"
)

// maybeHexoSyncSingle 在后台 Hexo 设置中开启 auto_sync 时异步同步单篇文章到 Hexo _posts。
func maybeHexoSyncSingle(svcCtx *svc.ServiceContext, articleID int64) {
	if svcCtx == nil {
		return
	}
	loadCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	eff, err := hexocfg.LoadEffective(loadCtx, svcCtx.DB, svcCtx.Config.HexoOptions)
	cancel()
	if err != nil {
		klog.ErrorS(err, "[hexo] LoadEffective failed")
		return
	}
	if !eff.AutoSync {
		return
	}
	syncSvc, err := sync.NewSyncService(
		eff.PostsDirAbs,
		eff.GenerateWorkdirAbs,
		eff.CleanArgs,
		eff.GenerateArgs,
		svcCtx.DB,
		svcCtx.Redis,
	)
	if err != nil {
		klog.ErrorS(err, "[hexo] failed to init sync service")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := syncSvc.SyncSingle(ctx, articleID); err != nil {
			klog.ErrorS(err, "[hexo] SyncSingle failed", "articleID", articleID)
			return
		}
		if !eff.RebuildAfterAutoSync {
			return
		}
		if len(eff.CleanArgs) == 0 && len(eff.GenerateArgs) == 0 {
			return
		}
		rebuildCtx, rebuildCancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer rebuildCancel()
		if err := syncSvc.RunHexoRebuild(rebuildCtx); err != nil {
			klog.ErrorS(err, "[hexo] RunHexoRebuild after SyncSingle", "articleID", articleID)
		}
	}()
}

// maybeHexoDeletePost 异步删除 beehive 文章 md（仅需文章 ID）。
func maybeHexoDeletePost(svcCtx *svc.ServiceContext, articleID int64) {
	if svcCtx == nil {
		return
	}
	loadCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	eff, err := hexocfg.LoadEffective(loadCtx, svcCtx.DB, svcCtx.Config.HexoOptions)
	cancel()
	if err != nil {
		klog.ErrorS(err, "[hexo] LoadEffective failed for delete")
		return
	}
	if !eff.AutoSync {
		return
	}
	syncSvc, err := sync.NewSyncService(
		eff.PostsDirAbs,
		eff.GenerateWorkdirAbs,
		eff.CleanArgs,
		eff.GenerateArgs,
		svcCtx.DB,
		svcCtx.Redis,
	)
	if err != nil {
		klog.ErrorS(err, "[hexo] failed to init sync service for delete")
		return
	}
	go func() {
		a := &models.Article{ID: articleID}
		if err := syncSvc.DeletePostFile(a); err != nil {
			klog.ErrorS(err, "[hexo] DeletePostFile failed", "articleID", articleID)
		}
	}()
}
