import { useState, useEffect, useMemo } from 'react';
import { Search, Edit, Trash2, FileStack } from 'lucide-react';
import Pagination from '../../components/Pagination';
import CustomSelect from '../../components/CustomSelect';
import ConfirmModal from '../../components/ConfirmModal';
import {
  getPages,
  deletePage,
  AdminPageListItem,
  PageListQuery,
} from '../../api/page';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';

export default function PageManagement() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('all');
  const [selectedSort, setSelectedSort] = useState('newest');
  const [selectedRows, setSelectedRows] = useState<number[]>([]);
  const [rows, setRows] = useState<AdminPageListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

  const [confirmState, setConfirmState] = useState<{
    open: boolean;
    title: string;
    message: string;
    confirmLabel: string;
    confirmVariant: 'danger' | 'warning' | 'primary';
    onConfirm: () => void;
  }>({
    open: false,
    title: '',
    message: '',
    confirmLabel: '确认',
    confirmVariant: 'danger',
    onConfirm: () => {},
  });

  const showConfirm = (opts: Omit<typeof confirmState, 'open'>) =>
    setConfirmState({ open: true, ...opts });
  const hideConfirm = () => setConfirmState((s) => ({ ...s, open: false }));

  const fetchList = async () => {
    setLoading(true);
    try {
      const query: PageListQuery = {
        page,
        pageSize: 10,
        keyword: searchQuery || undefined,
        status: selectedStatus !== 'all' ? selectedStatus : undefined,
        sort: selectedSort,
      };
      const res = await getPages(query);
      if (res.code === 200) {
        setRows(res.data.list || []);
        setTotal(res.data.total || 0);
      } else {
        toast.error(res.message || '获取页面列表失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '请求页面列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchList();
  }, [page, searchQuery, selectedStatus, selectedSort]);

  const runDelete = async (id: number) => {
    try {
      const res = await deletePage(id);
      if (res.code === 200) {
        toast.success('已删除');
        setSelectedRows((prev) => prev.filter((x) => x !== id));
        void fetchList();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '删除请求失败');
    }
  };

  const handleDeleteOne = (id: number) => {
    showConfirm({
      title: '删除页面',
      message: '确定要删除该独立页面吗？可稍后在回收站恢复。',
      confirmLabel: '删除',
      confirmVariant: 'danger',
      onConfirm: () => {
        hideConfirm();
        void runDelete(id);
      },
    });
  };

  const runBatchDelete = async () => {
    if (!selectedRows.length) return;
    for (const id of selectedRows) {
      try {
        const res = await deletePage(id);
        if (res.code !== 200) {
          toast.error(res.message || `删除失败 id=${id}`);
          void fetchList();
          return;
        }
      } catch {
        toast.error('批量删除中断');
        void fetchList();
        return;
      }
    }
    toast.success(`已删除 ${selectedRows.length} 个页面`);
    setSelectedRows([]);
    void fetchList();
  };

  const handleBatchDelete = () => {
    if (!selectedRows.length) {
      toast.warning('请先选择页面');
      return;
    }
    showConfirm({
      title: '批量删除页面',
      message: `确定删除选中的 ${selectedRows.length} 个页面吗？`,
      confirmLabel: '删除',
      confirmVariant: 'danger',
      onConfirm: () => {
        hideConfirm();
        void runBatchDelete();
      },
    });
  };

  const statusColors: Record<string, string> = {
    published: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
    draft: 'bg-muted text-foreground',
    archived: 'bg-yellow-500/15 text-yellow-800 dark:text-yellow-400',
    private: 'bg-purple-500/15 text-purple-800 dark:text-purple-300',
  };

  const statusLabels: Record<string, string> = {
    published: '已发布',
    draft: '草稿',
    archived: '已归档',
    private: '私密',
  };

  const toggleSelectAll = () => {
    if (selectedRows.length === rows.length && rows.length > 0) {
      setSelectedRows([]);
    } else {
      setSelectedRows(rows.map((r) => r.id));
    }
  };

  const toggleSelect = (id: number) => {
    setSelectedRows((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  };

  const statusFilterOptions = useMemo(
    () => [
      { value: 'all', label: '全部' },
      { value: 'published', label: '已发布' },
      { value: 'draft', label: '草稿' },
      { value: 'archived', label: '已归档' },
      { value: 'private', label: '私密' },
    ],
    [],
  );

  const sortFilterOptions = useMemo(
    () => [
      { value: 'newest', label: '最新更新' },
      { value: 'oldest', label: '最早创建' },
      { value: 'popular', label: '浏览最多' },
    ],
    [],
  );

  const batchToolbarBtnClass =
    'rounded border border-border bg-background px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent';

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <FileStack className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
          <h1 className="text-xl font-semibold tracking-tight text-foreground">独立页面</h1>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => navigate('/pages/trash')}
            className="rounded-md border border-border bg-background px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent"
          >
            回收站
          </button>
          <button
            type="button"
            onClick={() => navigate('/pages/create')}
            className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
          >
            新建
          </button>
        </div>
      </header>

      <div className="admin-card admin-card-glass rounded border">
        <div className="border-b border-border p-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 transform text-muted-foreground" />
            <input
              type="text"
              placeholder="输入页面标题搜索"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setPage(1);
              }}
              className="w-full rounded border border-border bg-input-background py-2 pl-9 pr-4 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
            />
          </div>
        </div>

        <div className="flex w-full flex-wrap items-center gap-3 border-b border-border p-4">
          {selectedRows.length > 0 && (
            <div className="flex flex-wrap items-center gap-2 shrink-0">
              <button
                type="button"
                className="rounded bg-red-600 px-3 py-1.5 text-sm text-white transition-colors hover:bg-red-700"
                onClick={handleBatchDelete}
              >
                删除 ({selectedRows.length})
              </button>
            </div>
          )}
          <div className="flex min-w-0 flex-1 flex-wrap items-center gap-3">
            <span className="text-sm text-muted-foreground">状态:</span>
            <CustomSelect
              value={selectedStatus}
              onChange={(v) => {
                setSelectedStatus(v);
                setPage(1);
              }}
              options={statusFilterOptions}
              className="w-[132px]"
              ariaLabel="页面状态筛选"
            />
            <span className="ml-3 text-sm text-muted-foreground">排序:</span>
            <CustomSelect
              value={selectedSort}
              onChange={(v) => {
                setSelectedSort(v);
                setPage(1);
              }}
              options={sortFilterOptions}
              className="w-[132px]"
              ariaLabel="页面排序"
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="admin-table w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={rows.length > 0 && selectedRows.length === rows.length}
                    onChange={toggleSelectAll}
                    className="h-4 w-4 rounded border-border text-primary focus:ring-ring"
                  />
                </th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">标题</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">路径</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">状态</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">菜单</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">浏览</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">更新时间</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-muted-foreground">
                    加载中...
                  </td>
                </tr>
              ) : rows.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-muted-foreground">
                    暂无页面
                  </td>
                </tr>
              ) : (
                rows.map((row) => (
                  <tr key={row.id} className="transition-colors hover:bg-muted/50">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedRows.includes(row.id)}
                        onChange={() => toggleSelect(row.id)}
                        className="h-4 w-4 rounded border-border text-primary focus:ring-ring"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="max-w-md truncate text-sm font-medium text-foreground">{row.title}</div>
                    </td>
                    <td className="px-4 py-3">
                      <code className="text-xs text-muted-foreground">/{row.slug}/</code>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded px-2 py-1 text-xs ${statusColors[row.status] || 'bg-muted text-foreground'}`}
                      >
                        {statusLabels[row.status] || row.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">{row.isInMenu ? '是' : '否'}</td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">{row.viewCount}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-xs text-muted-foreground">
                      {row.updatedAt ? new Date(row.updatedAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          type="button"
                          onClick={() => navigate(`/pages/edit/${row.id}`)}
                          className={`${batchToolbarBtnClass} p-1.5`}
                          title="编辑"
                        >
                          <Edit className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDeleteOne(row.id)}
                          className={`${batchToolbarBtnClass} p-1.5 text-red-600 hover:bg-red-500/10`}
                          title="删除"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <Pagination total={total} page={page} pageSize={10} onPageChange={setPage} unit="项结果" />
      </div>

      <ConfirmModal
        open={confirmState.open}
        title={confirmState.title}
        message={confirmState.message}
        confirmLabel={confirmState.confirmLabel}
        confirmVariant={confirmState.confirmVariant}
        onConfirm={confirmState.onConfirm}
        onCancel={hideConfirm}
      />
    </div>
  );
}
