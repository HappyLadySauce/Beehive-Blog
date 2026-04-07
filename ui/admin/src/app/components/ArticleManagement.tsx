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
    published: 'bg-green-100 text-green-800',
    draft: 'bg-gray-100 text-gray-800',
    scheduled: 'bg-blue-100 text-blue-800',
    archived: 'bg-yellow-100 text-yellow-800',
    private: 'bg-purple-100 text-purple-800',
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
          <FileText className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">文章</h2>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleBatchDelete}
            disabled={selectedArticles.length === 0}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-red-50 hover:text-red-600 hover:border-red-300 transition-colors disabled:opacity-50"
          >
            批量删除
          </button>
          <button
            onClick={() => navigate('/articles/create')}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors flex items-center gap-1.5"
          >
            <Plus className="w-4 h-4" />
            新建
          </button>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded">
        <div className="p-4 border-b border-gray-200">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="输入文章标题搜索"
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
              className="w-full pl-9 pr-4 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        <div className="p-4 flex items-center gap-3 flex-wrap border-b border-gray-200">
          <span className="text-sm text-gray-600">状态:</span>
          <CustomSelect
            value={selectedStatus}
            onChange={(v) => { setSelectedStatus(v); setPage(1); }}
            options={statusFilterOptions}
            className="w-[132px]"
            ariaLabel="文章状态筛选"
          />

          <span className="text-sm text-gray-600 ml-3">分类:</span>
          <CustomSelect
            value={selectedCategory}
            onChange={(v) => { setSelectedCategory(v); setPage(1); }}
            options={categoryFilterOptions}
            className="w-[132px]"
            ariaLabel="文章分类筛选"
          />

          <span className="text-sm text-gray-600 ml-3">标签:</span>
          <CustomSelect
            value={selectedTag}
            onChange={(v) => { setSelectedTag(v); setPage(1); }}
            options={tagFilterOptions}
            className="w-[132px]"
            ariaLabel="文章标签筛选"
          />

          <span className="text-sm text-gray-600 ml-3">排序:</span>
          <CustomSelect
            value={selectedSort}
            onChange={(v) => { setSelectedSort(v); setPage(1); }}
            options={sortFilterOptions}
            className="w-[132px]"
            ariaLabel="文章排序"
          />
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={articles.length > 0 && selectedArticles.length === articles.length}
                    onChange={toggleSelectAll}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                </th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">标题</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">分类 / 标签</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">状态</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">作者</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">浏览 / 评论</th>
                <th className="px-4 py-3 text-sm font-medium text-gray-600">发布时间</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500">加载中...</td>
                </tr>
              ) : articles.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500">暂无文章</td>
                </tr>
              ) : (
                articles.map((article) => (
                  <tr key={article.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedArticles.includes(article.id)}
                        onChange={() => toggleSelectArticle(article.id)}
                        className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-sm text-gray-900 font-medium max-w-md truncate">{article.title}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col gap-1">
                        <span className="text-xs text-gray-600">{article.category?.name || '-'}</span>
                        <div className="flex gap-1 flex-wrap">
                          {article.tags?.map((tag) => (
                            <span key={tag.id} className="px-1.5 py-0.5 bg-gray-100 text-gray-600 text-xs rounded">
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded text-xs ${statusColors[article.status] || 'bg-gray-100 text-gray-800'}`}>
                        {statusLabels[article.status] || article.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {article.author?.nickname || article.author?.username}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col text-xs text-gray-600">
                        <span>{article.viewCount} 浏览</span>
                        <span>{article.commentCount} 评论</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-600 whitespace-nowrap">
                      {article.publishedAt ? new Date(article.publishedAt).toLocaleString() : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => navigate(`/articles/edit/${article.id}`)}
                          className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors"
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
