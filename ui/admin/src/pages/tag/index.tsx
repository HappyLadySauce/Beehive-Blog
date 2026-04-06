import { useState, useEffect } from 'react';
import {
  getTags,
  createTag,
  updateTag,
  deleteTag,
  TagListItem,
} from '../../api/taxonomy';
import { toast } from 'sonner';
import { Tag as TagIcon, Edit, Trash2, X } from 'lucide-react';

interface TagForm {
  name: string;
  slug: string;
  color: string;
  description: string;
}

const emptyForm: TagForm = { name: '', slug: '', color: '#3B82F6', description: '' };

export default function Tags() {
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<TagListItem | null>(null);
  const [form, setForm] = useState<TagForm>(emptyForm);
  const [submitting, setSubmitting] = useState(false);

  const fetchTags = async () => {
    setLoading(true);
    try {
      const res = await getTags({ pageSize: 200 });
      if (res.code === 200) {
        setTags(res.data.list || []);
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

  const openCreate = () => {
    setEditTarget(null);
    setForm(emptyForm);
    setShowModal(true);
  };

  const openEdit = (tag: TagListItem) => {
    setEditTarget(tag);
    setForm({
      name: tag.name,
      slug: tag.slug,
      color: tag.color || '#3B82F6',
      description: tag.description || '',
    });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditTarget(null);
    setForm(emptyForm);
  };

  const handleSubmit = async () => {
    if (!form.name.trim()) {
      toast.error('请输入标签名称');
      return;
    }
    setSubmitting(true);
    try {
      const payload = {
        name: form.name.trim(),
        slug: form.slug.trim() || undefined,
        color: form.color || undefined,
        description: form.description.trim() || undefined,
      };
      let res;
      if (editTarget) {
        res = await updateTag(editTarget.id, payload);
      } else {
        res = await createTag(payload);
      }
      if (res.code === 200) {
        toast.success(editTarget ? '更新成功' : '创建成功');
        closeModal();
        fetchTags();
      } else {
        toast.error(res.message || '操作失败');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || '请求失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (tag: TagListItem) => {
    if (!window.confirm(`确定要删除标签「${tag.name}」吗？`)) return;
    try {
      const res = await deleteTag(tag.id, true);
      if (res.code === 200) {
        toast.success('删除成功');
        fetchTags();
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
          <TagIcon className="w-5 h-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">标签管理</h2>
        </div>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
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
                    <div className="flex items-center gap-2">
                      <span
                        className="inline-block w-4 h-4 rounded-full border border-gray-200"
                        style={{ backgroundColor: tag.color || '#ccc' }}
                      />
                      <span>{tag.color || '-'}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">{tag.articleCount}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(tag)}
                      className="p-1.5 text-blue-600 hover:bg-blue-50 rounded"
                      title="编辑"
                    >
                      <Edit className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(tag)}
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
                {editTarget ? '编辑标签' : '新建标签'}
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
                  placeholder="标签名称"
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
                <label className="block text-sm font-medium text-gray-700 mb-1">颜色</label>
                <div className="flex items-center gap-3">
                  <input
                    type="color"
                    value={form.color}
                    onChange={(e) => setForm(f => ({ ...f, color: e.target.value }))}
                    className="w-10 h-10 rounded border border-gray-300 cursor-pointer p-0.5"
                  />
                  <input
                    type="text"
                    value={form.color}
                    onChange={(e) => setForm(f => ({ ...f, color: e.target.value }))}
                    placeholder="#3B82F6"
                    className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
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
