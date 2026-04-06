import { useState, useEffect } from 'react';
import { getComments, updateCommentStatus, Comment } from '../../api/comment';
import { toast } from 'sonner';
import { MessageSquare, Check, X, ShieldAlert, Trash2 } from 'lucide-react';

export default function Comments() {
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchComments = async () => {
    setLoading(true);
    try {
      const res = await getComments();
      if (res.code === 200) {
        setComments(res.data.items || []);
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
  }, []);

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
          <MessageSquare className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">评论管理</h2>
        </div>
      </div>

      <div className="bg-white border border-gray-200 rounded overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-3 text-sm font-medium text-gray-600">ID</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">文章</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">作者</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">内容</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">状态</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">时间</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-gray-500">加载中...</td></tr>
            ) : comments.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-gray-500">暂无评论</td></tr>
            ) : (
              comments.map((comment) => (
                <tr key={comment.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-600">{comment.id}</td>
                  <td className="px-4 py-3 text-sm text-gray-900 truncate max-w-[150px]">{comment.articleTitle}</td>
                  <td className="px-4 py-3 text-sm text-gray-900">{comment.author?.nickname || comment.author?.username}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 max-w-xs truncate">{comment.content}</td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      comment.status === 'approved' ? 'bg-green-100 text-green-800' :
                      comment.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                      comment.status === 'rejected' ? 'bg-red-100 text-red-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {comment.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{new Date(comment.createdAt).toLocaleString()}</td>
                  <td className="px-4 py-3 text-right">
                    {comment.status !== 'approved' && (
                      <button onClick={() => handleStatusChange(comment.id, 'approved')} className="p-1.5 text-green-600 hover:bg-green-50 rounded" title="通过"><Check className="w-4 h-4" /></button>
                    )}
                    {comment.status !== 'rejected' && (
                      <button onClick={() => handleStatusChange(comment.id, 'rejected')} className="p-1.5 text-red-600 hover:bg-red-50 rounded" title="拒绝"><X className="w-4 h-4" /></button>
                    )}
                    {comment.status !== 'spam' && (
                      <button onClick={() => handleStatusChange(comment.id, 'spam')} className="p-1.5 text-orange-600 hover:bg-orange-50 rounded" title="标记为垃圾"><ShieldAlert className="w-4 h-4" /></button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
