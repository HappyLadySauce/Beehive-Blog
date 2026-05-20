"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { FileText, Loader2, Pencil, Plus, RefreshCw, Save, Search, Trash2, X } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import {
  createContent,
  deleteContent,
  listContents,
  setContentTags,
  transitionContentStatus,
  updateContent
} from "@/lib/api/contents";
import { listTags } from "@/lib/api/tags";
import type { ContentItem, ListContentsResponse, ListTagsResponse, TagItem } from "@/lib/api/types";
import { ToastMessage } from "@/components/toast/ToastProvider";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

const pageSize = 20;
const searchDebounceMs = 400;

type Message = { tone: "success" | "error"; text: string } | null;
type ContentFormMode = "create" | "edit";

type ContentFormState = {
  type: string;
  title: string;
  slug: string;
  excerpt: string;
  body: string;
  coverAttachmentID: string;
  status: string;
  visibility: string;
  aiAccess: string;
  tagIDs: number[];
};

type ContentListFilters = {
  page: number;
  status: string;
  type: string;
  visibility: string;
  tagID: string;
  search: string;
};

let contentListInflight: { key: string; promise: Promise<ListContentsResponse> } | null = null;
let contentMetadataInflight: Promise<ListTagsResponse> | null = null;

function contentListKey(filters: ContentListFilters) {
  return `${filters.page}\x1e${filters.status}\x1e${filters.type}\x1e${filters.visibility}\x1e${filters.tagID}\x1e${filters.search}`;
}

function requestContentList(filters: ContentListFilters) {
  return listContents({
    page: filters.page,
    page_size: pageSize,
    search: filters.search || undefined,
    status: filters.status || undefined,
    tag_id: filters.tagID ? Number(filters.tagID) : undefined,
    type: filters.type || undefined,
    visibility: filters.visibility || undefined
  });
}

function loadContentList(filters: ContentListFilters) {
  const key = contentListKey(filters);
  if (contentListInflight?.key === key) {
    return contentListInflight.promise;
  }
  const promise = requestContentList(filters).finally(() => {
    if (contentListInflight?.promise === promise) {
      contentListInflight = null;
    }
  });
  contentListInflight = { key, promise };
  return promise;
}

function requestContentMetadata() {
  return listTags({ page: 1, page_size: 100, status: "active" });
}

function loadContentMetadata() {
  if (contentMetadataInflight) {
    return contentMetadataInflight;
  }
  const promise = requestContentMetadata().finally(() => {
    contentMetadataInflight = null;
  });
  contentMetadataInflight = promise;
  return promise;
}

// resetContentPageModuleStateForTests clears module-level request limiters between tests.
// resetContentPageModuleStateForTests 在测试之间清空模块级请求限制器。
export function resetContentPageModuleStateForTests() {
  contentListInflight = null;
  contentMetadataInflight = null;
}

const typeOptions = [
  { value: "", label: "类型：全部" },
  { value: "article", label: "文章" },
  { value: "note", label: "笔记" },
  { value: "project", label: "项目" },
  { value: "experience", label: "经历" },
  { value: "reflection", label: "复盘" },
  { value: "portfolio", label: "作品集" }
] as const;

const contentTypeFormOptions = typeOptions.filter((option) => option.value !== "");

const statusOptions = [
  { value: "", label: "状态：全部" },
  { value: "draft", label: "草稿" },
  { value: "review", label: "审核" },
  { value: "published", label: "已发布" },
  { value: "archived", label: "已归档" }
] as const;

const contentStatusFormOptions = statusOptions.filter((option) => option.value !== "");

const visibilityOptions = [
  { value: "", label: "可见性：全部" },
  { value: "public", label: "公开" },
  { value: "member", label: "成员" },
  { value: "private", label: "私有" }
] as const;

const visibilityFormOptions = visibilityOptions.filter((option) => option.value !== "");

const aiAccessOptions = [
  { value: "allowed", label: "AI 可访问" },
  { value: "denied", label: "AI 禁止" }
] as const;

const emptyForm: ContentFormState = {
  aiAccess: "allowed",
  body: "",
  coverAttachmentID: "",
  excerpt: "",
  slug: "",
  status: "draft",
  tagIDs: [],
  title: "",
  type: "article",
  visibility: "public"
};

export function StudioContentPage() {
  const [data, setData] = useState<ListContentsResponse | null>(null);
  const [tags, setTags] = useState<TagItem[]>([]);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [visibilityFilter, setVisibilityFilter] = useState("");
  const [tagFilter, setTagFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [formMode, setFormMode] = useState<ContentFormMode>("create");
  const [formOpen, setFormOpen] = useState(false);
  const [formTarget, setFormTarget] = useState<ContentItem | null>(null);
  const [form, setForm] = useState<ContentFormState>(emptyForm);
  const [deleteTarget, setDeleteTarget] = useState<ContentItem | null>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [search]);

  const filters = useMemo(
    () => ({
      page,
      search: debouncedSearch,
      status: statusFilter,
      tagID: tagFilter,
      type: typeFilter,
      visibility: visibilityFilter
    }),
    [debouncedSearch, page, statusFilter, tagFilter, typeFilter, visibilityFilter]
  );

  const tagOptions = useMemo(
    () => [{ value: "", label: "标签：全部" }, ...tags.map((tag) => ({ value: String(tag.id), label: tag.name }))],
    [tags]
  );

  const totalPages = useMemo(() => Math.max(1, Math.ceil((data?.total ?? 0) / pageSize)), [data?.total]);

  const refreshContents = useCallback(async () => {
    try {
      const result = await requestContentList(filters);
      setData(result);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    }
  }, [filters]);

  useEffect(() => {
    let active = true;
    loadContentMetadata()
      .then((result) => {
        if (active) setTags(result.items);
      })
      .catch((error: unknown) => {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    loadContentList(filters)
      .then((result) => {
        if (active) setData(result);
      })
      .catch((error: unknown) => {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [filters]);

  function resetPage() {
    setPage(1);
  }

  function openCreate() {
    setFormMode("create");
    setFormTarget(null);
    setForm(emptyForm);
    setMessage(null);
    setFormOpen(true);
  }

  function openEdit(item: ContentItem) {
    setFormMode("edit");
    setFormTarget(item);
    setForm({
      aiAccess: item.ai_access,
      body: item.body ?? "",
      coverAttachmentID: item.cover_attachment_id ? String(item.cover_attachment_id) : "",
      excerpt: item.excerpt ?? "",
      slug: item.slug,
      status: item.status,
      tagIDs: item.tags?.map((tag) => tag.id) ?? [],
      title: item.title,
      type: item.type,
      visibility: item.visibility
    });
    setMessage(null);
    setFormOpen(true);
  }

  function closeForm() {
    setFormOpen(false);
    setFormTarget(null);
  }

  function setFormField<K extends keyof ContentFormState>(key: K, value: ContentFormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function toggleFormTag(id: number, checked: boolean) {
    setForm((current) => ({
      ...current,
      tagIDs: checked ? Array.from(new Set([...current.tagIDs, id])) : current.tagIDs.filter((tagID) => tagID !== id)
    }));
  }

  async function onSubmitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const title = form.title.trim();
    const slug = form.slug.trim();
    if (!title || !slug || !form.type) {
      setMessage({ tone: "error", text: "类型、标题和 Slug 不能为空。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const coverAttachmentID = form.coverAttachmentID ? Number(form.coverAttachmentID) : null;
      if (coverAttachmentID !== null && (!Number.isInteger(coverAttachmentID) || coverAttachmentID <= 0)) {
        setMessage({ tone: "error", text: "封面附件 ID 必须是正整数。" });
        return;
      }

      if (formMode === "create") {
        const created = await createContent({
          ai_access: form.aiAccess,
          body: form.body || null,
          cover_attachment_id: coverAttachmentID,
          excerpt: form.excerpt || null,
          slug,
          status: form.status === "published" || form.status === "archived" ? "draft" : form.status,
          title,
          type: form.type,
          visibility: form.visibility
        });
        if (form.tagIDs.length > 0) {
          await setContentTags(created.id, { tag_ids: form.tagIDs });
        }
        setMessage({ tone: "success", text: "内容已创建。" });
      } else if (formTarget) {
        await updateContent(formTarget.id, {
          ai_access: form.aiAccess,
          body: form.body || null,
          cover_attachment_id: coverAttachmentID,
          excerpt: form.excerpt || null,
          slug,
          title,
          type: form.type,
          visibility: form.visibility
        });
        await setContentTags(formTarget.id, { tag_ids: form.tagIDs });
        if (form.status !== formTarget.status) {
          await transitionContentStatus(formTarget.id, { status: form.status });
        }
        setMessage({ tone: "success", text: "内容已更新。" });
      }

      closeForm();
      await refreshContents();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onDeleteConfirm() {
    if (!deleteTarget) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteContent(deleteTarget.id);
      setDeleteTarget(null);
      setMessage({ tone: "success", text: "内容已删除。" });
      await refreshContents();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <StudioTopbar
        actions={
          <>
            <button className="secondary-button" disabled={loading} type="button" onClick={refreshContents}>
              <RefreshCw aria-hidden size={18} />
              刷新
            </button>
            <button className="primary-button" type="button" onClick={openCreate}>
              <Plus aria-hidden size={18} />
              新建内容
            </button>
          </>
        }
        description="管理文章、笔记、项目与发布状态。"
        eyebrow="Content studio"
        title="内容"
      />

      <ToastMessage message={message} />

      <section className={styles.studioListShell} aria-label="内容工作台">
        <div className={`${styles.filterBar} ${styles.studioListToolbar}`}>
          <div className={styles.searchInput}>
            <Search aria-hidden size={18} />
            <input
              aria-label="搜索内容"
              placeholder="搜索标题或摘要"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                resetPage();
              }}
            />
          </div>
          <StudioSelect ariaLabel="筛选类型" className={styles.filterSelect} options={typeOptions} value={typeFilter} onChange={(value) => { setTypeFilter(value); resetPage(); }} />
          <StudioSelect ariaLabel="筛选状态" className={styles.filterSelect} options={statusOptions} value={statusFilter} onChange={(value) => { setStatusFilter(value); resetPage(); }} />
          <StudioSelect ariaLabel="筛选可见性" className={styles.filterSelect} options={visibilityOptions} value={visibilityFilter} onChange={(value) => { setVisibilityFilter(value); resetPage(); }} />
          <StudioSelect ariaLabel="筛选标签" className={styles.filterSelect} options={tagOptions} value={tagFilter} onChange={(value) => { setTagFilter(value); resetPage(); }} />
        </div>
        <div className={styles.studioListReservedStrip} data-testid="content-category-reserved-strip" aria-hidden="true" />

        {loading ? (
          <div className={styles.emptyState}>
            <Loader2 aria-hidden className="spin" size={28} />
            <strong>正在加载内容...</strong>
          </div>
        ) : (
          <>
            <div className={`${styles.tableScroll} ${styles.studioListTableFrame}`}>
              <table className={`${styles.table} ${styles.userTable}`}>
                <thead>
                  <tr>
                    <th>标题</th>
                    <th>类型</th>
                    <th>状态</th>
                    <th>可见性</th>
                    <th>标签</th>
                    <th>更新时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {data?.items.length ? (
                    data.items.map((item) => (
                      <tr key={item.id}>
                        <td>
                          <strong>{item.title}</strong>
                          <br />
                          <span className={styles.codePill}>{item.slug}</span>
                        </td>
                        <td>{contentTypeLabel(item.type)}</td>
                        <td><span className={styles.statusPill}>{contentStatusLabel(item.status)}</span></td>
                        <td>{visibilityLabel(item.visibility)}</td>
                        <td>{item.tags?.length ? item.tags.map((tag) => tag.name).join("、") : "未绑定"}</td>
                        <td>{formatDate(item.updated_at)}</td>
                        <td>
                          <div className={styles.tableActions}>
                            <button className="secondary-button icon-button" type="button" aria-label={`编辑 ${item.title}`} onClick={() => openEdit(item)}>
                              <Pencil aria-hidden size={16} />
                            </button>
                            <button className="danger-button icon-button" type="button" aria-label={`删除 ${item.title}`} onClick={() => setDeleteTarget(item)}>
                              <Trash2 aria-hidden size={16} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td className={styles.studioListEmptyCell} colSpan={7}>
                        <div className={styles.emptyState}>
                          <FileText aria-hidden size={28} />
                          <strong>暂无内容</strong>
                          <span>创建第一条内容后会显示在这里。</span>
                        </div>
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
            <StudioPagePagination disabled={loading} page={page} totalPages={totalPages} onPageChange={setPage} />
          </>
        )}
      </section>

      {formOpen ? renderContentFormModal(formMode, form, tags, saving, formTarget, closeForm, setFormField, toggleFormTag, onSubmitForm) : null}
      {deleteTarget ? renderDeleteModal(deleteTarget, saving, () => setDeleteTarget(null), onDeleteConfirm) : null}
    </>
  );
}

function renderContentFormModal(
  mode: ContentFormMode,
  form: ContentFormState,
  tags: TagItem[],
  saving: boolean,
  target: ContentItem | null,
  onClose: () => void,
  onField: <K extends keyof ContentFormState>(key: K, value: ContentFormState[K]) => void,
  onTag: (id: number, checked: boolean) => void,
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <form className={styles.modalTall} role="dialog" aria-modal="true" aria-labelledby="content-form-title" onSubmit={onSubmit}>
        <div className={styles.modalHeader}>
          <div>
            <h3 id="content-form-title">{mode === "create" ? "新建内容" : `编辑 ${target?.title ?? "内容"}`}</h3>
            <p>创建时只能直接进入草稿或审核；发布与归档由后端状态流转控制。</p>
          </div>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>

        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>类型</span>
            <StudioSelect ariaLabel="内容类型" options={contentTypeFormOptions} value={form.type} onChange={(value) => onField("type", value)} />
          </label>
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="内容状态" options={contentStatusFormOptions} value={form.status} onChange={(value) => onField("status", value)} />
          </label>
          <label className={styles.field}>
            <span>标题</span>
            <input value={form.title} onChange={(event) => onField("title", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Slug</span>
            <input value={form.slug} onChange={(event) => onField("slug", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>可见性</span>
            <StudioSelect ariaLabel="内容可见性" options={visibilityFormOptions} value={form.visibility} onChange={(value) => onField("visibility", value)} />
          </label>
          <label className={styles.field}>
            <span>AI 访问</span>
            <StudioSelect ariaLabel="AI 访问" options={aiAccessOptions} value={form.aiAccess} onChange={(value) => onField("aiAccess", value)} />
          </label>
          <label className={styles.field}>
            <span>封面附件 ID</span>
            <input inputMode="numeric" value={form.coverAttachmentID} onChange={(event) => onField("coverAttachmentID", event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>摘要</span>
            <textarea className={styles.textarea} rows={3} value={form.excerpt} onChange={(event) => onField("excerpt", event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>正文</span>
            <textarea className={styles.textarea} rows={10} value={form.body} onChange={(event) => onField("body", event.target.value)} />
          </label>
          <fieldset className={`${styles.checkboxGrid} ${styles.fieldFull}`}>
            <legend>标签</legend>
            {tags.length > 0 ? (
              tags.map((tag) => (
                <label key={tag.id}>
                  <span>{tag.name}</span>
                  <input
                    checked={form.tagIDs.includes(tag.id)}
                    type="checkbox"
                    onChange={(event) => onTag(tag.id, event.target.checked)}
                  />
                </label>
              ))
            ) : (
              <span>暂无可用标签</span>
            )}
          </fieldset>
        </div>

        <div className={styles.modalActions}>
          <button className="secondary-button" disabled={saving} type="button" onClick={onClose}>取消</button>
          <button className="primary-button" disabled={saving} type="submit">
            {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            保存
          </button>
        </div>
      </form>
    </div>,
    document.body
  );
}

function renderDeleteModal(target: ContentItem, saving: boolean, onClose: () => void, onConfirm: () => void) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="content-delete-title">
        <div className={styles.modalHeader}>
          <h3 id="content-delete-title">删除内容</h3>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <p>确认删除「{target.title}」？该操作会请求后端执行软删除。</p>
        <div className={styles.modalActions}>
          <button className="secondary-button" disabled={saving} type="button" onClick={onClose}>取消</button>
          <button className="danger-button" disabled={saving} type="button" onClick={onConfirm}>
            <Trash2 aria-hidden size={18} />
            删除
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function contentTypeLabel(value: string) {
  return contentTypeFormOptions.find((option) => option.value === value)?.label ?? value;
}

function contentStatusLabel(value: string) {
  return contentStatusFormOptions.find((option) => option.value === value)?.label ?? value;
}

function visibilityLabel(value: string) {
  return visibilityFormOptions.find((option) => option.value === value)?.label ?? value;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
