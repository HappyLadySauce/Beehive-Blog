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
    <div className="p-4 flex items-center justify-between border-t border-gray-200">
      <div className="text-sm text-gray-600">
        共 {total} {unit}
      </div>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={!hasPrev}
          className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors disabled:opacity-50"
        >
          上一页
        </button>
        <button className="px-3 py-1 text-sm bg-blue-600 text-white rounded cursor-default">
          {page}
        </button>
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={!hasNext}
          className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors disabled:opacity-50"
        >
          下一页
        </button>
      </div>
    </div>
  );
}
