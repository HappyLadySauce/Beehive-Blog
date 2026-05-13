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

// RefreshFunc reloads application settings after a database notification.
// RefreshFunc 在数据库通知后重新加载应用设置。
type RefreshFunc func(context.Context) error

// StartNotifyListener runs in the background: dedicated pgx connection, LISTEN, refresh callback on NOTIFY until ctx is canceled.
// Reconnects with exponential backoff on disconnect or errors.
// StartNotifyListener 在后台运行：专用 pgx 连接 LISTEN，收到 NOTIFY 时调用 refresh 回调，直到 ctx 取消；断线或错误时指数退避重连。
func StartNotifyListener(ctx context.Context, connString string, refresh RefreshFunc) {
	if refresh == nil || connString == "" {
		return
	}
	go func() {
		backoff := time.Second
		for {
			if ctx.Err() != nil {
				return
			}
			err := listenUntilError(ctx, connString, refresh)
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

func listenUntilError(ctx context.Context, connString string, refresh RefreshFunc) error {
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
		if err := refresh(ctx); err != nil {
			klog.ErrorS(err, "settings refresh after NOTIFY failed")
		}
	}
}
