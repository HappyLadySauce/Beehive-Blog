package sync

// SyncAction 单篇文章同步意图，用于统计。
type SyncAction string

const (
	SyncActionCreate SyncAction = "created"
	SyncActionUpdate SyncAction = "updated"
	SyncActionDelete SyncAction = "deleted"
)

// SyncResult 一次同步任务的汇总结果。
type SyncResult struct {
	Total   int
	Created int
	Updated int
	Deleted int
	Files   []string
}
