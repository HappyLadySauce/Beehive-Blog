import { useEffect, useMemo, useState } from 'react';
import { Search, RotateCcw, Trash2 } from 'lucide-react';
import Pagination from '../../components/Pagination';
import CustomSelect from '../../components/CustomSelect';
import ConfirmModal from '../../components/ConfirmModal';
import {
  getTrashedArticles,
  restoreArticle,
  permanentDeleteArticle,
  AdminArticleListItem,
  ArticleListQuery,
} from '../../api/article';
import { toast } from 'sonner';

export default function ArticleTrash() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSort, setSelectedSort] = useState('newest');
  const [articles, setArticles] = useState<AdminArticleListItem[]>([]);
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

  const sortOptions = useMemo(
    () => [
      { value: 'newest', label: '最新删除' },
      { value: 'oldest', label: '最早删除' },
      { value: 'popular', label: '最受欢迎' },
    ],
    [],
  );

  const fetchList = async () => {
    setLoading(true);
    try {
      const query: ArticleListQuery = {
        page,
        pageSize: 10,
        keyword: searchQuery || undefined,
        sort: selectedSort,
      };
      const res = await getTrashedArticles(query);
      if (res.code === 200) {
        setArticles(res.data.list || []);
        setTotal(res.data.total || 0);
      } else {
        toast.error(res.message || '获取回收站失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '请求回收站列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchList();
  }, [page, searchQuery, selectedSort]);

  const runRestore = async (id: number) => {
    try {
      const res = await restoreArticle(id);
      if (res.code === 200) {
        toast.success('已恢复');
        void fetchList();
      } else {
        toast.error(res.message || '恢复失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '恢复请求失败');
    }
  };

  const runPermanent = async (id: number) => {
    try {
      const res = await permanentDeleteArticle(id);
      if (res.code === 200) {
        toast.success('已彻底删除');
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

  return (
    <div className="space-y-4">
      <div className="admin-card admin-card-glass rounded border">
        <div className="border-b border-border p-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 transform text-muted-foreground" />
            <input
              type="text"
              placeholder="搜索已删除文章标题或摘要"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setPage(1);
              }}
              className="w-full rounded border border-border bg-input-background py-2 pl-9 pr-4 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
            />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3 border-b border-border p-4">
          <span className="text-sm text-muted-foreground">排序:</span>
          <CustomSelect
            value={selectedSort}
            onChange={(v) => {
              setSelectedSort(v);
              setPage(1);
            }}
            options={sortOptions}
            className="w-[132px]"
            ariaLabel="回收站排序"
          />
        </div>

        <div className="overflow-x-auto">
          <table className="admin-table w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">标题</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">分类 / 标签</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">作者</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">更新时间</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-muted-foreground">
                    加载中...
                  </td>
                </tr>
              ) : articles.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-10 text-center text-sm text-muted-foreground">
                    回收站为空。删除的文章会出现在这里，可恢复或彻底删除。
                  </td>
                </tr>
              ) : (
                articles.map((article) => (
                  <tr key={article.id} className="transition-colors hover:bg-muted/50">
                    <td className="px-4 py-3">
                      <div className="max-w-md truncate text-sm font-medium text-foreground">{article.title}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col gap-1">
                        <span className="text-xs text-muted-foreground">{article.category?.name || '-'}</span>
                        <div className="flex flex-wrap gap-1">
                          {article.tags?.map((tag) => (
                            <span
                              key={tag.id}
                              className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground"
                            >
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">
                      {article.author?.nickname || article.author?.username}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-xs text-muted-foreground">
                      {article.updatedAt ? new Date(article.updatedAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          type="button"
                          onClick={() => {
                            showConfirm({
                              title: '恢复文章',
                              message: '确定将此文恢复到文章列表吗？',
                              confirmLabel: '恢复',
                              confirmVariant: 'primary',
                              onConfirm: () => {
                                hideConfirm();
                                void runRestore(article.id);
                              },
                            });
                          }}
                          className="rounded p-1.5 text-primary transition-colors hover:bg-primary/10"
                          title="恢复"
                        >
                          <RotateCcw className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => {
                            showConfirm({
                              title: '彻底删除',
                              message: '彻底删除后无法恢复，确定继续吗？',
                              confirmLabel: '彻底删除',
                              confirmVariant: 'danger',
                              onConfirm: () => {
                                hideConfirm();
                                void runPermanent(article.id);
                              },
                            });
                          }}
                          className="rounded p-1.5 text-red-600 transition-colors hover:bg-red-50"
                          title="彻底删除"
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
        onCancel={hideConfirm}
        onConfirm={() => confirmState.onConfirm()}
      />
    </div>
  );
}
