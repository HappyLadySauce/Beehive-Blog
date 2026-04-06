import { useState, useEffect } from 'react';
import {
  getTags,
  createTag,
  updateTag,
  deleteTag,
  TagListItem,
} from '../../api/taxonomy';
import { toast } from 'sonner';
import { Tag as TagIcon, Edit, Trash2 } from 'lucide-react';
import AdminModal from '../../components/AdminModal';
import { TextField, TextareaField, ColorField } from '../../components/FormField';

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
      const res = editTarget
        ? await updateTag(editTarget.id, payload)
        : await createTag(payload);

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
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                  加载中...
                </td>
              </tr>
            ) : tags.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                  暂无标签
                </td>
              </tr>
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
                      className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                      title="编辑"
                    >
                      <Edit className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(tag)}
                      className="p-1.5 text-red-600 hover:bg-red-50 rounded transition-colors"
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
        <AdminModal
          title={editTarget ? '编辑标签' : '新建标签'}
          onClose={closeModal}
          onConfirm={handleSubmit}
          confirmLabel={editTarget ? '保存' : '创建'}
          loading={submitting}
          maxWidth="md"
        >
          <TextField
            label="名称"
            value={form.name}
            onChange={(v) => setForm((f) => ({ ...f, name: v }))}
            placeholder="标签名称"
            required
          />
          <TextField
            label="别名 (Slug)"
            hint="可选，留空自动生成"
            value={form.slug}
            onChange={(v) => setForm((f) => ({ ...f, slug: v }))}
            placeholder="url-friendly-slug"
          />
          <ColorField
            label="颜色"
            value={form.color}
            onChange={(v) => setForm((f) => ({ ...f, color: v }))}
          />
          <TextareaField
            label="描述"
            value={form.description}
            onChange={(v) => setForm((f) => ({ ...f, description: v }))}
            placeholder="可选"
          />
        </AdminModal>
      )}
    </div>
  );
}
