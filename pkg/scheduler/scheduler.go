// Package scheduler 提供可注册的周期性 Job Runner，供应用内定时业务（如定时发布）复用。
package scheduler

import (
	"context"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// Job 单次 tick 内执行的任务，仅依赖 context，业务依赖由闭包注入。
type Job func(ctx context.Context) error

// Runner 按固定间隔顺序执行已注册的 Job。
type Runner struct {
	interval time.Duration
	mu       sync.Mutex
	jobs     map[string]Job
}

// NewRunner 创建 Runner；interval 过短时回退为 1 分钟。
func NewRunner(interval time.Duration) *Runner {
	if interval < time.Second {
		interval = time.Minute
	}
	return &Runner{
		interval: interval,
		jobs:     make(map[string]Job),
	}
}

// Register 注册具名任务；同名覆盖。
func (r *Runner) Register(name string, fn Job) {
	if name == "" || fn == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[name] = fn
}

// Start 阻塞运行，直到 parent 取消；每个 Job 使用不超过 interval 的超时 context（至少 10s）。
func (r *Runner) Start(parent context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-parent.Done():
			return
		case <-ticker.C:
			r.runTick(parent)
		}
	}
}

// StartAsync 在 goroutine 中启动 Start。
func (r *Runner) StartAsync(parent context.Context) {
	go r.Start(parent)
}

func (r *Runner) runTick(parent context.Context) {
	r.mu.Lock()
	type named struct {
		name string
		fn   Job
	}
	list := make([]named, 0, len(r.jobs))
	for n, f := range r.jobs {
		list = append(list, named{name: n, fn: f})
	}
	r.mu.Unlock()

	jobTimeout := r.interval - time.Second
	if jobTimeout < 10*time.Second {
		jobTimeout = 10 * time.Second
	}

	for _, j := range list {
		select {
		case <-parent.Done():
			return
		default:
		}
		tickCtx, cancel := context.WithTimeout(parent, jobTimeout)
		err := j.fn(tickCtx)
		cancel()
		if err != nil {
			klog.ErrorS(err, "[scheduler] job failed", "job", j.name)
		}
	}
}
