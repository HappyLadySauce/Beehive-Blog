package settings

import (
	"sync"

	settingtypes "github.com/HappyLadySauce/Beehive-Blog/pkg/settings/types"
)

// Provider holds a validated in-memory snapshot for hot reads.
// Provider 保存已校验的内存快照供热读。
type Provider struct {
	mu   sync.RWMutex
	snap settingtypes.ApplicationSettings
	rev  int64
}

// NewProvider builds a Provider with the default in-memory snapshot.
// NewProvider 使用默认内存快照构造 Provider。
func NewProvider() *Provider {
	return &Provider{
		snap: settingtypes.DefaultApplicationSettings(),
		rev:  0,
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

// Replace swaps the in-memory snapshot after persistence has succeeded.
// Replace 在持久化成功后替换内存快照。
func (p *Provider) Replace(s settingtypes.ApplicationSettings, rev int64) {
	p.mu.Lock()
	p.snap = s
	p.rev = rev
	p.mu.Unlock()
}
