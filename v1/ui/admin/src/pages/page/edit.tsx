import { useState, useEffect, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Editor } from '@bytemd/react';
import 'bytemd/dist/index.css';
import 'github-markdown-css/github-markdown-light.css';
import '../article/github-markdown-dark-scoped.css';
import '../article/article-edit-bytemd-hljs.css';
import '../article/article-edit-bytemd-layout.css';
import { articleBytemdPlugins } from '../article/bytemd-plugins';
import { getPage, createPage, updatePage } from '../../api/page';
import request from '../../utils/request';
import { toast } from 'sonner';
import { ArrowLeft, Save } from 'lucide-react';
import CustomSelect from '../../components/CustomSelect';
import AdminModal from '../../components/AdminModal';

export default function PageEdit() {
  const { id: idParam } = useParams();
  const navigate = useNavigate();
  const pageId = idParam ? parseInt(idParam, 10) : NaN;
  const isCreate = !idParam || Number.isNaN(pageId) || pageId <= 0;

  const [title, setTitle] = useState('');
  const [slug, setSlug] = useState('');
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('draft');
  const [isInMenu, setIsInMenu] = useState(false);
  const [sortOrder, setSortOrder] = useState(0);
  const [loading, setLoading] = useState(false);
  const [loadDetail, setLoadDetail] = useState(!isCreate);
  const [settingsOpen, setSettingsOpen] = useState(false);

  useEffect(() => {
    if (isCreate) {
      setLoadDetail(false);
      return;
    }
    let cancelled = false;
    (async () => {
      setLoadDetail(true);
      try {
        const res = await getPage(pageId);
        if (cancelled) return;
        if (res.code === 200) {
          const d = res.data;
          setTitle(d.title);
          setSlug(d.slug || '');
          setContent(d.content || '');
          setStatus(d.status || 'draft');
          setIsInMenu(Boolean(d.isInMenu));
          setSortOrder(d.sortOrder ?? 0);
        } else {
          toast.error(res.message || '加载页面失败');
          navigate('/pages');
        }
      } catch {
        if (!cancelled) {
          toast.error('加载页面失败');
          navigate('/pages');
        }
      } finally {
        if (!cancelled) setLoadDetail(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [isCreate, pageId, navigate]);

  const uploadImages = async (files: File[]) => {
    const results = await Promise.all(
      files.map(async (file) => {
        const formData = new FormData();
        formData.append('file', file);
        try {
          const response = await request.post<any, any>('/api/v1/admin/upload-image', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
          if (response.code === 200) {
            return { url: response.data.url, alt: response.data.alt || file.name };
          }
        } catch {
          /* ignore */
        }
        return null;
      }),
    );
    return results.filter(Boolean) as { url: string; alt: string }[];
  };

  const statusOptions = useMemo(
    () => [
      { value: 'draft', label: '草稿' },
      { value: 'published', label: '发布' },
      { value: 'private', label: '私密' },
      { value: 'archived', label: '归档' },
    ],
    [],
  );

  const handleSave = async () => {
    if (!title.trim()) {
      toast.error('请填写页面标题');
      setSettingsOpen(true);
      return;
    }
    if (!content.trim()) {
      toast.error('请输入页面内容');
      return;
    }

    setLoading(true);
    try {
      if (isCreate) {
        const res = await createPage({
          title: title.trim(),
          content,
          status,
          isInMenu,
          sortOrder,
          ...(slug.trim() ? { slug: slug.trim() } : {}),
        });
        if (res.code === 200) {
          toast.success('创建成功');
          if (res.data?.id) {
            navigate(`/pages/edit/${res.data.id}`, { replace: true });
          } else {
            navigate('/pages');
          }
        } else {
          toast.error(res.message || '创建失败');
        }
      } else {
        const res = await updatePage(pageId, {
          title: title.trim(),
          content,
          status,
          isInMenu,
          sortOrder,
          slug: slug.trim() || undefined,
        });
        if (res.code === 200) {
          toast.success('更新成功');
        } else {
          toast.error(res.message || '更新失败');
        }
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '保存请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSettingsConfirm = async () => {
    if (!title.trim()) {
      toast.error('请填写页面标题');
      return;
    }
    if (!content.trim()) {
      toast.error('请输入页面内容');
      return;
    }
    setSettingsOpen(false);
    await handleSave();
  };

  if (loadDetail) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center text-muted-foreground">
        加载中...
      </div>
    );
  }

  return (
    <div className="article-edit-bytemd flex w-full flex-col gap-4">
      <div className="sticky top-[var(--admin-mobile-header-height)] z-20 flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-border/60 bg-background/95 py-1 backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:top-0">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => navigate('/pages')}
            className="rounded p-1.5 text-muted-foreground transition-colors hover:bg-accent"
            aria-label="返回列表"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h2 className="text-xl font-semibold tracking-tight text-foreground">
            {isCreate ? '新建页面' : '编辑页面'}
          </h2>
        </div>
        <div className="flex flex-wrap items-center gap-2 sm:gap-3">
          <button
            type="button"
            onClick={() => setSettingsOpen(true)}
            className="rounded border border-border bg-background px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent"
          >
            页面设置
          </button>
          <CustomSelect
            value={status}
            onChange={setStatus}
            options={statusOptions}
            className="w-[132px]"
            size="sm"
            ariaLabel="页面状态"
          />
          <button
            type="button"
            onClick={() => void handleSave()}
            disabled={loading}
            className="inline-flex items-center gap-1.5 rounded bg-primary px-4 py-1.5 text-sm text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
          >
            <Save className="h-4 w-4" />
            {loading ? '保存中...' : '保存'}
          </button>
        </div>
      </div>

      <div className="editor-container flex min-h-[max(20rem,calc(100dvh-10rem))] flex-col overflow-x-hidden rounded border border-border bg-card">
        <Editor
          value={content}
          plugins={articleBytemdPlugins}
          onChange={(v) => setContent(v)}
          uploadImages={uploadImages}
        />
      </div>

      {settingsOpen && (
        <AdminModal
          title="页面设置"
          onClose={() => setSettingsOpen(false)}
          onConfirm={() => void handleSettingsConfirm()}
          confirmLabel="保存"
          loading={loading}
          maxWidth="md"
        >
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground" htmlFor="page-title">
                标题
              </label>
              <input
                id="page-title"
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground" htmlFor="page-slug">
                URL 路径（slug，可选）
              </label>
              <input
                id="page-slug"
                type="text"
                placeholder="留空则自动生成"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
              />
              <p className="text-xs text-muted-foreground">生成后访问路径形如 /your-slug/</p>
            </div>
            <div className="flex items-center gap-2">
              <input
                id="page-in-menu"
                type="checkbox"
                checked={isInMenu}
                onChange={(e) => setIsInMenu(e.target.checked)}
                className="h-4 w-4 rounded border-border text-primary focus:ring-ring"
              />
              <label htmlFor="page-in-menu" className="text-sm text-foreground">
                加入站点菜单
              </label>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground" htmlFor="page-sort">
                排序权重
              </label>
              <input
                id="page-sort"
                type="number"
                value={sortOrder}
                onChange={(e) => setSortOrder(parseInt(e.target.value, 10) || 0)}
                className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>
        </AdminModal>
      )}
    </div>
  );
}
