package admin

import (
	"context"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/sync"
	"k8s.io/klog/v2"
)

// maybeHexoSyncSingle 在 hexo.auto_sync 为 true 时异步同步单篇文章到 Hexo _posts。
func maybeHexoSyncSingle(svcCtx *svc.ServiceContext, articleID int64) {
	if svcCtx == nil || !svcCtx.Config.HexoOptions.AutoSync {
		return
	}
	h := svcCtx.Config.HexoOptions
	syncSvc, err := sync.NewSyncService(h.PostsDir, h.GenerateWorkdir, h.GenerateArgs, svcCtx.DB, svcCtx.Redis)
	if err != nil {
		klog.ErrorS(err, "[hexo] failed to init sync service")
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := syncSvc.SyncSingle(ctx, articleID); err != nil {
			klog.ErrorS(err, "[hexo] SyncSingle failed", "articleID", articleID)
		}
	}()
}

// maybeHexoDeletePost 异步删除 beehive 文章 md（仅需文章 ID）。
func maybeHexoDeletePost(svcCtx *svc.ServiceContext, articleID int64) {
	if svcCtx == nil || !svcCtx.Config.HexoOptions.AutoSync {
		return
	}
	h := svcCtx.Config.HexoOptions
	syncSvc, err := sync.NewSyncService(h.PostsDir, h.GenerateWorkdir, h.GenerateArgs, svcCtx.DB, svcCtx.Redis)
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
