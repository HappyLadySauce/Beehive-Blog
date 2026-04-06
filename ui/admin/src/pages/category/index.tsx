import { useState, useEffect } from 'react';
import { getCategories, Category } from '../../api/taxonomy';
import { toast } from 'sonner';
import { FolderOpen, Edit, Trash2 } from 'lucide-react';

export default function Categories() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchCategories = async () => {
    setLoading(true);
    try {
      const res = await getCategories();
      if (res.code === 200) {
        setCategories(res.data.items || []);
      } else {
        toast.error(res.message || '获取分类失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求分类失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCategories();
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FolderOpen className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">分类管理</h2>
        </div>
        <button className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
          新建分类
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-3 text-sm font-medium text-gray-600">ID</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">名称</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">别名 (Slug)</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">文章数</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">加载中...</td></tr>
            ) : categories.length === 0 ? (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">暂无分类</td></tr>
            ) : (
              categories.map((cat) => (
                <tr key={cat.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-600">{cat.id}</td>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{cat.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{cat.slug}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{cat.articleCount}</td>
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
