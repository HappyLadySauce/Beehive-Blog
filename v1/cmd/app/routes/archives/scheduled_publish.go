package archives

import (
	"context"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	routehexo "github.com/HappyLadySauce/Beehive-Blog/cmd/app/routes/hexo"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/svc"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/scheduler"
)

// RegisterScheduledPublishJob 将「到期定时文章改为已发布」注册到 Runner。
func RegisterScheduledPublishJob(r *scheduler.Runner, svcCtx *svc.ServiceContext) {
	r.Register("promote_scheduled_articles", func(ctx context.Context) error {
		return PromoteDueScheduledArticles(ctx, svcCtx)
	})
}

// PromoteDueScheduledArticles 将 published_at 已到的 scheduled 文章更新为 published，并触发 Hexo 同步。
func PromoteDueScheduledArticles(ctx context.Context, svcCtx *svc.ServiceContext) error {
	if svcCtx == nil || svcCtx.DB == nil {
		return nil
	}
	now := time.Now()
	var ids []int64
	err := svcCtx.DB.WithContext(ctx).Model(&models.Article{}).
		Where("status = ? AND deleted_at IS NULL AND published_at IS NOT NULL AND published_at <= ?", models.ArticleStatusScheduled, now).
		Pluck("id", &ids).Error
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	res := svcCtx.DB.WithContext(ctx).Model(&models.Article{}).
		Where("id IN ? AND status = ? AND deleted_at IS NULL", ids, models.ArticleStatusScheduled).
		Update("status", models.ArticleStatusPublished)
	if res.Error != nil {
		return res.Error
	}
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		routehexo.MaybeSyncArticle(svcCtx, id)
	}
	return nil
}
