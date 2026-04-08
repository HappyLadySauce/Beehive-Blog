import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Editor } from '@bytemd/react';
import 'bytemd/dist/index.css';
import 'github-markdown-css/github-markdown-light.css';
import './github-markdown-dark-scoped.css';
import './article-edit-bytemd-hljs.css';
import './article-edit-bytemd-layout.css';
import { articleBytemdPlugins } from './bytemd-plugins';
import {
  getArticle,
  createArticle,
  updateArticle,
  listArticleVersions,
  restoreArticleVersion,
  updateArticleVersionTitle,
  deleteArticleVersion,
  ArticleDetailResponse,
  ArticleVersionItem,
} from '../../api/article';
import { getCategories, getTags, CategoryBrief, TagListItem } from '../../api/taxonomy';
import request from '../../utils/request';
import { toast } from 'sonner';
import { ArrowLeft, Save, History, Settings, Timer, Pencil, Trash2, Clock } from 'lucide-react';
import CustomSelect from '../../components/CustomSelect';
import AdminModal from '../../components/AdminModal';

/** 本地日期时间输入（datetime-local）格式，用于定时发布弹窗 */
function formatLocalDatetime(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function stableArticlePayloadSnapshot(p: {
  title: string;
  content: string;
  summary: string;
  status: string;
  publishedAt: string;
  categoryId: number | null;
  tagIds: number[];
}) {
  return JSON.stringify({
    title: p.title.trim(),
    content: p.content,
    summary: p.summary.trim(),
    status: p.status,
    publishedAt: p.publishedAt,
    categoryId: p.categoryId,
    tagIds: [...p.tagIds].sort((a, b) => a - b),
  });
}

function applyDetailToState(
  data: ArticleDetailResponse,
  setters: {
    setTitle: (v: string) => void;
    setContent: (v: string) => void;
    setSummary: (v: string) => void;
    setStatus: (v: string) => void;
    setPublishedAt: (v: string) => void;
    setCategoryId: (v: number | null) => void;
    setSelectedTagIds: (v: number[]) => void;
  },
) {
  setters.setTitle(data.title);
  setters.setContent(data.content);
  setters.setSummary(data.summary || '');
  setters.setStatus(data.status);
  setters.setPublishedAt(data.publishedAt || '');
  setters.setCategoryId(data.category?.id ?? null);
  setters.setSelectedTagIds(data.tags?.map((t) => t.id) || []);
}

export default function ArticleEdit() {
  const { id } = useParams();
  const navigate = useNavigate();
  const articleId = id ? parseInt(id, 10) : undefined;

  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [summary, setSummary] = useState('');
  const [status, setStatus] = useState('draft');
  /** RFC3339，定时发布必填 */
  const [publishedAt, setPublishedAt] = useState('');
  const [scheduleModalOpen, setScheduleModalOpen] = useState(false);
  const [scheduleLocalInput, setScheduleLocalInput] = useState('');
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<CategoryBrief[]>([]);
  const [tags, setTags] = useState<TagListItem[]>([]);

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [versions, setVersions] = useState<ArticleVersionItem[]>([]);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [restoringId, setRestoringId] = useState<number | null>(null);
  const [deletingVersionId, setDeletingVersionId] = useState<number | null>(null);
  const [editingVersionId, setEditingVersionId] = useState<number | null>(null);
  const [editingVersionTitle, setEditingVersionTitle] = useState('');
  const [savingVersionTitle, setSavingVersionTitle] = useState(false);
  /** 编辑已存在文章（URL 含 id）时默认开启自动保存；新建（/articles/create）默认关闭 */
  const [autoSaveEnabled, setAutoSaveEnabled] = useState(() => Boolean(id));

  const stateRef = useRef({
    title,
    content,
    summary,
    status,
    publishedAt,
    categoryId,
    selectedTagIds,
  });
  const lastSyncedRef = useRef<string | null>(null);
  const savingRef = useRef(false);
  const loadingRef = useRef(false);

  useEffect(() => {
    stateRef.current = {
      title,
      content,
      summary,
      status,
      publishedAt,
      categoryId,
      selectedTagIds,
    };
  }, [title, content, summary, status, publishedAt, categoryId, selectedTagIds]);

  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);

  useEffect(() => {
    const loadFilters = async () => {
      try {
        const [catRes, tagRes] = await Promise.all([
          getCategories({ pageSize: 200 }),
          getTags({ pageSize: 200 }),
        ]);
        if (catRes.code === 200) setCategories(catRes.data.list || []);
        if (tagRes.code === 200) setTags(tagRes.data.list || []);
      } catch {
        // 筛选器加载失败不阻断主流程
      }
    };
    loadFilters();
  }, []);

  useEffect(() => {
    if (!articleId) return;
    const fetchArticle = async () => {
      try {
        const res = await getArticle(articleId);
        if (res.code === 200) {
          applyDetailToState(res.data, {
            setTitle,
            setContent,
            setSummary,
            setStatus,
            setPublishedAt,
            setCategoryId,
            setSelectedTagIds,
          });
          lastSyncedRef.current = stableArticlePayloadSnapshot({
            title: res.data.title,
            content: res.data.content,
            summary: res.data.summary || '',
            status: res.data.status,
            publishedAt: res.data.publishedAt || '',
            categoryId: res.data.category?.id ?? null,
            tagIds: res.data.tags?.map((t) => t.id) || [],
          });
        } else {
          toast.error(res.message || '获取文章失败');
        }
      } catch (error: unknown) {
        const msg =
          error && typeof error === 'object' && 'response' in error
            ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
            : undefined;
        toast.error(msg || '请求文章失败');
      }
    };
    fetchArticle();
  }, [articleId]);

  const loadVersions = useCallback(async () => {
    if (!articleId) return;
    setVersionsLoading(true);
    try {
      const res = await listArticleVersions(articleId);
      if (res.code === 200) {
        setVersions(res.data.items || []);
      } else {
        toast.error(res.message || '获取版本历史失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '请求版本历史失败');
    } finally {
      setVersionsLoading(false);
    }
  }, [articleId]);

  useEffect(() => {
    if (historyOpen && articleId) {
      void loadVersions();
    }
  }, [historyOpen, articleId, loadVersions]);

  useEffect(() => {
    if (!articleId || !autoSaveEnabled) return;
    const timer = window.setInterval(() => {
      void (async () => {
        if (loadingRef.current || savingRef.current) return;
        const s = stateRef.current;
        if (!s.title.trim() || !s.content.trim()) return;
        if (s.status === 'scheduled' && !s.publishedAt.trim()) return;
        const snap = stableArticlePayloadSnapshot({
          title: s.title,
          content: s.content,
          summary: s.summary,
          status: s.status,
          publishedAt: s.publishedAt,
          categoryId: s.categoryId,
          tagIds: s.selectedTagIds,
        });
        if (snap === lastSyncedRef.current) return;
        savingRef.current = true;
        try {
          const res = await updateArticle(articleId, {
            title: s.title.trim(),
            content: s.content,
            summary: s.summary.trim() || undefined,
            status: s.status,
            publishedAt: s.publishedAt.trim() || undefined,
            categoryId: s.categoryId ?? undefined,
            tagIds: s.selectedTagIds.length > 0 ? s.selectedTagIds : undefined,
            autoSave: true,
          });
          if (res.code === 200) {
            lastSyncedRef.current = snap;
          } else {
            toast.error(res.message || '自动保存失败');
          }
        } catch (error: unknown) {
          const msg =
            error && typeof error === 'object' && 'response' in error
              ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
              : undefined;
          toast.error(msg || '自动保存请求失败');
        } finally {
          savingRef.current = false;
        }
      })();
    }, 1000);
    return () => window.clearInterval(timer);
  }, [articleId, autoSaveEnabled]);

  const handleSave = async () => {
    if (!title.trim()) {
      setSettingsOpen(true);
      toast.error('请在设置中填写文章标题');
      return;
    }
    if (!content.trim()) {
      toast.error('请输入文章内容');
      return;
    }
    if (status === 'scheduled') {
      const t = publishedAt.trim();
      if (!t) {
        setScheduleModalOpen(true);
        toast.error('请先设置定时发布时间');
        return;
      }
      if (new Date(t).getTime() <= Date.now()) {
        toast.error('定时发布时间须晚于当前时间');
        setScheduleModalOpen(true);
        return;
      }
    }

    setLoading(true);
    try {
      const payload = {
        title: title.trim(),
        content,
        summary: summary.trim() || undefined,
        status,
        publishedAt: publishedAt.trim() || undefined,
        categoryId: categoryId ?? undefined,
        tagIds: selectedTagIds.length > 0 ? selectedTagIds : undefined,
      };

      let res;
      if (articleId) {
        res = await updateArticle(articleId, payload);
      } else {
        res = await createArticle(payload);
      }

      if (res.code === 200) {
        toast.success(articleId ? '更新成功' : '创建成功');
        setAutoSaveEnabled(true);
        if (!articleId && res.data?.id) {
          navigate(`/articles/edit/${res.data.id}`, { replace: true });
        } else if (articleId) {
          lastSyncedRef.current = stableArticlePayloadSnapshot({
            title: title.trim(),
            content,
            summary,
            status,
            publishedAt,
            categoryId,
            tagIds: selectedTagIds,
          });
        }
      } else {
        toast.error(res.message || '保存失败');
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

  /** 设置中点击「完成」：校验后关闭弹窗并执行与顶栏保存相同的保存逻辑 */
  const handleSettingsConfirm = async () => {
    if (!title.trim()) {
      toast.error('请填写文章标题');
      return;
    }
    if (!content.trim()) {
      toast.error('请输入文章内容');
      return;
    }
    setSettingsOpen(false);
    await handleSave();
  };

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
        } catch (error) {
          console.error('Upload failed', error);
        }
        return null;
      }),
    );
    return results.filter(Boolean) as { url: string; alt: string }[];
  };

  const toggleTag = (tagId: number) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId],
    );
  };

  const handleStartEditVersion = (v: ArticleVersionItem) => {
    setEditingVersionId(v.id);
    setEditingVersionTitle(v.title || '');
  };

  const handleCancelEditVersion = () => {
    setEditingVersionId(null);
    setEditingVersionTitle('');
  };

  const handleCommitVersionTitle = async () => {
    if (!articleId || editingVersionId === null) return;
    const t = editingVersionTitle.trim();
    if (!t) {
      toast.error('版本名称不能为空');
      return;
    }
    setSavingVersionTitle(true);
    try {
      const res = await updateArticleVersionTitle(articleId, editingVersionId, { title: t });
      if (res.code === 200) {
        toast.success('已更新版本名称');
        setEditingVersionId(null);
        setEditingVersionTitle('');
        void loadVersions();
      } else {
        toast.error(res.message || '更新失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '更新版本名称失败');
    } finally {
      setSavingVersionTitle(false);
    }
  };

  const handleDeleteVersion = async (versionId: number) => {
    if (!articleId) return;
    if (!window.confirm('确定删除该版本记录？此操作不可恢复。')) return;
    setDeletingVersionId(versionId);
    try {
      const res = await deleteArticleVersion(articleId, versionId);
      if (res.code === 200) {
        toast.success('已删除版本');
        if (editingVersionId === versionId) {
          handleCancelEditVersion();
        }
        void loadVersions();
      } else {
        toast.error(res.message || '删除失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '删除版本失败');
    } finally {
      setDeletingVersionId(null);
    }
  };

  const handleRestoreVersion = async (versionId: number) => {
    if (!articleId) return;
    setRestoringId(versionId);
    try {
      const res = await restoreArticleVersion(articleId, versionId);
      if (res.code === 200) {
        toast.success('已恢复到所选版本');
        applyDetailToState(res.data, {
          setTitle,
          setContent,
          setSummary,
          setStatus,
          setPublishedAt,
          setCategoryId,
          setSelectedTagIds,
        });
        lastSyncedRef.current = stableArticlePayloadSnapshot({
          title: res.data.title,
          content: res.data.content,
          summary: res.data.summary || '',
          status: res.data.status,
          publishedAt: res.data.publishedAt || '',
          categoryId: res.data.category?.id ?? null,
          tagIds: res.data.tags?.map((t) => t.id) || [],
        });
        setHistoryOpen(false);
      } else {
        toast.error(res.message || '恢复失败');
      }
    } catch (error: unknown) {
      const msg =
        error && typeof error === 'object' && 'response' in error
          ? (error as { response?: { data?: { message?: string } } }).response?.data?.message
          : undefined;
      toast.error(msg || '恢复请求失败');
    } finally {
      setRestoringId(null);
    }
  };

  const openScheduleModal = useCallback(() => {
    let base: Date;
    if (publishedAt.trim()) {
      const p = new Date(publishedAt);
      base = Number.isNaN(p.getTime()) ? new Date(Date.now() + 60 * 60 * 1000) : p;
    } else {
      base = new Date(Date.now() + 60 * 60 * 1000);
    }
    base.setSeconds(0, 0);
    setScheduleLocalInput(formatLocalDatetime(base));
    setScheduleModalOpen(true);
  }, [publishedAt]);

  const handleConfirmSchedule = useCallback(() => {
    const d = new Date(scheduleLocalInput);
    if (Number.isNaN(d.getTime())) {
      toast.error('请选择有效的日期与时间');
      return;
    }
    if (d.getTime() <= Date.now()) {
      toast.error('定时发布时间须晚于当前时间');
      return;
    }
    setPublishedAt(d.toISOString());
    setStatus('scheduled');
    setScheduleModalOpen(false);
  }, [scheduleLocalInput]);

  const handleStatusSelect = useCallback(
    (v: string) => {
      if (v === 'scheduled') {
        openScheduleModal();
        return;
      }
      if (status === 'scheduled' && v !== 'scheduled') {
        setPublishedAt('');
      }
      setStatus(v);
    },
    [status, openScheduleModal],
  );

  const statusOptions = useMemo(
    () => [
      { value: 'draft', label: '草稿' },
      { value: 'published', label: '发布' },
      { value: 'scheduled', label: '定时发布' },
      { value: 'private', label: '私密' },
      { value: 'archived', label: '归档' },
    ],
    [],
  );

  const categoryOptions = useMemo(
    () => [
      { value: '', label: '无分类' },
      ...categories.map((c) => ({ value: String(c.id), label: c.name })),
    ],
    [categories],
  );

  return (
    <div className="article-edit-bytemd flex h-full min-h-0 flex-col gap-4">
      <div className="flex shrink-0 items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => navigate('/articles')}
            className="rounded p-1.5 text-muted-foreground transition-colors hover:bg-accent"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h2 className="text-lg font-medium text-foreground">
            {articleId ? '编辑文章' : '新建文章'}
          </h2>
        </div>
        <div className="flex flex-wrap items-center gap-2 sm:gap-3">
          <button
            type="button"
            disabled={!articleId}
            title={!articleId ? '保存为正式文章后可使用自动保存' : undefined}
            onClick={() => setAutoSaveEnabled((v) => !v)}
            className={`inline-flex items-center gap-1.5 rounded border px-3 py-1.5 text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
              autoSaveEnabled && articleId
                ? 'border-primary bg-primary/10 text-foreground'
                : 'border-border bg-background text-foreground hover:bg-accent'
            }`}
          >
            <Timer className="h-4 w-4" />
            自动保存 {articleId ? (autoSaveEnabled ? '开' : '关') : ''}
          </button>
          <button
            type="button"
            disabled={!articleId}
            title={!articleId ? '保存文章后可查看版本历史' : undefined}
            onClick={() => setHistoryOpen(true)}
            className="inline-flex items-center gap-1.5 rounded border border-border bg-background px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
          >
            <History className="h-4 w-4" />
            历史版本
          </button>
          <button
            type="button"
            onClick={() => setSettingsOpen(true)}
            className="inline-flex items-center gap-1.5 rounded border border-border bg-background px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent"
          >
            <Settings className="h-4 w-4" />
            设置
          </button>
          <CustomSelect
            value={status}
            onChange={handleStatusSelect}
            options={statusOptions}
            className="w-[132px]"
            size="sm"
            ariaLabel="文章状态"
          />
          {status === 'scheduled' && (
            <button
              type="button"
              onClick={() => openScheduleModal()}
              className="inline-flex items-center gap-1 rounded border border-border bg-background px-2 py-1.5 text-sm text-foreground hover:bg-accent"
              title="修改计划发布时间"
            >
              <Clock className="h-4 w-4" />
              计划时间
            </button>
          )}
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

      <div className="editor-container min-h-0 flex-1 overflow-hidden rounded border border-border bg-card">
        <Editor
          value={content}
          plugins={articleBytemdPlugins}
          onChange={(v) => setContent(v)}
          uploadImages={uploadImages}
        />
      </div>

      {scheduleModalOpen && (
        <AdminModal
          title="定时发布"
          onClose={() => setScheduleModalOpen(false)}
          onConfirm={handleConfirmSchedule}
          confirmLabel="确认定时"
          maxWidth="md"
        >
          <p className="text-sm text-muted-foreground">
            请选择文章自动转为「发布」状态的日期与时间（使用本机时区）。
          </p>
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground" htmlFor="article-schedule-at">
              发布时间
            </label>
            <input
              id="article-schedule-at"
              type="datetime-local"
              value={scheduleLocalInput}
              onChange={(e) => setScheduleLocalInput(e.target.value)}
              className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
            />
          </div>
        </AdminModal>
      )}

      {settingsOpen && (
        <AdminModal
          title="文章设置"
          onClose={() => setSettingsOpen(false)}
          onConfirm={() => void handleSettingsConfirm()}
          confirmLabel="完成"
          loading={loading}
          maxWidth="md"
        >
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">文章标题</label>
              <input
                type="text"
                placeholder="输入文章标题..."
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="space-y-2">
              <span className="text-sm font-medium text-foreground">分类</span>
              <CustomSelect
                value={categoryId === null ? '' : String(categoryId)}
                onChange={(v) => setCategoryId(v === '' ? null : parseInt(v, 10))}
                options={categoryOptions}
                className="w-full"
                size="sm"
                ariaLabel="文章分类"
              />
            </div>
            <div className="space-y-2">
              <span className="text-sm font-medium text-foreground">标签</span>
              <div className="flex max-h-40 flex-wrap gap-2 overflow-y-auto rounded border border-border p-2">
                {tags.map((tag) => (
                  <button
                    key={tag.id}
                    type="button"
                    onClick={() => toggleTag(tag.id)}
                    className={`rounded border px-2 py-1 text-xs transition-colors ${
                      selectedTagIds.includes(tag.id)
                        ? 'border-primary bg-primary text-primary-foreground'
                        : 'border-border bg-card text-foreground hover:border-primary/50'
                    }`}
                  >
                    {tag.name}
                  </button>
                ))}
                {tags.length === 0 && (
                  <span className="text-xs text-muted-foreground">暂无标签</span>
                )}
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">摘要</label>
              <textarea
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                placeholder="可选，留空则自动截取正文..."
                rows={4}
                className="w-full resize-none rounded-md border border-border bg-input-background px-3 py-2 text-sm text-foreground focus:ring-2 focus:ring-ring focus:border-transparent"
              />
            </div>
          </div>
        </AdminModal>
      )}

      {historyOpen && articleId && (
        <AdminModal
          title="版本历史"
          onClose={() => {
            setHistoryOpen(false);
            handleCancelEditVersion();
          }}
          maxWidth="lg"
        >
          <div className="max-h-[min(60vh,420px)] overflow-auto">
            {versionsLoading ? (
              <p className="text-sm text-muted-foreground">加载中...</p>
            ) : versions.length === 0 ? (
              <p className="text-sm text-muted-foreground">暂无历史版本（保存修改后会生成版本记录）</p>
            ) : (
              <ul className="divide-y divide-border">
                {versions.map((v) => (
                  <li
                    key={v.id}
                    className="flex flex-col gap-2 py-3 sm:flex-row sm:items-start sm:justify-between"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="text-sm font-medium text-foreground">
                        {v.isAutosave ? (
                          <span className="rounded bg-muted px-1.5 py-0.5 text-xs font-normal text-muted-foreground">
                            自动保存
                          </span>
                        ) : (
                          <>版本 {v.version}</>
                        )}
                      </div>
                      {editingVersionId === v.id ? (
                        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
                          <input
                            type="text"
                            value={editingVersionTitle}
                            onChange={(e) => setEditingVersionTitle(e.target.value)}
                            className="w-full min-w-0 rounded-md border border-border bg-input-background px-3 py-1.5 text-sm text-foreground focus:border-transparent focus:ring-2 focus:ring-ring"
                            placeholder="版本显示名称"
                            maxLength={200}
                          />
                          <div className="flex shrink-0 gap-2">
                            <button
                              type="button"
                              disabled={savingVersionTitle}
                              onClick={() => void handleCommitVersionTitle()}
                              className="rounded border border-border bg-background px-2 py-1 text-xs hover:bg-accent disabled:opacity-50"
                            >
                              {savingVersionTitle ? '保存中...' : '保存'}
                            </button>
                            <button
                              type="button"
                              disabled={savingVersionTitle}
                              onClick={handleCancelEditVersion}
                              className="rounded border border-border bg-background px-2 py-1 text-xs hover:bg-accent disabled:opacity-50"
                            >
                              取消
                            </button>
                          </div>
                        </div>
                      ) : (
                        <div className="mt-1 text-sm font-normal text-muted-foreground">
                          {v.title ? (
                            <span className="break-words">{v.title}</span>
                          ) : (
                            '（无标题）'
                          )}
                        </div>
                      )}
                      <div className="mt-1 text-xs text-muted-foreground">
                        {v.createdAt ? new Date(v.createdAt).toLocaleString() : ''}
                      </div>
                    </div>
                    {editingVersionId !== v.id && (
                      <div className="flex shrink-0 flex-wrap items-center gap-2">
                        <button
                          type="button"
                          disabled={
                            restoringId !== null ||
                            deletingVersionId !== null ||
                            savingVersionTitle
                          }
                          onClick={() => handleStartEditVersion(v)}
                          className="inline-flex items-center gap-1 rounded border border-border bg-background px-2 py-1.5 text-xs hover:bg-accent disabled:opacity-50"
                          title="修改名称"
                        >
                          <Pencil className="h-3.5 w-3.5" />
                          改名
                        </button>
                        <button
                          type="button"
                          disabled={
                            restoringId !== null ||
                            deletingVersionId !== null ||
                            savingVersionTitle
                          }
                          onClick={() => void handleDeleteVersion(v.id)}
                          className="inline-flex items-center gap-1 rounded border border-destructive/50 bg-background px-2 py-1.5 text-xs text-destructive hover:bg-destructive/10 disabled:opacity-50"
                          title="删除此版本记录"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                          {deletingVersionId === v.id ? '删除中...' : '删除'}
                        </button>
                        <button
                          type="button"
                          disabled={restoringId !== null || deletingVersionId !== null}
                          onClick={() => void handleRestoreVersion(v.id)}
                          className="rounded border border-border bg-background px-3 py-1.5 text-sm hover:bg-accent disabled:opacity-50"
                        >
                          {restoringId === v.id ? '恢复中...' : '恢复此版本'}
                        </button>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </AdminModal>
      )}

      <style>{`
        .article-edit-bytemd .editor-container .bytemd {
          height: calc(100vh - 12rem);
          min-height: 320px;
          border: none;
        }
        .article-edit-bytemd .bytemd-toolbar {
          border-bottom: 1px solid var(--border);
        }
      `}</style>
    </div>
  );
}
