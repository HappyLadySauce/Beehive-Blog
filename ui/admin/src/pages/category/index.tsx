import { useState, useEffect, useCallback } from 'react';
import { useOutletContext } from 'react-router-dom';
import type { ArticleSectionOutletContext } from '../../layouts/ArticleSectionLayout';
import {
  getCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  CategoryBrief,
} from '../../api/taxonomy';
import { toast } from 'sonner';
import { Edit, Trash2 } from 'lucide-react';
import AdminModal from '../../components/AdminModal';
import ConfirmModal from '../../components/ConfirmModal';
import { TextField, TextareaField } from '../../components/FormField';

interface CategoryForm {
  name: string;
  slug: string;
  description: string;
}

const emptyForm: CategoryForm = { name: '', slug: '', description: '' };

export default function Categories() {
  const { registerNewAction } = useOutletContext<ArticleSectionOutletContext>();
  const [categories, setCategories] = useState<CategoryBrief[]>([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editTarget, setEditTarget] = useState<CategoryBrief | null>(null);
  const [form, setForm] = useState<CategoryForm>(emptyForm);
  const [submitting, setSubmitting] = useState(false);

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

  const openCreate = useCallback(() => {
    setEditTarget(null);
    setForm(emptyForm);
    setShowModal(true);
  }, []);

  useEffect(() => {
    registerNewAction(openCreate);
    return () => registerNewAction(null);
  }, [registerNewAction, openCreate]);

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
      const res = editTarget
        ? await updateCategory(editTarget.id, payload)
        : await createCategory(payload);

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

  const handleDelete = (cat: CategoryBrief) => {
    showConfirm({
      title: '删除分类',
      message: `确定要删除分类「${cat.name}」吗？`,
      confirmLabel: '删除',
      confirmVariant: 'danger',
      onConfirm: async () => {
        hideConfirm();
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
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="admin-card admin-card-glass overflow-hidden rounded border">
        <table className="admin-table w-full border-collapse text-left">
          <thead>
            <tr className="bg-muted/50 border-b border-border">
              <th className="px-4 py-3 text-sm font-medium text-muted-foreground min-w-[180px]">名称</th>
              <th className="px-4 py-3 text-sm font-medium text-muted-foreground">描述</th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground whitespace-nowrap">
                文章数
              </th>
              <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {loading ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                  加载中...
                </td>
              </tr>
            ) : categories.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-muted-foreground">
                  暂无分类
                </td>
              </tr>
            ) : (
              categories.map((cat) => (
                <tr key={cat.id} className="hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <div className="text-sm font-semibold text-foreground">{cat.name}</div>
                    <div className="text-xs text-muted-foreground mt-0.5 font-mono">{cat.slug}</div>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground max-w-xs truncate">
                    {cat.description || '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground text-right tabular-nums">
                    {cat.articleCount}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(cat)}
                      className="p-1.5 text-primary hover:bg-primary/10 rounded transition-colors"
                      title="编辑"
                    >
                      <Edit className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => handleDelete(cat)}
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
          title={editTarget ? '编辑分类' : '新建分类'}
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
            placeholder="分类名称"
            required
          />
          <TextField
            label="别名 (Slug)"
            hint="可选，留空自动生成"
            value={form.slug}
            onChange={(v) => setForm((f) => ({ ...f, slug: v }))}
            placeholder="url-friendly-slug"
          />
          <TextareaField
            label="描述"
            value={form.description}
            onChange={(v) => setForm((f) => ({ ...f, description: v }))}
            placeholder="可选"
          />
        </AdminModal>
      )}

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
