import { useState, useEffect } from 'react';
import { getComments, updateCommentStatus, AdminCommentItem, AdminCommentListQuery } from '../../api/comment';
import { toast } from 'sonner';
import { MessageSquare, Check, X, ShieldAlert, Search } from 'lucide-react';
import Pagination from '../../components/Pagination';
import CustomSelect from '../../components/CustomSelect';

const statusColors: Record<string, string> = {
  approved: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  pending: 'bg-yellow-500/15 text-yellow-800 dark:text-yellow-400',
  rejected: 'bg-red-500/15 text-red-700 dark:text-red-400',
  spam: 'bg-orange-500/15 text-orange-800 dark:text-orange-400',
};

const statusLabels: Record<string, string> = {
  approved: '已通过',
  pending: '待审核',
  rejected: '已拒绝',
  spam: '垃圾',
};

const commentFilterOptions = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '待审核' },
  { value: 'approved', label: '已通过' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'spam', label: '垃圾' },
];

export default function Comments() {
  const [comments, setComments] = useState<AdminCommentItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [filterStatus, setFilterStatus] = useState('');
  const [keyword, setKeyword] = useState('');

  const fetchComments = async () => {
    setLoading(true);
    try {
      const query: AdminCommentListQuery = {
        page,
        pageSize: 20,
        status: filterStatus || undefined,
        keyword: keyword || undefined,
      };
      const res = await getComments(query);
      if (res.code === 200) {
        setComments(res.data.items || []);
        setTotal(res.data.total || 0);
      } else {
        toast.error(res.message || '获取评论失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求评论失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchComments();
  }, [page, filterStatus, keyword]);

  const handleStatusChange = async (id: number, status: string) => {
    try {
      const res = await updateCommentStatus(id, status);
      if (res.code === 200) {
        toast.success('状态更新成功');
        fetchComments();
      } else {
        toast.error(res.message || '状态更新失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '状态更新请求失败');
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MessageSquare className="w-5 h-5 text-muted-foreground" />
          <h2 className="text-lg font-medium text-foreground">评论管理</h2>
        </div>
      </div>

      <div className="admin-card admin-card-glass rounded border">
        <div className="p-4 border-b border-border flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-52">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="搜索评论内容..."
              value={keyword}
              onChange={(e) => { setKeyword(e.target.value); setPage(1); }}
              className="w-full pl-9 pr-4 py-2 text-sm border border-border rounded bg-input-background text-foreground focus:ring-2 focus:ring-ring focus:border-transparent"
            />
          </div>
          <CustomSelect
            value={filterStatus}
            onChange={(v) => { setFilterStatus(v); setPage(1); }}
            options={commentFilterOptions}
            className="w-[132px]"
            ariaLabel="评论状态筛选"
          />
        </div>

        <div className="overflow-x-auto">
          <table className="admin-table w-full border-collapse text-left">
            <thead>
              <tr className="bg-muted/50 border-b border-border">
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground min-w-[200px]">内容</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground whitespace-nowrap">文章</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">作者</th>
                <th className="px-4 py-3 text-sm font-medium text-muted-foreground">状态</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground whitespace-nowrap">
                  时间
                </th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">加载中...</td></tr>
              ) : comments.length === 0 ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">暂无评论</td></tr>
              ) : (
                comments.map((comment) => (
                  <tr key={comment.id} className="hover:bg-muted/50">
                    <td className="px-4 py-3 text-sm text-foreground max-w-md truncate">
                      {comment.content}
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                      文章 #{comment.articleId}
                    </td>
                    <td className="px-4 py-3 text-sm text-foreground">
                      {comment.author?.nickname || comment.author?.username || '匿名'}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`px-2 py-1 rounded text-xs ${statusColors[comment.status] || 'bg-muted text-foreground'}`}>
                        {statusLabels[comment.status] || comment.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground whitespace-nowrap text-right">
                      {new Date(comment.createdAt).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        {comment.status !== 'approved' && (
                          <button
                            onClick={() => handleStatusChange(comment.id, 'approved')}
                            className="p-1.5 text-green-600 hover:bg-green-50 rounded"
                            title="通过"
                          >
                            <Check className="w-4 h-4" />
                          </button>
                        )}
                        {comment.status !== 'rejected' && (
                          <button
                            onClick={() => handleStatusChange(comment.id, 'rejected')}
                            className="p-1.5 text-red-600 hover:bg-red-50 rounded"
                            title="拒绝"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        )}
                        {comment.status !== 'spam' && (
                          <button
                            onClick={() => handleStatusChange(comment.id, 'spam')}
                            className="p-1.5 text-orange-600 hover:bg-orange-50 rounded"
                            title="标记为垃圾"
                          >
                            <ShieldAlert className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <Pagination total={total} page={page} pageSize={20} onPageChange={setPage} unit="条评论" />
      </div>
    </div>
  );
}
