package settings

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"k8s.io/klog/v2"
)

// notifyChannelSettingRevision is the PostgreSQL NOTIFY/LISTEN channel name.
// notifyChannelSettingRevision 为 PostgreSQL NOTIFY/LISTEN 通道名。
const notifyChannelSettingRevision = "setting_revision"

// StartNotifyListener runs in the background: dedicated pgx connection, LISTEN, Refresh on NOTIFY until ctx is canceled.
// Reconnects with exponential backoff on disconnect or errors.
// StartNotifyListener 在后台运行：专用 pgx 连接 LISTEN，收到 NOTIFY 时 Refresh，直到 ctx 取消；断线或错误时指数退避重连。
func StartNotifyListener(ctx context.Context, connString string, p *Provider) {
	if p == nil || connString == "" {
		return
	}
	go func() {
		backoff := time.Second
		for {
			if ctx.Err() != nil {
				return
			}
			err := listenUntilError(ctx, connString, p)
			if ctx.Err() != nil {
				return
			}
			klog.ErrorS(err, "settings NOTIFY listener session ended; reconnecting", "backoff", backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
		}
	}()
}

func listenUntilError(ctx context.Context, connString string, p *Provider) error {
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return err
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(ctx) }()

	if _, err := conn.Exec(ctx, "LISTEN "+notifyChannelSettingRevision); err != nil {
		return err
	}
	klog.InfoS("Application settings NOTIFY listener subscribed", "channel", notifyChannelSettingRevision)

	for {
		_, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		if err := p.Refresh(ctx); err != nil {
			klog.ErrorS(err, "settings refresh after NOTIFY failed")
		}
	}
}
