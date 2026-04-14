interface PaginationProps {
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  /** 「共 N 条」的单位词，默认「条」*/
  unit?: string;
}

/**
 * 通用分页组件，展示总数与上一页/当前页/下一页按钮。
 *
 * @example
 * <Pagination total={total} page={page} pageSize={20} onPageChange={setPage} />
 */
export default function Pagination({
  total,
  page,
  pageSize,
  onPageChange,
  unit = '条',
}: PaginationProps) {
  const hasPrev = page > 1;
  const hasNext = page * pageSize < total;

  return (
    <div className="flex items-center justify-between border-t border-border p-[clamp(0.75rem,0.4vw+0.6rem,1rem)]">
      <div className="text-[clamp(0.86rem,0.12vw+0.82rem,0.95rem)] text-muted-foreground">
        共 {total} {unit}
      </div>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={!hasPrev}
          className="admin-control-sm rounded border border-border bg-background px-3 text-[clamp(0.84rem,0.1vw+0.8rem,0.92rem)] hover:bg-accent transition-colors disabled:opacity-50"
        >
          上一页
        </button>
        <button className="admin-control-sm cursor-default rounded bg-primary px-3 text-[clamp(0.84rem,0.1vw+0.8rem,0.92rem)] text-primary-foreground">
          {page}
        </button>
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={!hasNext}
          className="admin-control-sm rounded border border-border bg-background px-3 text-[clamp(0.84rem,0.1vw+0.8rem,0.92rem)] hover:bg-accent transition-colors disabled:opacity-50"
        >
          下一页
        </button>
      </div>
    </div>
  );
}
