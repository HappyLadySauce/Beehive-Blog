"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { FileText, Loader2, MoreHorizontal, Pencil, Plus, Save, Search, Trash2, X } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import { createCategory, deleteCategory, listCategories, updateCategory } from "@/lib/api/categories";
import { deleteContent, listContents } from "@/lib/api/contents";
import { listTags } from "@/lib/api/tags";
import type { CategoryItem, ContentItem, ListCategoriesResponse, ListContentsResponse, ListTagsResponse, TagItem } from "@/lib/api/types";
import { ToastMessage } from "@/components/toast/ToastProvider";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioSelect } from "./StudioSelect";
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

type ContentListFilters = {
  page: number;
  status: string;
  type: string;
  visibility: string;
  tagID: string;
  categoryID: string;
  search: string;
};

type ContentMetadata = {
  categories: ListCategoriesResponse;
  tags: ListTagsResponse;
};

let contentListInflight: { key: string; promise: Promise<ListContentsResponse> } | null = null;
let contentMetadataInflight: Promise<ContentMetadata> | null = null;

function contentListKey(filters: ContentListFilters) {
  return `${filters.page}\x1e${filters.status}\x1e${filters.type}\x1e${filters.visibility}\x1e${filters.tagID}\x1e${filters.categoryID}\x1e${filters.search}`;
}

function requestContentList(filters: ContentListFilters) {
  return listContents({
    page: filters.page,
    page_size: pageSize,
    search: filters.search || undefined,
    category_id: filters.categoryID ? Number(filters.categoryID) : undefined,
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
  return Promise.all([listTags({ page: 1, page_size: 100 }), listCategories({ page: 1, page_size: 100 })]).then(([tags, categories]) => ({
    categories,
    tags
  }));
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
  { value: "project", label: "项目" }
] as const;

const contentTypeOptions = typeOptions.filter((option) => option.value !== "");

const statusOptions = [
  { value: "", label: "状态：全部" },
  { value: "draft", label: "草稿" },
  { value: "review", label: "审核" },
  { value: "published", label: "已发布" },
  { value: "archived", label: "已归档" }
] as const;

const contentStatusOptions = statusOptions.filter((option) => option.value !== "");

const visibilityOptions = [
  { value: "", label: "可见性：全部" },
  { value: "public", label: "公开" },
  { value: "member", label: "成员" },
  { value: "private", label: "私有" }
] as const;

const visibilityLabelOptions = visibilityOptions.filter((option) => option.value !== "");

const emptyCategoryForm: CategoryFormState = {
  description: "",
  name: "",
  parentId: "",
  slug: "",
  sortOrder: "0"
};

export function StudioContentPage() {
  const [data, setData] = useState<ListContentsResponse | null>(null);
  const [tags, setTags] = useState<TagItem[]>([]);
  const [categories, setCategories] = useState<CategoryItem[]>([]);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [visibilityFilter, setVisibilityFilter] = useState("");
  const [tagFilter, setTagFilter] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [deleteTarget, setDeleteTarget] = useState<ContentItem | null>(null);
  const [categoryFormMode, setCategoryFormMode] = useState<CategoryFormMode>("create");
  const [categoryFormOpen, setCategoryFormOpen] = useState(false);
  const [categoryFormTarget, setCategoryFormTarget] = useState<CategoryItem | null>(null);
  const [categoryForm, setCategoryForm] = useState<CategoryFormState>(emptyCategoryForm);
  const [categoryDeleteTarget, setCategoryDeleteTarget] = useState<CategoryItem | null>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [search]);

  const filters = useMemo(
    () => ({
      page,
      categoryID: categoryFilter,
      search: debouncedSearch,
      status: statusFilter,
      tagID: tagFilter,
      type: typeFilter,
      visibility: visibilityFilter
    }),
    [categoryFilter, debouncedSearch, page, statusFilter, tagFilter, typeFilter, visibilityFilter]
  );

  const tagOptions = useMemo(
    () => [{ value: "", label: "标签：全部" }, ...tags.map((tag) => ({ value: String(tag.id), label: tag.name }))],
    [tags]
  );

  const totalPages = useMemo(() => Math.max(1, Math.ceil((data?.total ?? 0) / pageSize)), [data?.total]);

  const excludedCategoryIDs = useMemo(
    () => (categoryFormTarget ? descendantIDs(categoryFormTarget.id, categories) : new Set<number>()),
    [categories, categoryFormTarget]
  );

  const categoryParentOptions = useMemo(
    () => [
      { value: "", label: "无（顶级分类）" },
      ...categories
        .filter((cat) => {
          if (categoryFormTarget && cat.id === categoryFormTarget.id) return false;
          if (excludedCategoryIDs.has(cat.id)) return false;
          return true;
        })
        .map((cat) => ({ value: String(cat.id), label: cat.name }))
    ],
    [categories, categoryFormTarget, excludedCategoryIDs]
  );

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
        if (active) {
          setTags(result.tags.items);
          setCategories(result.categories.items);
        }
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

  async function refreshContentMetadata() {
    const result = await requestContentMetadata();
    setTags(result.tags.items);
    setCategories(result.categories.items);
  }

  function openCategoryCreate() {
    setCategoryFormMode("create");
    setCategoryFormTarget(null);
    setCategoryForm(emptyCategoryForm);
    setMessage(null);
    setCategoryFormOpen(true);
  }

  function openCategoryEdit(category: CategoryItem) {
    setCategoryFormMode("edit");
    setCategoryFormTarget(category);
    setCategoryForm({
      description: category.description ?? "",
      name: category.name,
      parentId: category.parent_id ? String(category.parent_id) : "",
      slug: category.slug,
      sortOrder: String(category.sort_order)
    });
    setMessage(null);
    setCategoryFormOpen(true);
  }

  function closeCategoryForm() {
    setCategoryFormOpen(false);
    setCategoryFormTarget(null);
  }

  function setCategoryFormField<K extends keyof CategoryFormState>(key: K, value: CategoryFormState[K]) {
    setCategoryForm((current) => ({ ...current, [key]: value }));
  }

  async function onDeleteConfirm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
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

  async function onCategorySubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = categoryForm.name.trim();
    const slug = categoryForm.slug.trim();
    if (!name || !slug) {
      setMessage({ tone: "error", text: "分类名称和 Slug 不能为空。" });
      return;
    }
    const sortOrder = parseInt(categoryForm.sortOrder, 10);
    if (isNaN(sortOrder) || sortOrder < 0) {
      setMessage({ tone: "error", text: "排序必须为非负整数。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      if (categoryFormMode === "create") {
        await createCategory({
          description: categoryForm.description || null,
          name,
          parent_id: categoryForm.parentId ? Number(categoryForm.parentId) : null,
          slug,
          sort_order: sortOrder
        });
        setMessage({ tone: "success", text: "分类已创建。" });
      } else if (categoryFormTarget) {
        await updateCategory(categoryFormTarget.id, {
          description: categoryForm.description || null,
          name,
          parent_id: categoryForm.parentId ? Number(categoryForm.parentId) : null,
          slug,
          sort_order: sortOrder
        });
        setMessage({ tone: "success", text: "分类已更新。" });
      }
      closeCategoryForm();
      await refreshContentMetadata();
      await refreshContents();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onCategoryDeleteConfirm() {
    if (!categoryDeleteTarget) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteCategory(categoryDeleteTarget.id);
      const wasActiveFilter = categoryFilter === String(categoryDeleteTarget.id);
      if (wasActiveFilter) {
        setCategoryFilter("");
        resetPage();
      }
      setCategoryDeleteTarget(null);
      setMessage({ tone: "success", text: "分类已删除。" });
      await refreshContentMetadata();
      const nextFilters = wasActiveFilter ? { ...filters, categoryID: "", page: 1 } : filters;
      const result = await requestContentList(nextFilters);
      setData(result);
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
          <Link className="primary-button" href="/studio/content/new" prefetch={false}>
            <Plus aria-hidden size={18} />
            新建内容
          </Link>
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
        <div className={styles.studioListReservedStrip} data-testid="content-category-reserved-strip">
          <CategoryCard active={!categoryFilter} count={data?.total ?? 0} title="全部" onClick={() => {
            setCategoryFilter("");
            resetPage();
          }} />
          {categories.map((category) => (
            <CategoryCard
              active={categoryFilter === String(category.id)}
              count={category.content_count ?? 0}
              key={category.id}
              title={category.name}
              onEdit={() => openCategoryEdit(category)}
              onClick={() => {
                setCategoryFilter(String(category.id));
                resetPage();
              }}
            />
          ))}
          <button className={styles.categoryCard} type="button" onClick={openCategoryCreate}>
            <span>新建</span>
            <Plus aria-hidden size={18} />
          </button>
        </div>

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
                            <Link className="secondary-button icon-button" href={`/studio/content/${item.id}/edit`} aria-label={`编辑 ${item.title}`} prefetch={false}>
                              <Pencil aria-hidden size={16} />
                            </Link>
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

      {deleteTarget ? renderDeleteModal(deleteTarget, saving, () => setDeleteTarget(null), onDeleteConfirm) : null}
      {categoryFormOpen ? renderCategoryFormModal(
        categoryFormMode,
        categoryForm,
        saving,
        categoryFormTarget,
        closeCategoryForm,
        setCategoryFormField,
        onCategorySubmit,
        categoryParentOptions,
        categoryFormTarget ? () => {
          setCategoryFormOpen(false);
          setCategoryDeleteTarget(categoryFormTarget);
        } : undefined
      ) : null}
      {categoryDeleteTarget ? renderCategoryDeleteModal(categoryDeleteTarget, saving, () => setCategoryDeleteTarget(null), onCategoryDeleteConfirm) : null}
    </>
  );
}

function descendantIDs(root: number, categories: CategoryItem[]): Set<number> {
  const ids = new Set<number>();
  const queue = [root];
  while (queue.length > 0) {
    const current = queue.shift()!;
    for (const category of categories) {
      if (category.parent_id === current && !ids.has(category.id)) {
        ids.add(category.id);
        queue.push(category.id);
      }
    }
  }
  return ids;
}

function CategoryCard({
  active,
  count,
  title,
  onClick,
  onEdit
}: {
  active?: boolean;
  count?: number;
  title: string;
  onClick?: () => void;
  onEdit?: () => void;
}) {
  return (
    <button className={`${styles.categoryCard} ${active ? styles.categoryCardActive : ""}`} type="button" onClick={onClick}>
      <span>{title}</span>
      {typeof count === "number" ? <small>{count}</small> : null}
      {onEdit ? (
        <MoreHorizontal
          aria-hidden
          size={16}
          onClick={(event) => {
            event.stopPropagation();
            onEdit();
          }}
        />
      ) : null}
    </button>
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
  parentOptions: { value: string; label: string }[],
  onDelete?: () => void
) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.modalWide} role="dialog" aria-modal="true" aria-labelledby="content-cat-form-title">
        <div className={styles.modalHeader}>
          <div>
            <h3 id="content-cat-form-title">{mode === "create" ? "新建分类" : `编辑 ${target?.name ?? "分类"}`}</h3>
            <p>分类用于内容主题归类，创建后可在内容编辑页绑定。</p>
          </div>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>

        <form className={styles.formGrid} id="content-category-form" onSubmit={onSubmit}>
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
            <StudioSelect ariaLabel="父级分类" options={parentOptions} value={form.parentId} onChange={(value) => onField("parentId", value)} />
          </label>
          <label className={styles.field}>
            <span>排序</span>
            <input type="number" min={0} value={form.sortOrder} onChange={(event) => onField("sortOrder", event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>描述</span>
            <textarea className={styles.textarea} rows={4} value={form.description} onChange={(event) => onField("description", event.target.value)} />
          </label>
        </form>

        <div className={styles.modalActions}>
          {onDelete ? (
            <button className="danger-button" disabled={saving} type="button" onClick={onDelete}>
              删除分类
            </button>
          ) : null}
          <button className="primary-button" disabled={saving} form="content-category-form" type="submit">
            {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            {mode === "create" ? "创建分类" : "保存分类"}
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function renderCategoryDeleteModal(target: CategoryItem, saving: boolean, onClose: () => void, onConfirm: () => void) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="content-cat-delete-title">
        <div className={styles.modalHeader}>
          <h3 id="content-cat-delete-title">删除分类</h3>
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

function renderDeleteModal(target: ContentItem, saving: boolean, onClose: () => void, onSubmit: (event: FormEvent<HTMLFormElement>) => void) {
  return createPortal(
    <div className={styles.overlay} role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <form className={styles.modal} role="dialog" aria-modal="true" aria-labelledby="content-delete-title" onSubmit={onSubmit}>
        <div className={styles.modalHeader}>
          <h3 id="content-delete-title">删除内容</h3>
          <button className="secondary-button icon-button" type="button" aria-label="关闭" onClick={onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <p>确认删除「{target.title}」？该操作会请求后端执行软删除。</p>
        <div className={styles.modalActions}>
          <button className="secondary-button" disabled={saving} type="button" onClick={onClose}>取消</button>
          <button className="danger-button" disabled={saving} type="submit">
            <Trash2 aria-hidden size={18} />
            删除
          </button>
        </div>
      </form>
    </div>,
    document.body
  );
}

function contentTypeLabel(value: string) {
  return contentTypeOptions.find((option) => option.value === value)?.label ?? value;
}

function contentStatusLabel(value: string) {
  return contentStatusOptions.find((option) => option.value === value)?.label ?? value;
}

function visibilityLabel(value: string) {
  return visibilityLabelOptions.find((option) => option.value === value)?.label ?? value;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
