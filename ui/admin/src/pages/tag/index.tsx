import { useState, useEffect } from 'react';
import { getTags, Tag } from '../../api/taxonomy';
import { toast } from 'sonner';
import { Tag as TagIcon, Edit, Trash2 } from 'lucide-react';

export default function Tags() {
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchTags = async () => {
    setLoading(true);
    try {
      const res = await getTags();
      if (res.code === 200) {
        setTags(res.data.items || []);
      } else {
        toast.error(res.message || '获取标签失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求标签失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTags();
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <TagIcon className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">标签管理</h2>
        </div>
        <button className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
          新建标签
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-3 text-sm font-medium text-gray-600">ID</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">名称</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">别名 (Slug)</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">颜色</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">文章数</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">加载中...</td></tr>
            ) : tags.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">暂无标签</td></tr>
            ) : (
              tags.map((tag) => (
                <tr key={tag.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-600">{tag.id}</td>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{tag.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{tag.slug}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    <span className="inline-block w-4 h-4 rounded-full align-middle mr-2" style={{ backgroundColor: tag.color || '#ccc' }}></span>
                    {tag.color || '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{tag.articleCount}</td>
                  <td className="px-4 py-3 text-right">
                    <button className="p-1.5 text-blue-600 hover:bg-blue-50 rounded" title="编辑"><Edit className="w-4 h-4" /></button>
                    <button className="p-1.5 text-red-600 hover:bg-red-50 rounded" title="删除"><Trash2 className="w-4 h-4" /></button>
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
