"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { Loader2, Pencil, Plus, RefreshCw, Save, Search, Tags, Trash2, X } from "lucide-react";

import { ToastMessage } from "@/components/toast/ToastProvider";
import { humanizeApiError } from "@/lib/api/client";
import { createTag, deleteTag, listTags, updateTag } from "@/lib/api/tags";
import type { ListTagsResponse, TagItem } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

const pageSize = 20;
const searchDebounceMs = 400;

type Message = { tone: "success" | "error"; text: string } | null;
type TagFormMode = "create" | "edit";

type TagFormState = {
  name: string;
  slug: string;
  description: string;
  color: string;
  status: string;
};

type TagListFilters = {
  page: number;
  status: string;
  search: string;
};

let tagsListInflight: { key: string; promise: Promise<ListTagsResponse> } | null = null;

function tagsListKey(filters: TagListFilters) {
  return `${filters.page}\x1e${filters.status}\x1e${filters.search}`;
}

function requestTagsList(filters: TagListFilters) {
  return listTags({
    page: filters.page,
    page_size: pageSize,
    search: filters.search || undefined,
    status: filters.status || undefined
  });
}

function loadTagsList(filters: TagListFilters) {
  const key = tagsListKey(filters);
  if (tagsListInflight?.key === key) {
    return tagsListInflight.promise;
  }
  const promise = requestTagsList(filters).finally(() => {
    if (tagsListInflight?.promise === promise) {
      tagsListInflight = null;
    }
  });
  tagsListInflight = { key, promise };
  return promise;
}

// resetTagsPageModuleStateForTests clears module-level request limiters between tests.
// resetTagsPageModuleStateForTests 在测试之间清空模块级请求限制器。
export function resetTagsPageModuleStateForTests() {
  tagsListInflight = null;
}

const statusOptions = [
  { value: "", label: "状态：全部" },
  { value: "active", label: "启用" },
  { value: "archived", label: "归档" }
] as const;

const statusFormOptions = statusOptions.filter((option) => option.value !== "");

const emptyForm: TagFormState = {
  color: "",
  description: "",
  name: "",
  slug: "",
  status: "active"
};

export function StudioTagsPage() {
  const [data, setData] = useState<ListTagsResponse | null>(null);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [formMode, setFormMode] = useState<TagFormMode>("create");
  const [formOpen, setFormOpen] = useState(false);
  const [formTarget, setFormTarget] = useState<TagItem | null>(null);
  const [form, setForm] = useState<TagFormState>(emptyForm);
  const [deleteTarget, setDeleteTarget] = useState<TagItem | null>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [search]);

  const filters = useMemo(
    () => ({
      page,
      search: debouncedSearch,
      status: statusFilter
    }),
    [debouncedSearch, page, statusFilter]
  );

  const totalPages = useMemo(() => Math.max(1, Math.ceil((data?.total ?? 0) / pageSize)), [data?.total]);

  const refreshTags = useCallback(async () => {
    try {
      const result = await requestTagsList(filters);
      setData(result);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    }
  }, [filters]);

  useEffect(() => {
    let active = true;
    loadTagsList(filters)
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

  function openEdit(tag: TagItem) {
    setFormMode("edit");
    setFormTarget(tag);
    setForm({
      color: tag.color ?? "",
      description: tag.description ?? "",
      name: tag.name,
      slug: tag.slug,
      status: tag.status ?? "active"
    });
    setMessage(null);
    setFormOpen(true);
  }

  function closeForm() {
    setFormOpen(false);
    setFormTarget(null);
  }

  function setFormField<K extends keyof TagFormState>(key: K, value: TagFormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function onSubmitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = form.name.trim();
    const slug = form.slug.trim();
    if (!name || !slug) {
      setMessage({ tone: "error", text: "标签名称和 Slug 不能为空。" });
      return;
    }
    const color = form.color.trim();
    if (color && !/^#[0-9a-fA-F]{6}$/.test(color)) {
      setMessage({ tone: "error", text: "颜色必须使用 #RRGGBB 格式。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      if (formMode === "create") {
        await createTag({
          color: color || null,
          description: form.description || null,
          name,
          slug
        });
        setMessage({ tone: "success", text: "标签已创建。" });
      } else if (formTarget) {
        await updateTag(formTarget.id, {
          color: color || null,
          description: form.description || null,
          name,
          slug,
          status: form.status
        });
        setMessage({ tone: "success", text: "标签已更新。" });
      }
      closeForm();
      await refreshTags();
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
      await deleteTag(deleteTarget.id);
      setDeleteTarget(null);
      setMessage({ tone: "success", text: "标签已删除。" });
      await refreshTags();
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
            <button className="secondary-button" disabled={loading} type="button" onClick={refreshTags}>
              <RefreshCw aria-hidden size={18} />
              刷新
            </button>
            <button className="primary-button" type="button" onClick={openCreate}>
              <Plus aria-hidden size={18} />
              新建标签
            </button>
          </>
        }
        description="统一整理标签、专题与内容关系，避免 Public 与 Studio 信息架构分裂。"
        eyebrow="Content taxonomy"
        title="标签"
      />

      <ToastMessage message={message} />

      <section className={styles.studioListShell} aria-label="标签管理">
        <div className={`${styles.filterBar} ${styles.studioListToolbar}`}>
          <div className={styles.searchInput}>
            <Search aria-hidden size={18} />
            <input
              aria-label="搜索标签"
              placeholder="搜索名称或 Slug"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                resetPage();
              }}
            />
          </div>
          <StudioSelect
            ariaLabel="筛选标签状态"
            className={styles.filterSelect}
            options={statusOptions}
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value);
              resetPage();
            }}
          />
        </div>

        {loading ? (
          <div className={styles.emptyState}>
            <Loader2 aria-hidden className="spin" size={28} />
            <strong>正在加载标签...</strong>
          </div>
        ) : (
          <>
            <div className={`${styles.tableScroll} ${styles.studioListTableFrame}`}>
              <table className={`${styles.table} ${styles.userTable}`}>
                <thead>
                  <tr>
                    <th>标签</th>
                    <th>Slug</th>
                    <th>状态</th>
                    <th>内容数</th>
                    <th>更新时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {data?.items.length ? (
                    data.items.map((tag) => (
                      <tr key={tag.id}>
                        <td>
                          <span aria-hidden className={styles.statusDot} style={{ background: tag.color || "#d7efe8" }} />
                          <strong>{tag.name}</strong>
                          {tag.description ? <p>{tag.description}</p> : null}
                        </td>
                        <td><span className={styles.codePill}>{tag.slug}</span></td>
                        <td><span className={styles.statusPill}>{tagStatusLabel(tag.status ?? "active")}</span></td>
                        <td>{tag.content_count ?? 0}</td>
                        <td>{formatDate(tag.updated_at)}</td>
                        <td>
                          <div className={styles.tableActions}>
                            <button className="secondary-button icon-button" type="button" aria-label={`编辑 ${tag.name}`} onClick={() => openEdit(tag)}>
                              <Pencil aria-hidden size={16} />
                            </button>
                            <button className="danger-button icon-button" type="button" aria-label={`删除 ${tag.name}`} onClick={() => setDeleteTarget(tag)}>
                              <Trash2 aria-hidden size={16} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td className={styles.studioListEmptyCell} colSpan={6}>
                        <div className={styles.emptyState}>
                          <Tags aria-hidden size={28} />
                          <strong>暂无标签</strong>
                          <span>创建标签后可在内容编辑里绑定。</span>
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

      {formOpen ? renderTagFormModal(formMode, form, saving, formTarget, closeForm, setFormField, onSubmitForm) : null}
      {deleteTarget ? renderDeleteModal(deleteTarget, saving, () => setDeleteTarget(null), onDeleteConfirm) : null}
    </>
  );
}

function renderTagFormModal(
  mode: TagFormMode,
  form: TagFormState,
  saving: boolean,
  target: TagItem | null,
  onClose: () => void,
  onField: <K extends keyof TagFormState>(key: K, value: TagFormState[K]) => void,
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <form className={styles.modalWide} role="dialog" aria-modal="true" aria-labelledby="tag-form-title" onSubmit={onSubmit}>
        <div className={styles.modalHeader}>
          <div>
            <h3 id="tag-form-title">{mode === "create" ? "新建标签" : `编辑 ${target?.name ?? "标签"}`}</h3>
            <p>Slug 用于内容筛选和未来公开路由，请保持稳定、简短、可读。</p>
          </div>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>

        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>名称</span>
            <input value={form.name} onChange={(event) => onField("name", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Slug</span>
            <input value={form.slug} onChange={(event) => onField("slug", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>颜色</span>
            <input placeholder="#2f8f79" value={form.color} onChange={(event) => onField("color", event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="标签状态" disabled={mode === "create"} options={statusFormOptions} value={form.status} onChange={(value) => onField("status", value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>描述</span>
            <textarea className={styles.textarea} rows={4} value={form.description} onChange={(event) => onField("description", event.target.value)} />
          </label>
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

function renderDeleteModal(target: TagItem, saving: boolean, onClose: () => void, onConfirm: () => void) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="tag-delete-title">
        <div className={styles.modalHeader}>
          <h3 id="tag-delete-title">删除标签</h3>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <p>确认删除「{target.name}」？如果已有内容引用，后端会拒绝删除并返回冲突原因。</p>
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

function tagStatusLabel(value: string) {
  return statusFormOptions.find((option) => option.value === value)?.label ?? value;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
