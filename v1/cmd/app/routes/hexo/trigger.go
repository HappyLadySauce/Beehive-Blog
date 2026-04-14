// Package hexo 提供 DB 变更后的 Hexo 文件异步同步触发（文章 _posts、独立页面 beehive-pages）。
package hexo

import (
	"context"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/v1/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/hexocfg"
	"github.com/HappyLadySauce/Beehive-Blog/v1/pkg/sync"
	"k8s.io/klog/v2"
)

// MaybeSyncArticle 在后台开启 auto_sync 时异步同步单篇文章到 Hexo _posts。
func MaybeSyncArticle(svcCtx *svc.ServiceContext, articleID int64) {
	maybeRunSync(svcCtx, effRebuildAll, func(ctx context.Context, syncSvc *sync.SyncService) error {
		return syncSvc.SyncSingle(ctx, articleID)
	}, "SyncSingle", "articleID", articleID)
}

// MaybeDeletePost 异步删除 beehive 文章 md（仅需文章 ID）。
func MaybeDeletePost(svcCtx *svc.ServiceContext, articleID int64) {
	maybeRunSync(svcCtx, effNoRebuild, func(ctx context.Context, syncSvc *sync.SyncService) error {
		_ = ctx
		return syncSvc.DeletePostFile(&models.Article{ID: articleID})
	}, "DeletePostFile", "articleID", articleID)
}

// MaybeSyncPage 在后台开启 auto_sync 时异步同步独立页面到 source/beehive-pages。
func MaybeSyncPage(svcCtx *svc.ServiceContext, pageID int64) {
	maybeRunSync(svcCtx, effRebuildAll, func(ctx context.Context, syncSvc *sync.SyncService) error {
		return syncSvc.SyncSinglePage(ctx, pageID)
	}, "SyncSinglePage", "pageID", pageID)
}

// MaybeDeletePage 异步删除 beehive 独立页面目录。
func MaybeDeletePage(svcCtx *svc.ServiceContext, pageID int64) {
	maybeRunSync(svcCtx, effNoRebuild, func(ctx context.Context, syncSvc *sync.SyncService) error {
		_ = ctx
		return syncSvc.DeletePageFile(&models.Page{ID: pageID})
	}, "DeletePageFile", "pageID", pageID)
}

type effMode int

const (
	effNoRebuild effMode = iota
	effRebuildAll
)

// WriteHexoTaxonomyFile 将当前库中标签/分类映射写入 Hexo source/_data/beehive_taxonomy.json。
// 不依赖 auto_sync；标签/分类变更后调用，便于本地或下次 hexo g 使用。
func WriteHexoTaxonomyFile(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil {
		return
	}
	eff, err := hexocfg.LoadEffective(ctx, svcCtx.DB, svcCtx.Config.HexoOptions)
	if err != nil {
		klog.ErrorS(err, "[hexo] LoadEffective for taxonomy file")
		return
	}
	syncSvc, err := sync.NewSyncService(
		eff.PostsDirAbs,
		eff.GenerateWorkdirAbs,
		eff.CleanArgs,
		eff.GenerateArgs,
		strings.TrimSpace(svcCtx.Config.StorageOptions.BaseURL),
		svcCtx.DB,
		svcCtx.Redis,
	)
	if err != nil {
		klog.ErrorS(err, "[hexo] NewSyncService for taxonomy file")
		return
	}
	if err := syncSvc.WriteBeehiveTaxonomyJSON(ctx); err != nil {
		klog.ErrorS(err, "[hexo] WriteBeehiveTaxonomyJSON")
	}
}

func maybeRunSync(
	svcCtx *svc.ServiceContext,
	mode effMode,
	fn func(context.Context, *sync.SyncService) error,
	opName string,
	idKey string,
	id int64,
) {
	if svcCtx == nil {
		return
	}
	loadCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	eff, err := hexocfg.LoadEffective(loadCtx, svcCtx.DB, svcCtx.Config.HexoOptions)
	cancel()
	if err != nil {
		klog.ErrorS(err, "[hexo] LoadEffective failed", "op", opName)
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
		strings.TrimSpace(svcCtx.Config.StorageOptions.BaseURL),
		svcCtx.DB,
		svcCtx.Redis,
	)
	if err != nil {
		klog.ErrorS(err, "[hexo] failed to init sync service", "op", opName)
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := fn(ctx, syncSvc); err != nil {
			klog.ErrorS(err, "[hexo] sync op failed", "op", opName, idKey, id)
			return
		}
		if mode == effNoRebuild || !eff.RebuildAfterAutoSync {
			return
		}
		if len(eff.CleanArgs) == 0 && len(eff.GenerateArgs) == 0 {
			return
		}
		rebuildCtx, rebuildCancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer rebuildCancel()
		if err := syncSvc.RunHexoRebuild(rebuildCtx); err != nil {
			klog.ErrorS(err, "[hexo] RunHexoRebuild after op", "op", opName, idKey, id)
		}
	}()
}
