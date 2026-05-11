package settings

import (
	"context"
	"sync"

	"k8s.io/klog/v2"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// Provider holds a validated in-memory snapshot for hot reads.
// Provider 保存已校验的内存快照供热读。
type Provider struct {
	store *Store
	mu    sync.RWMutex
	snap  settingtypes.ApplicationSettings
	rev   int64
}

// NewProvider builds a Provider; call Refresh before serving traffic.
// NewProvider 构造 Provider；对外服务前应调用 Refresh。
func NewProvider(store *Store) *Provider {
	return &Provider{
		store: store,
		snap:  settingtypes.DefaultApplicationSettings(),
		rev:   0,
	}
}

// Current returns a copy of the last successful snapshot.
// Current 返回最近一次成功快照的拷贝。
func (p *Provider) Current() settingtypes.ApplicationSettings {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.snap
}

// CachedRevision returns the revision of the in-memory snapshot (0 before first Refresh).
// CachedRevision 返回内存快照对应的 revision（首次 Refresh 前为 0）。
func (p *Provider) CachedRevision() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rev
}

// Refresh reloads from the database, validates, and swaps the snapshot.
// On failure logs in English and keeps the previous snapshot.
// Refresh 从数据库加载、校验并替换快照；失败时记录英文日志并保留旧快照。
func (p *Provider) Refresh(ctx context.Context) error {
	s, rev, err := p.store.Load(ctx)
	if err != nil {
		klog.ErrorS(err, "Failed to refresh application settings snapshot")
		return err
	}
	p.mu.Lock()
	p.snap = s
	p.rev = rev
	p.mu.Unlock()
	return nil
}

// SaveAndRefresh persists settings and swaps the in-process snapshot.
// SaveAndRefresh 持久化设置并替换进程内快照。
func (p *Provider) SaveAndRefresh(ctx context.Context, next settingtypes.ApplicationSettings) error {
	rev, err := p.store.Save(ctx, next)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.snap = next
	p.rev = rev
	p.mu.Unlock()
	return nil
}
