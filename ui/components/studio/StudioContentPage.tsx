"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { FileText, Loader2, Pencil, Plus, Search, Trash2, X } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import { deleteContent, listContents } from "@/lib/api/contents";
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
    </>
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
