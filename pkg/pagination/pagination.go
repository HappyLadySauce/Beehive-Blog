package pagination

// DefaultPageSize is the fallback page size when none is provided.
// DefaultPageSize 是未提供分页大小时的默认值。
const DefaultPageSize = 10

// NormalizeOffset clamps page to ≥1 and pageSize to DefaultPageSize when ≤0.
// NormalizeOffset 将 page 规范化为 ≥1，pageSize ≤0 时用 DefaultPageSize。
func NormalizeOffset(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	return page, pageSize
}
