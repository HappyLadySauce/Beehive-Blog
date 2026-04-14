import { useState, useEffect, useLayoutEffect, useMemo, useCallback, useRef } from 'react';
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
import { alignTaxonomyIds, fetchLatestTaxonomy } from '../../lib/ensureTaxonomy';
import { closeAdminWebSocket, sendArticleAutosave } from '../../lib/adminWs';
import { useArticleNotePropsStore } from './articleNotePropsStore';
import { createNotePropertiesBytemdPlugin } from './note-properties-bytemd-plugin';
import request from '../../utils/request';
import { toast } from 'sonner';
import { ArrowLeft, Save, History, Timer, Pencil, Trash2, Clock } from 'lucide-react';
import AdminModal from '../../components/AdminModal';
import ConfirmModal from '../../components/ConfirmModal';
import { Switch } from '../../app/components/ui/switch';
import {
  splitFrontMatter,
  buildMatterDocument,
  buildArticleContent,
} from '../../lib/frontMatter';

const LS_SHOW_LINE_NUMBERS = 'beehive.articleEdit.showLineNumbers';
const LS_SHOW_NOTE_PROPERTIES = 'beehive.articleEdit.showNoteProperties';

function readBoolPref(key: string, defaultVal: boolean): boolean {
  try {
    const v = localStorage.getItem(key);
    if (v === null) {
      return defaultVal;
    }
    return v === 'true';
  } catch {
    return defaultVal;
  }
}

/**
 * 由编辑区正文与元数据生成写入 API 的完整 content（含 YAML Front Matter）。
 */
function computeStoredArticleContent(
  editorBody: string,
  p: {
    title: string;
    slug: string;
    summary: string;
    status: string;
    publishedAt: string;
    categoryId: number | null;
    selectedTagIds: number[];
  },
  categories: CategoryBrief[],
  tags: TagListItem[],
): string {
  const categoryNames: string[] = [];
  if (p.categoryId != null) {
    const c = categories.find((x) => x.id === p.categoryId);
    if (c) {
      categoryNames.push(c.name);
    }
  }
  const tagNames = p.selectedTagIds
    .map((id) => tags.find((t) => t.id === id)?.name)
    .filter((n): n is string => Boolean(n));
  const matter = buildMatterDocument({
    title: p.title,
    slug: p.slug,
    summary: p.summary,
    status: p.status,
    categoryNames,
    tagNames,
    publishedAt: p.publishedAt.trim() || undefined,
  });
  return buildArticleContent(matter, editorBody);
}

/** 与详情接口数据对齐，用于初次/恢复后的 lastSynced 快照（不依赖分类下拉是否已请求）。 */
function computeSyncedFullContentFromDetail(data: ArticleDetailResponse, editorBody: string): string {
  const categoryNames = data.category ? [data.category.name] : [];
  const tagNames = data.tags?.map((t) => t.name) || [];
  const matter = buildMatterDocument({
    title: data.title,
    slug: data.slug || '',
    summary: data.summary || '',
    status: data.status,
    categoryNames,
    tagNames,
    publishedAt: data.publishedAt || undefined,
  });
  return buildArticleContent(matter, editorBody);
}

/** 本地日期时间输入（datetime-local）格式，用于定时发布弹窗 */
function formatLocalDatetime(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function stableArticlePayloadSnapshot(p: {
  title: string;
  slug: string;
  content: string;
  summary: string;
  status: string;
  publishedAt: string;
  categoryId: number | null;
  tagIds: number[];
}) {
  return JSON.stringify({
    title: p.title.trim(),
    slug: p.slug.trim(),
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
    setEditorBody: (v: string) => void;
    setSlug: (v: string) => void;
    setSummary: (v: string) => void;
    setStatus: (v: string) => void;
    setPublishedAt: (v: string) => void;
    setCategoryId: (v: number | null) => void;
    setSelectedTagIds: (v: number[]) => void;
  },
) {
  setters.setTitle(data.title);
  setters.setSlug(data.slug || '');
  const split = splitFrontMatter(data.content);
  setters.setEditorBody(split ? split.body : data.content);
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
  /** 不含 Front Matter，仅 Markdown 正文 */
  const [editorBody, setEditorBody] = useState('');
  const [slug, setSlug] = useState('');
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
  const categoriesRef = useRef(categories);
  const tagsRef = useRef(tags);

  const [historyOpen, setHistoryOpen] = useState(false);
  const [versions, setVersions] = useState<ArticleVersionItem[]>([]);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [restoringId, setRestoringId] = useState<number | null>(null);
  const [deletingVersionId, setDeletingVersionId] = useState<number | null>(null);
  /** 待确认删除的版本 id，非 null 时显示 ConfirmModal */
  const [versionDeleteConfirmId, setVersionDeleteConfirmId] = useState<number | null>(null);
  const [editingVersionId, setEditingVersionId] = useState<number | null>(null);
  const [editingVersionTitle, setEditingVersionTitle] = useState('');
  const [savingVersionTitle, setSavingVersionTitle] = useState(false);
  /** 编辑已存在文章（URL 含 id）时默认开启自动保存；新建（/articles/create）默认关闭 */
  const [autoSaveEnabled, setAutoSaveEnabled] = useState(() => Boolean(id));

  const [showLineNumbers, setShowLineNumbers] = useState(() =>
    readBoolPref(LS_SHOW_LINE_NUMBERS, true),
  );
  const [showNoteProperties, setShowNoteProperties] = useState(() =>
    readBoolPref(LS_SHOW_NOTE_PROPERTIES, true),
  );

  useEffect(() => {
    try {
      localStorage.setItem(LS_SHOW_LINE_NUMBERS, String(showLineNumbers));
    } catch {
      /* ignore */
    }
  }, [showLineNumbers]);

  useEffect(() => {
    try {
      localStorage.setItem(LS_SHOW_NOTE_PROPERTIES, String(showNoteProperties));
    } catch {
      /* ignore */
    }
  }, [showNoteProperties]);

  const stateRef = useRef({
    title,
    editorBody,
    slug,
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
      editorBody,
      slug,
      summary,
      status,
      publishedAt,
      categoryId,
      selectedTagIds,
    };
  }, [title, editorBody, slug, summary, status, publishedAt, categoryId, selectedTagIds]);

  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);

  useEffect(() => {
    categoriesRef.current = categories;
  }, [categories]);

  useEffect(() => {
    tagsRef.current = tags;
  }, [tags]);

  useEffect(() => {
    return () => {
      closeAdminWebSocket();
    };
  }, []);

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
          const split = splitFrontMatter(res.data.content);
          const eb = split ? split.body : res.data.content;
          applyDetailToState(res.data, {
            setTitle,
            setEditorBody,
            setSlug,
            setSummary,
            setStatus,
            setPublishedAt,
            setCategoryId,
            setSelectedTagIds,
          });
          lastSyncedRef.current = stableArticlePayloadSnapshot({
            title: res.data.title,
            slug: res.data.slug || '',
            content: computeSyncedFullContentFromDetail(res.data, eb),
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
        if (!s.title.trim() || !s.editorBody.trim()) return;
        if (s.status === 'scheduled' && !s.publishedAt.trim()) return;
        savingRef.current = true;
        try {
          const cats = categoriesRef.current;
          const tgs = tagsRef.current;
          const aligned = alignTaxonomyIds(s.categoryId, s.selectedTagIds, cats, tgs);
          setCategoryId(aligned.categoryId);
          setSelectedTagIds(aligned.tagIds);
          const fullContentEnsured = computeStoredArticleContent(
            s.editorBody,
            {
              title: s.title,
              slug: s.slug,
              summary: s.summary,
              status: s.status,
              publishedAt: s.publishedAt,
              categoryId: aligned.categoryId,
              selectedTagIds: aligned.tagIds,
            },
            cats,
            tgs,
          );
          const snapEnsured = stableArticlePayloadSnapshot({
            title: s.title,
            slug: s.slug,
            content: fullContentEnsured,
            summary: s.summary,
            status: s.status,
            publishedAt: s.publishedAt,
            categoryId: aligned.categoryId,
            tagIds: aligned.tagIds,
          });
          if (snapEnsured === lastSyncedRef.current) {
            savingRef.current = false;
            return;
          }
          const requestId =
            typeof crypto !== 'undefined' && crypto.randomUUID
              ? crypto.randomUUID()
              : `${Date.now()}-${Math.random()}`;
          const res = await sendArticleAutosave(
            articleId,
            {
              title: s.title.trim(),
              content: fullContentEnsured,
              slug: s.slug.trim() || undefined,
              summary: s.summary.trim() || undefined,
              status: s.status,
              publishedAt: s.publishedAt.trim() || undefined,
              categoryId: aligned.categoryId ?? undefined,
              tagIds: aligned.tagIds.length > 0 ? aligned.tagIds : undefined,
              autoSave: true,
            },
            requestId,
          );
          if (res.code === 200) {
            lastSyncedRef.current = snapEnsured;
          } else {
            toast.error(res.message || '自动保存失败');
          }
        } catch (error: unknown) {
          const msg =
            error instanceof Error
              ? error.message
              : error && typeof error === 'object' && 'message' in error
                ? String((error as { message?: string }).message)
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
      toast.error('请填写文章标题（笔记属性中的 title）');
      return;
    }
    if (!editorBody.trim()) {
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
      const { categories: cats, tags: tgs } = await fetchLatestTaxonomy();
      setCategories(cats);
      setTags(tgs);
      const aligned = alignTaxonomyIds(categoryId, selectedTagIds, cats, tgs);
      setCategoryId(aligned.categoryId);
      setSelectedTagIds(aligned.tagIds);

      const fullContent = computeStoredArticleContent(
        editorBody,
        {
          title,
          slug,
          summary,
          status,
          publishedAt,
          categoryId: aligned.categoryId,
          selectedTagIds: aligned.tagIds,
        },
        cats,
        tgs,
      );
      const payload = {
        title: title.trim(),
        slug: slug.trim() || undefined,
        content: fullContent,
        summary: summary.trim() || undefined,
        status,
        publishedAt: publishedAt.trim() || undefined,
        categoryId: aligned.categoryId ?? undefined,
        tagIds: aligned.tagIds.length > 0 ? aligned.tagIds : undefined,
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
            slug: slug.trim(),
            content: fullContent,
            summary,
            status,
            publishedAt,
            categoryId: aligned.categoryId,
            tagIds: aligned.tagIds,
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

  const executeDeleteVersion = async (versionId: number) => {
    if (!articleId) return;
    setDeletingVersionId(versionId);
    try {
      const res = await deleteArticleVersion(articleId, versionId);
      if (res.code === 200) {
        toast.success('已删除版本');
        setVersionDeleteConfirmId(null);
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
        const split = splitFrontMatter(res.data.content);
        const eb = split ? split.body : res.data.content;
        applyDetailToState(res.data, {
          setTitle,
          setEditorBody,
          setSlug,
          setSummary,
          setStatus,
          setPublishedAt,
          setCategoryId,
          setSelectedTagIds,
        });
        lastSyncedRef.current = stableArticlePayloadSnapshot({
          title: res.data.title,
          slug: res.data.slug || '',
          content: computeSyncedFullContentFromDetail(res.data, eb),
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

  const bytemdEditorConfig = useMemo(
    () => ({ lineNumbers: showLineNumbers }),
    [showLineNumbers],
  );

  const bytemdPlugins = useMemo(
    () => [...articleBytemdPlugins, createNotePropertiesBytemdPlugin()],
    [],
  );

  useLayoutEffect(() => {
    useArticleNotePropsStore.setState({
      showNoteProperties,
      title,
      setTitle,
      slug,
      setSlug,
      summary,
      setSummary,
      status,
      handleStatusSelect,
      statusOptions,
      categoryId,
      setCategoryId,
      categories,
      setCategories,
      tags,
      setTags,
      selectedTagIds,
      setSelectedTagIds,
      toggleTag,
    });
  }, [
    showNoteProperties,
    title,
    slug,
    summary,
    status,
    handleStatusSelect,
    statusOptions,
    categoryId,
    categories,
    tags,
    selectedTagIds,
    toggleTag,
  ]);

  return (
    <div className="article-edit-bytemd flex w-full flex-col gap-4">
      <div className="sticky top-[var(--admin-mobile-header-height)] z-20 flex shrink-0 items-center justify-between border-b border-border/60 bg-background/95 py-1 backdrop-blur supports-[backdrop-filter]:bg-background/80 lg:top-0">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => navigate('/articles')}
            className="rounded p-1.5 text-muted-foreground transition-colors hover:bg-accent"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <h2 className="text-xl font-semibold tracking-tight text-foreground">
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
          <span
            className="hidden text-sm text-muted-foreground sm:inline"
            title="在编辑区「笔记属性」中修改状态"
          >
            {statusOptions.find((o) => o.value === status)?.label ?? status}
          </span>
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

      <div className="flex min-h-[calc(100dvh-12rem)] flex-col gap-0 overflow-x-hidden rounded border border-border bg-card">
        <div className="flex shrink-0 flex-wrap items-center gap-4 border-b border-border px-3 py-2">
          <label className="flex cursor-pointer items-center gap-2 text-sm text-foreground">
            <Switch checked={showLineNumbers} onCheckedChange={setShowLineNumbers} />
            行号
          </label>
          <label className="flex cursor-pointer items-center gap-2 text-sm text-foreground">
            <Switch checked={showNoteProperties} onCheckedChange={setShowNoteProperties} />
            笔记属性
          </label>
        </div>
        <div className="editor-container flex min-h-0 flex-1 flex-col overflow-x-hidden">
          <Editor
            key={`bytemd-${showLineNumbers}-${showNoteProperties}`}
            value={editorBody}
            plugins={bytemdPlugins}
            editorConfig={bytemdEditorConfig}
            onChange={(v) => setEditorBody(v)}
            uploadImages={uploadImages}
          />
        </div>
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

      {historyOpen && articleId && (
        <AdminModal
          title="版本历史"
          onClose={() => {
            setHistoryOpen(false);
            handleCancelEditVersion();
            setVersionDeleteConfirmId(null);
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
                          onClick={() => setVersionDeleteConfirmId(v.id)}
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

      <ConfirmModal
        open={versionDeleteConfirmId !== null}
        title="删除版本"
        message="确定删除该版本记录？此操作不可恢复。"
        confirmLabel="删除"
        confirmVariant="danger"
        loading={
          versionDeleteConfirmId !== null &&
          deletingVersionId !== null &&
          deletingVersionId === versionDeleteConfirmId
        }
        onCancel={() => setVersionDeleteConfirmId(null)}
        onConfirm={() => {
          if (versionDeleteConfirmId === null) return;
          void executeDeleteVersion(versionDeleteConfirmId);
        }}
      />

      <style>{`
        .article-edit-bytemd .editor-container .bytemd {
          border: none;
        }
        .article-edit-bytemd .bytemd-toolbar {
          border-bottom: 1px solid var(--border);
        }
      `}</style>
    </div>
  );
}
