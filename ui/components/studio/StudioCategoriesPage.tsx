"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { List, Loader2, Pencil, Plus, Save, Search, Trash2, X } from "lucide-react";

import { ToastMessage } from "@/components/toast/ToastProvider";
import { humanizeApiError } from "@/lib/api/client";
import { createCategory, deleteCategory, listCategories, updateCategory } from "@/lib/api/categories";
import type { CategoryItem, ListCategoriesResponse } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioTopbar } from "./StudioTopbar";

const pageSize = 20;
const searchDebounceMs = 400;

type Message = { tone: "success" | "error"; text: string } | null;
type CategoryFormMode = "create" | "edit";

type CategoryFormState = {
  name: string;
  slug: string;
  description: string;
  parentId: string;
  sortOrder: string;
};

type CategoryListFilters = {
  page: number;
  search: string;
};

let catListInflight: { key: string; promise: Promise<ListCategoriesResponse> } | null = null;

function catListKey(filters: CategoryListFilters) {
  return `${filters.page}\x1e${filters.search}`;
}

function requestCategoriesList(filters: CategoryListFilters) {
  return listCategories({
    page: filters.page,
    page_size: pageSize,
    search: filters.search || undefined
  });
}

function loadCategoriesList(filters: CategoryListFilters) {
  const key = catListKey(filters);
  if (catListInflight?.key === key) {
    return catListInflight.promise;
  }
  const promise = requestCategoriesList(filters).finally(() => {
    if (catListInflight?.promise === promise) {
      catListInflight = null;
    }
  });
  catListInflight = { key, promise };
  return promise;
}

// resetCategoriesPageModuleStateForTests clears module-level request limiters between tests.
// resetCategoriesPageModuleStateForTests 在测试之间清空模块级请求限制器。
export function resetCategoriesPageModuleStateForTests() {
  catListInflight = null;
}

const emptyForm: CategoryFormState = {
  description: "",
  name: "",
  parentId: "",
  slug: "",
  sortOrder: "0"
};

export function StudioCategoriesPage() {
  const [data, setData] = useState<ListCategoriesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [formMode, setFormMode] = useState<CategoryFormMode>("create");
  const [formOpen, setFormOpen] = useState(false);
  const [formTarget, setFormTarget] = useState<CategoryItem | null>(null);
  const [form, setForm] = useState<CategoryFormState>(emptyForm);
  const [deleteTarget, setDeleteTarget] = useState<CategoryItem | null>(null);
  const [allCategories, setAllCategories] = useState<CategoryItem[]>([]);

  // Load all categories (high page size) for parent dropdown.
  // 加载全量分类用于父级下拉选择。
  useEffect(() => {
    listCategories({ page_size: 100 })
      .then((result) => setAllCategories(result.items))
      .catch(() => { /* non-critical */ });
  }, [data?.total]);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [search]);

  const filters = useMemo(
    () => ({
      page,
      search: debouncedSearch
    }),
    [debouncedSearch, page]
  );

  const totalPages = useMemo(() => Math.max(1, Math.ceil((data?.total ?? 0) / pageSize)), [data?.total]);

  const refreshCategories = useCallback(async () => {
    try {
      const result = await requestCategoriesList(filters);
      setData(result);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    }
  }, [filters]);

  useEffect(() => {
    let active = true;
    loadCategoriesList(filters)
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

  function openEdit(cat: CategoryItem) {
    setFormMode("edit");
    setFormTarget(cat);
    setForm({
      description: cat.description ?? "",
      name: cat.name,
      parentId: cat.parent_id ? String(cat.parent_id) : "",
      slug: cat.slug,
      sortOrder: String(cat.sort_order)
    });
    setMessage(null);
    setFormOpen(true);
  }

  function closeForm() {
    setFormOpen(false);
    setFormTarget(null);
  }

  function setFormField<K extends keyof CategoryFormState>(key: K, value: CategoryFormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  // get descendant IDs for the given category to exclude from parent dropdown.
  // 获取指定分类的后代 ID，以便在父级下拉中排除。
  function descendantIDs(root: number, cats: CategoryItem[]): Set<number> {
    const ids = new Set<number>();
    const queue = [root];
    while (queue.length > 0) {
      const current = queue.shift()!;
      for (const c of cats) {
        if (c.parent_id === current && !ids.has(c.id)) {
          ids.add(c.id);
          queue.push(c.id);
        }
      }
    }
    return ids;
  }

  async function onSubmitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = form.name.trim();
    const slug = form.slug.trim();
    if (!name || !slug) {
      setMessage({ tone: "error", text: "分类名称和 Slug 不能为空。" });
      return;
    }
    const sortOrder = parseInt(form.sortOrder, 10);
    if (isNaN(sortOrder) || sortOrder < 0) {
      setMessage({ tone: "error", text: "排序必须为非负整数。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      if (formMode === "create") {
        await createCategory({
          description: form.description || null,
          name,
          parent_id: form.parentId ? Number(form.parentId) : null,
          slug,
          sort_order: sortOrder
        });
        setMessage({ tone: "success", text: "分类已创建。" });
      } else if (formTarget) {
        await updateCategory(formTarget.id, {
          description: form.description || null,
          name,
          parent_id: form.parentId ? Number(form.parentId) : null,
          slug,
          sort_order: sortOrder
        });
        setMessage({ tone: "success", text: "分类已更新。" });
      }
      closeForm();
      await refreshCategories();
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
      await deleteCategory(deleteTarget.id);
      setDeleteTarget(null);
      setMessage({ tone: "success", text: "分类已删除。" });
      await refreshCategories();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  // Build parent options: exclude self (on edit) and descendants.
  // 构建父级选项：编辑时排除自身及后代。
  const excludedIDs = formTarget ? descendantIDs(formTarget.id, allCategories) : new Set<number>();
  const parentOptions = allCategories.filter((c) => {
    if (formTarget && c.id === formTarget.id) return false;
    if (excludedIDs.has(c.id)) return false;
    return true;
  });

  function parentName(parentID: number | null | undefined) {
    if (!parentID) return "—";
    const parent = allCategories.find((c) => c.id === parentID);
    return parent ? parent.name : "—";
  }

  return (
    <>
      <StudioTopbar
        actions={
          <button className="primary-button" type="button" onClick={openCreate}>
            <Plus aria-hidden size={18} />
            新建分类
          </button>
        }
        description="管理内容分类与层级关系，用于内容主题归类和前端导航。"
        eyebrow="Content taxonomy"
        title="分类"
      />

      <ToastMessage message={message} />

      <section className={styles.studioListShell} aria-label="分类管理">
        <div className={`${styles.filterBar} ${styles.studioListToolbar}`}>
          <div className={styles.searchInput}>
            <Search aria-hidden size={18} />
            <input
              aria-label="搜索分类"
              placeholder="搜索名称或 Slug"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                resetPage();
              }}
            />
          </div>
        </div>

        {loading ? (
          <div className={styles.emptyState}>
            <Loader2 aria-hidden className="spin" size={28} />
            <strong>正在加载分类...</strong>
          </div>
        ) : (
          <>
            <div className={`${styles.tableScroll} ${styles.studioListTableFrame}`}>
              <table className={`${styles.table} ${styles.userTable}`}>
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>Slug</th>
                    <th>父级分类</th>
                    <th>内容数</th>
                    <th>排序</th>
                    <th>更新时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {data?.items.length ? (
                    data.items.map((cat) => (
                      <tr key={cat.id}>
                        <td>
                          <strong>{cat.name}</strong>
                          {cat.description ? <p>{cat.description}</p> : null}
                        </td>
                        <td><span className={styles.codePill}>{cat.slug}</span></td>
                        <td>{parentName(cat.parent_id)}</td>
                        <td>{cat.content_count ?? 0}</td>
                        <td>{cat.sort_order}</td>
                        <td>{formatDate(cat.updated_at)}</td>
                        <td>
                          <div className={styles.tableActions}>
                            <button className="secondary-button icon-button" type="button" aria-label={`编辑 ${cat.name}`} onClick={() => openEdit(cat)}>
                              <Pencil aria-hidden size={16} />
                            </button>
                            <button className="danger-button icon-button" type="button" aria-label={`删除 ${cat.name}`} onClick={() => setDeleteTarget(cat)}>
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
                          <List aria-hidden size={28} />
                          <strong>暂无分类</strong>
                          <span>创建分类后可在内容编辑里绑定。</span>
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

      {formOpen ? renderCategoryFormModal(formMode, form, saving, formTarget, closeForm, setFormField, onSubmitForm, parentOptions) : null}
      {deleteTarget ? renderDeleteModal(deleteTarget, saving, () => setDeleteTarget(null), onDeleteConfirm) : null}
    </>
  );
}

function renderCategoryFormModal(
  mode: CategoryFormMode,
  form: CategoryFormState,
  saving: boolean,
  target: CategoryItem | null,
  onClose: () => void,
  onField: <K extends keyof CategoryFormState>(key: K, value: CategoryFormState[K]) => void,
  onSubmit: (event: FormEvent<HTMLFormElement>) => void,
  parentOptions: CategoryItem[]
) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <form className={styles.modalWide} role="dialog" aria-modal="true" aria-labelledby="cat-form-title" onSubmit={onSubmit}>
        <div className={styles.modalHeader}>
          <div>
            <h3 id="cat-form-title">{mode === "create" ? "新建分类" : `编辑 ${target?.name ?? "分类"}`}</h3>
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
            <span>父级分类</span>
            <select value={form.parentId} onChange={(event) => onField("parentId", event.target.value)}>
              <option value="">无（顶级分类）</option>
              {parentOptions.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </label>
          <label className={styles.field}>
            <span>排序</span>
            <input type="number" min={0} value={form.sortOrder} onChange={(event) => onField("sortOrder", event.target.value)} />
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

function renderDeleteModal(target: CategoryItem, saving: boolean, onClose: () => void, onConfirm: () => void) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="cat-delete-title">
        <div className={styles.modalHeader}>
          <h3 id="cat-delete-title">删除分类</h3>
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

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
