import React, { useState, useEffect } from 'react';
import { Search, Plus, Edit, Trash2, Eye, FileText } from 'lucide-react';
import Pagination from '../../components/Pagination';
import CustomSelect from '../../components/CustomSelect';
import ConfirmModal from '../../components/ConfirmModal';
import { getArticles, deleteArticle, batchOperateArticles, AdminArticleListItem, ArticleListQuery } from '../../api/article';
import { getCategories, getTags, CategoryBrief, TagListItem } from '../../api/taxonomy';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';

export default function ArticleManagement() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('all');
  const [selectedSort, setSelectedSort] = useState('newest');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [selectedTag, setSelectedTag] = useState('all');
  const [selectedArticles, setSelectedArticles] = useState<number[]>([]);
  const [articles, setArticles] = useState<AdminArticleListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<CategoryBrief[]>([]);
  const [tags, setTags] = useState<TagListItem[]>([]);

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

  const showConfirm = (
    opts: Omit<typeof confirmState, 'open'>,
  ) => setConfirmState({ open: true, ...opts });
  const hideConfirm = () => setConfirmState((s) => ({ ...s, open: false }));

  const fetchArticles = async () => {
    setLoading(true);
    try {
      const query: ArticleListQuery = {
        page,
        pageSize: 10,
        keyword: searchQuery || undefined,
        status: selectedStatus !== 'all' ? selectedStatus : undefined,
        category: selectedCategory !== 'all' ? selectedCategory : undefined,
        tag: selectedTag !== 'all' ? selectedTag : undefined,
        sort: selectedSort,
      };
      const res = await getArticles(query);
      if (res.code === 200) {
        setArticles(res.data.list || []);
        setTotal(res.data.total || 0);
      } else {
        toast.error(res.message || '获取文章列表失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求文章列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchFilters = async () => {
    try {
      const [catRes, tagRes] = await Promise.all([
        getCategories({ pageSize: 200 }),
        getTags({ pageSize: 200 }),
      ]);
      if (catRes.code === 200) setCategories(catRes.data.list || []);
      if (tagRes.code === 200) setTags(tagRes.data.list || []);
    } catch {
      // 筛选器加载失败不阻断主流程
    }
  };

  useEffect(() => {
    fetchFilters();
  }, []);

  useEffect(() => {
    fetchArticles();
  }, [page, searchQuery, selectedStatus, selectedSort, selectedCategory, selectedTag]);

  const runBatchDelete = async () => {
    try {
      const res = await batchOperateArticles({
        action: 'delete',
        ids: selectedArticles,
      });
      if (res.code === 200) {
        toast.success(`成功删除 ${res.data.affected} 篇文章`);
        setSelectedArticles([]);
        fetchArticles();
      } else {
        toast.error(res.message || '批量删除失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '批量删除请求失败');
    }
  };

  const handleBatchDelete = () => {
    if (!selectedArticles.length) {
      toast.warning('请先选择要删除的文章');
      return;
    }
    showConfirm({
      title: '批量删除文章',
      message: `确定要删除选中的 ${selectedArticles.length} 篇文章吗？`,
      confirmLabel: '删除',
      confirmVariant: 'danger',
      onConfirm: () => {
        hideConfirm();
        void runBatchDelete();
      },
    });
  };

  const runDeleteArticle = async (id: number) => {
    try {
      const res = await deleteArticle(id);
      if (res.code === 200) {
        toast.success('删除成功');
        setSelectedArticles((prev) => prev.filter((a) => a !== id));
        fetchArticles();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '删除请求失败');
    }
  };

  const handleDelete = (id: number) => {
    showConfirm({
      title: '删除文章',
      message: '确定要删除该文章吗？',
      confirmLabel: '删除',
      confirmVariant: 'danger',
      onConfirm: () => {
        hideConfirm();
        void runDeleteArticle(id);
      },
    });
  };

  const statusColors: Record<string, string> = {
    published: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
    draft: 'bg-muted text-foreground',
    scheduled: 'bg-primary/15 text-primary',
    archived: 'bg-yellow-500/15 text-yellow-800 dark:text-yellow-400',
    private: 'bg-purple-500/15 text-purple-800 dark:text-purple-300',
  };

  const statusLabels: Record<string, string> = {
    published: '已发布',
    draft: '草稿',
    scheduled: '定时发布',
    archived: '已归档',
    private: '私密',
  };

  const toggleSelectAll = () => {
    if (selectedArticles.length === articles.length && articles.length > 0) {
      setSelectedArticles([]);
    } else {
      setSelectedArticles(articles.map(a => a.id));
    }
  };

  const toggleSelectArticle = (id: number) => {
    setSelectedArticles(prev =>
      prev.includes(id) ? prev.filter(a => a !== id) : [...prev, id]
    );
  };

  const statusFilterOptions = [
    { value: 'all', label: '全部' },
    { value: 'published', label: '已发布' },
    { value: 'draft', label: '草稿' },
    { value: 'scheduled', label: '定时发布' },
    { value: 'archived', label: '已归档' },
    { value: 'private', label: '私密' },
  ];
  const categoryFilterOptions = [
    { value: 'all', label: '全部' },
    ...categories.map((cat) => ({ value: cat.slug, label: cat.name })),
  ];
  const tagFilterOptions = [
    { value: 'all', label: '全部' },
    ...tags.map((tag) => ({ value: tag.slug, label: tag.name })),
  ];
  const sortFilterOptions = [
    { value: 'newest', label: '最新发布' },
    { value: 'oldest', label: '最早发布' },
    { value: 'popular', label: '最受欢迎' },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FileText className="w-5 h-5 text-muted-foreground" />
          <h2 className="text-lg font-medium text-foreground">文章</h2>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleBatchDelete}
            disabled={selectedArticles.length === 0}
            className="px-3 py-1.5 text-sm border border-border rounded bg-background hover:bg-red-500/10 hover:text-red-600 hover:border-red-500/40 transition-colors disabled:opacity-50"
          >
            批量删除
          </button>
          <button
            onClick={() => navigate('/articles/create')}
            className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors flex items-center gap-1.5"
          >
            <Plus className="w-4 h-4" />
            新建
          </button>
        </div>
      </div>

      <div className="admin-card admin-card-glass rounded border">
        <div className="p-4 border-b border-border">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="输入文章标题搜索"
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
              className="w-full pl-9 pr-4 py-2 text-sm border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent"
            />
          </div>
        </div>

        <div className="p-4 flex items-center gap-3 flex-wrap border-b border-border">
          <span className="text-sm text-muted-foreground">状态:</span>
          <CustomSelect
            value={selectedStatus}
            onChange={(v) => { setSelectedStatus(v); setPage(1); }}
            options={statusFilterOptions}
            className="w-[132px]"
            ariaLabel="文章状态筛选"
          />

          <span className="text-sm text-muted-foreground ml-3">分类:</span>
          <CustomSelect
            value={selectedCategory}
            onChange={(v) => { setSelectedCategory(v); setPage(1); }}
            options={categoryFilterOptions}
            className="w-[132px]"
            ariaLabel="文章分类筛选"
          />

          <span className="text-sm text-muted-foreground ml-3">标签:</span>
          <CustomSelect
            value={selectedTag}
            onChange={(v) => { setSelectedTag(v); setPage(1); }}
            options={tagFilterOptions}
            className="w-[132px]"
            ariaLabel="文章标签筛选"
          />

          <span className="text-sm text-muted-foreground ml-3">排序:</span>
          <CustomSelect
            value={selectedSort}
            onChange={(v) => { setSelectedSort(v); setPage(1); }}
            options={sortFilterOptions}
            className="w-[132px]"
            ariaLabel="文章排序"
          />
        </div>

        <div className="overflow-x-auto">
          <table className="admin-table w-full border-collapse text-left">
            <thead>
              <tr className="bg-muted/50 border-b border-border">
                <th className="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={articles.length > 0 && selectedArticles.length === articles.length}
                    onChange={toggleSelectAll}
                    className="w-4 h-4 text-primary border-border rounded focus:ring-ring"
                  />
                </th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">标题</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">分类 / 标签</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">状态</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">作者</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">浏览 / 评论</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">发布时间</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-muted-foreground">加载中...</td>
                </tr>
              ) : articles.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-muted-foreground">暂无文章</td>
                </tr>
              ) : (
                articles.map((article) => (
                  <tr key={article.id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedArticles.includes(article.id)}
                        onChange={() => toggleSelectArticle(article.id)}
                        className="w-4 h-4 text-primary border-border rounded focus:ring-ring"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-sm text-foreground font-medium max-w-md truncate">{article.title}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col gap-1">
                        <span className="text-xs text-muted-foreground">{article.category?.name || '-'}</span>
                        <div className="flex gap-1 flex-wrap">
                          {article.tags?.map((tag) => (
                            <span key={tag.id} className="px-1.5 py-0.5 bg-muted text-muted-foreground text-xs rounded">
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded text-xs ${statusColors[article.status] || 'bg-muted text-foreground'}`}>
                        {statusLabels[article.status] || article.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">
                      {article.author?.nickname || article.author?.username}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col text-xs text-muted-foreground">
                        <span>{article.viewCount} 浏览</span>
                        <span>{article.commentCount} 评论</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground whitespace-nowrap">
                      {article.publishedAt ? new Date(article.publishedAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => navigate(`/articles/edit/${article.id}`)}
                          className="p-1.5 text-primary hover:bg-primary/10 rounded transition-colors"
                          title="编辑"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(article.id)}
                          className="p-1.5 text-red-600 hover:bg-red-50 rounded transition-colors"
                          title="删除"
                        >
                          <Trash2 className="w-4 h-4" />
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
