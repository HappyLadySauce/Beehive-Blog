import { useState, useEffect } from 'react';
import { Search, Filter, Plus, MoreVertical, Edit, Trash2, Eye, FileText, ChevronDown } from 'lucide-react';
import { getArticles, batchOperateArticles, ArticleItem, ArticleListQuery } from '../../api/article';
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
  const [articles, setArticles] = useState<ArticleItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

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
        setArticles(res.data.items || []);
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

  useEffect(() => {
    fetchArticles();
  }, [page, searchQuery, selectedStatus, selectedSort, selectedCategory, selectedTag]);

  const handleBatchDelete = async () => {
    if (!selectedArticles.length) {
      toast.warning('请先选择要删除的文章');
      return;
    }
    if (!window.confirm(`确定要删除选中的 ${selectedArticles.length} 篇文章吗？`)) return;
    
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

  const categories = ['前端开发', '后端开发', '编程语言', '网络安全', 'DevOps'];
  const tags = ['React', 'JavaScript', 'TypeScript', 'CSS', 'Tailwind', 'Node.js', 'Performance', 'Security', 'Web', 'Docker', 'Container'];

  const statusColors = {
    published: 'bg-green-100 text-green-800',
    draft: 'bg-gray-100 text-gray-800',
    scheduled: 'bg-blue-100 text-blue-800',
    archived: 'bg-yellow-100 text-yellow-800',
    private: 'bg-purple-100 text-purple-800',
  };

  const statusLabels = {
    published: '已发布',
    draft: '草稿',
    scheduled: '定时发布',
    archived: '已归档',
    private: '私密',
  };

  const toggleSelectAll = () => {
    if (selectedArticles.length === articles.length) {
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

  return (
    <div className="space-y-4">
      {/* 页面标题 */}
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

      {/* 搜索和筛选栏 */}
      <div className="bg-white border border-gray-200 rounded">
        <div className="p-4 border-b border-gray-200">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="输入文章标题搜索"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-4 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        <div className="p-4 flex items-center gap-3 flex-wrap border-b border-gray-200">
          <span className="text-sm text-gray-600">状态:</span>
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">全部</option>
            <option value="published">已发布</option>
            <option value="draft">草稿</option>
            <option value="scheduled">定时发布</option>
            <option value="archived">已归档</option>
            <option value="private">私密</option>
          </select>

          <span className="text-sm text-gray-600 ml-3">分类:</span>
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">全部</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>

          <span className="text-sm text-gray-600 ml-3">标签:</span>
          <select
            value={selectedTag}
            onChange={(e) => setSelectedTag(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">全部</option>
            {tags.map((tag) => (
              <option key={tag} value={tag}>{tag}</option>
            ))}
          </select>

          <span className="text-sm text-gray-600 ml-3">排序:</span>
          <select
            value={selectedSort}
            onChange={(e) => setSelectedSort(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="newest">最新发布</option>
            <option value="oldest">最早发布</option>
            <option value="popular">最受欢迎</option>
          </select>
        </div>

        {/* 文章表格 */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50">
                <th className="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedArticles.length === articles.length}
                    onChange={toggleSelectAll}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">标题</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">分类 / 标签</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">作者</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">浏览 / 评论</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-600">发布时间</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-gray-600">操作</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500">
                    加载中...
                  </td>
                </tr>
              ) : articles.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500">
                    暂无文章
                  </td>
                </tr>
              ) : (
                articles.map((article) => (
                  <tr key={article.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
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
                      <span className={`px-2 py-1 rounded text-xs ${statusColors[article.status as keyof typeof statusColors] || 'bg-gray-100 text-gray-800'}`}>
                        {statusLabels[article.status as keyof typeof statusLabels] || article.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{article.author?.nickname || article.author?.username}</td>
                    <td className="px-4 py-3">
                      <div className="flex flex-col text-xs text-gray-600">
                        <span>{article.viewCount} 浏览</span>
                        <span>{article.commentCount} 评论</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-600 whitespace-nowrap">{new Date(article.publishedAt || article.createdAt).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button className="p-1.5 text-gray-600 hover:bg-gray-100 rounded transition-colors" title="查看">
                          <Eye className="w-4 h-4" />
                        </button>
                        <button 
                          onClick={() => navigate(`/articles/edit/${article.id}`)}
                          className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors" title="编辑"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button className="p-1.5 text-gray-600 hover:bg-gray-100 rounded transition-colors" title="更多">
                          <MoreVertical className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* 分页 */}
        <div className="p-4 flex items-center justify-between border-t border-gray-200">
          <div className="text-sm text-gray-600">
            共 {total} 项结果
          </div>
          <div className="flex items-center gap-1">
            <button 
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors disabled:opacity-50"
            >
              上一页
            </button>
            <button className="px-3 py-1 text-sm bg-blue-600 text-white rounded">{page}</button>
            <button 
              onClick={() => setPage(p => p + 1)}
              disabled={page * 10 >= total}
              className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors disabled:opacity-50"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
