import { useState, useEffect } from 'react';
import {
  getCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  CategoryBrief,
} from '../../api/taxonomy';
import { toast } from 'sonner';
import { FolderOpen, Edit, Trash2, X } from 'lucide-react';

interface CategoryForm {
  name: string;
  slug: string;
  description: string;
}

const emptyForm: CategoryForm = { name: '', slug: '', description: '' };

export default function Categories() {
  const [categories, setCategories] = useState<CategoryBrief[]>([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<CategoryBrief | null>(null);
  const [form, setForm] = useState<CategoryForm>(emptyForm);
  const [submitting, setSubmitting] = useState(false);

  const fetchCategories = async () => {
    setLoading(true);
    try {
      const res = await getCategories({ pageSize: 200 });
      if (res.code === 200) {
        setCategories(res.data.list || []);
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

  const openCreate = () => {
    setEditTarget(null);
    setForm(emptyForm);
    setShowModal(true);
  };

  const openEdit = (cat: CategoryBrief) => {
    setEditTarget(cat);
    setForm({ name: cat.name, slug: cat.slug, description: cat.description || '' });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditTarget(null);
    setForm(emptyForm);
  };

  const handleSubmit = async () => {
    if (!form.name.trim()) {
      toast.error('请输入分类名称');
      return;
    }
    setSubmitting(true);
    try {
      const payload = {
        name: form.name.trim(),
        slug: form.slug.trim() || undefined,
        description: form.description.trim() || undefined,
      };
      let res;
      if (editTarget) {
        res = await updateCategory(editTarget.id, payload);
      } else {
        res = await createCategory(payload);
      }
      if (res.code === 200) {
        toast.success(editTarget ? '更新成功' : '创建成功');
        closeModal();
        fetchCategories();
      } else {
        toast.error(res.message || '操作失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (cat: CategoryBrief) => {
    if (!window.confirm(`确定要删除分类「${cat.name}」吗？`)) return;
    try {
      const res = await deleteCategory(cat.id);
      if (res.code === 200) {
        toast.success('删除成功');
        fetchCategories();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '删除请求失败');
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FolderOpen className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">分类管理</h2>
        </div>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
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
              <th className="px-4 py-3 text-sm font-medium text-gray-600">描述</th>
              <th className="px-4 py-3 text-sm font-medium text-gray-600">文章数</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {loading ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">加载中...</td></tr>
            ) : categories.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">暂无分类</td></tr>
            ) : (
              categories.map((cat) => (
                <tr key={cat.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-600">{cat.id}</td>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{cat.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{cat.slug}</td>
                  <td className="px-4 py-3 text-sm text-gray-500 max-w-xs truncate">{cat.description || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{cat.articleCount}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(cat)}
                      className="p-1.5 text-blue-600 hover:bg-blue-50 rounded"
                      title="编辑"
                    >
                      <Edit className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(cat)}
                      className="p-1.5 text-red-600 hover:bg-red-50 rounded"
                      title="删除"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h3 className="text-base font-medium text-gray-900">
                {editTarget ? '编辑分类' : '新建分类'}
              </h3>
              <button onClick={closeModal} className="p-1 text-gray-400 hover:text-gray-600 rounded">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  名称 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="分类名称"
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  别名 (Slug) <span className="text-gray-400 font-normal">可选，留空自动生成</span>
                </label>
                <input
                  type="text"
                  value={form.slug}
                  onChange={(e) => setForm(f => ({ ...f, slug: e.target.value }))}
                  placeholder="url-friendly-slug"
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">描述</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm(f => ({ ...f, description: e.target.value }))}
                  placeholder="可选"
                  rows={3}
                  className="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                />
              </div>
            </div>
            <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
              <button
                onClick={closeModal}
                className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSubmit}
                disabled={submitting}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50"
              >
                {submitting ? '提交中...' : (editTarget ? '保存' : '创建')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
