"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { CheckCircle2, EyeOff, Loader2, MessageSquare, RotateCcw, Search, Trash2 } from "lucide-react";

import { ToastMessage } from "@/components/toast/ToastProvider";
import { deleteAdminComment, listAdminComments, updateAdminCommentStatus } from "@/lib/api/comments";
import { humanizeApiError } from "@/lib/api/client";
import type { AdminCommentItem, ListAdminCommentsResponse, UpdateCommentStatusRequest } from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioSelect } from "./StudioSelect";
import { StudioTopbar } from "./StudioTopbar";

const pageSize = 20;
const searchDebounceMs = 400;

type Message = { tone: "success" | "error"; text: string } | null;
type CommentStatusFilter = "all" | "review" | "published" | "hidden";
type CommentStatus = UpdateCommentStatusRequest["status"];

const statusOptions = [
  { value: "all", label: "全部状态" },
  { value: "review", label: "待审核" },
  { value: "published", label: "已发布" },
  { value: "hidden", label: "已隐藏" }
] as const;

let commentsInflight: { key: string; promise: Promise<ListAdminCommentsResponse> } | null = null;

function commentsKey(page: number, status: CommentStatusFilter, search: string) {
  return `${page}\x1e${status}\x1e${search}`;
}

function requestComments(page: number, status: CommentStatusFilter, search: string) {
  return listAdminComments({
    page,
    page_size: pageSize,
    status: status === "all" ? undefined : status,
    search: search || undefined
  });
}

function loadComments(page: number, status: CommentStatusFilter, search: string) {
  const key = commentsKey(page, status, search);
  if (commentsInflight?.key === key) {
    return commentsInflight.promise;
  }
  const promise = requestComments(page, status, search).finally(() => {
    if (commentsInflight?.promise === promise) {
      commentsInflight = null;
    }
  });
  commentsInflight = { key, promise };
  return promise;
}

// resetCommentsPageModuleStateForTests clears request dedupe state between unit tests.
// resetCommentsPageModuleStateForTests 在单元测试之间清理请求去重状态。
export function resetCommentsPageModuleStateForTests() {
  commentsInflight = null;
}

export function StudioCommentsPage() {
  const [data, setData] = useState<ListAdminCommentsResponse | null>(null);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<CommentStatusFilter>("all");
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [actingId, setActingId] = useState<number | null>(null);
  const [message, setMessage] = useState<Message>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search.trim()), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [search]);

  const totalPages = useMemo(() => Math.max(1, Math.ceil((data?.total ?? 0) / pageSize)), [data?.total]);

  const refresh = useCallback(async () => {
    const result = await requestComments(page, status, debouncedSearch);
    setData(result);
  }, [debouncedSearch, page, status]);

  useEffect(() => {
    let active = true;
    loadComments(page, status, debouncedSearch)
      .then((result) => {
        if (active) setData(result);
      })
      .catch((error) => {
        if (active) setMessage({ tone: "error", text: humanizeApiError(error) });
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [debouncedSearch, page, status]);

  async function setCommentStatus(comment: AdminCommentItem, nextStatus: CommentStatus) {
    setActingId(comment.id);
    setMessage(null);
    try {
      await updateAdminCommentStatus(comment.id, { status: nextStatus });
      setMessage({ tone: "success", text: "评论状态已更新。" });
      await refresh();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setActingId(null);
    }
  }

  async function deleteComment(comment: AdminCommentItem) {
    if (!window.confirm(`确认删除 ${comment.nickname} 的评论？`)) {
      return;
    }
    setActingId(comment.id);
    setMessage(null);
    try {
      await deleteAdminComment(comment.id);
      setMessage({ tone: "success", text: "评论已删除。" });
      await refresh();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setActingId(null);
    }
  }

  return (
    <>
      <StudioTopbar
        description="审核公开评论、隐藏争议内容并保留待审队列，前台只展示已发布评论。"
        eyebrow="Community moderation"
        title="评论"
      />

      <ToastMessage message={message} />

      <section className={styles.studioListShell} aria-label="评论管理">
        <div className={`${styles.filterBar} ${styles.studioListToolbar}`}>
          <div className={styles.searchInput}>
            <Search aria-hidden size={18} />
            <input
              aria-label="搜索评论"
              placeholder="搜索昵称、正文或文章"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(1);
              }}
            />
          </div>
          <StudioSelect
            ariaLabel="评论状态"
            className={styles.filterSelect}
            options={statusOptions}
            value={status}
            onChange={(value) => {
              setStatus(value as CommentStatusFilter);
              setPage(1);
            }}
          />
        </div>

        {loading ? (
          <div className={styles.emptyState}>
            <Loader2 aria-hidden className="spin" size={28} />
            <strong>正在加载评论...</strong>
          </div>
        ) : data?.items.length ? (
          <>
            <div className={`${styles.tableScroll} ${styles.studioListTableFrame}`}>
              <table className={`${styles.table} ${styles.userTable}`}>
                <thead>
                  <tr>
                    <th>评论</th>
                    <th>内容</th>
                    <th>状态</th>
                    <th>时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {data.items.map((comment) => (
                    <tr key={comment.id}>
                      <td>
                        <strong>{comment.nickname}</strong>
                        {comment.website ? <p><a className="text-link" href={comment.website} rel="noreferrer" target="_blank">{comment.website}</a></p> : null}
                        <p>{comment.body}</p>
                      </td>
                      <td>
                        <strong>{comment.content_title || "未命名内容"}</strong>
                        <p className={styles.muted}>#{comment.content_id} / {comment.content_slug || "-"}</p>
                      </td>
                      <td>{renderStatus(comment.status)}</td>
                      <td>{formatDate(comment.created_at)}</td>
                      <td>
                        <div className={styles.tableActions}>
                          <button
                            className="secondary-button icon-button"
                            disabled={actingId === comment.id || comment.status === "published"}
                            type="button"
                            aria-label="通过评论"
                            onClick={() => setCommentStatus(comment, "published")}
                          >
                            <CheckCircle2 aria-hidden size={16} />
                          </button>
                          <button
                            className="secondary-button icon-button"
                            disabled={actingId === comment.id || comment.status === "review"}
                            type="button"
                            aria-label="恢复待审"
                            onClick={() => setCommentStatus(comment, "review")}
                          >
                            <RotateCcw aria-hidden size={16} />
                          </button>
                          <button
                            className="secondary-button icon-button"
                            disabled={actingId === comment.id || comment.status === "hidden"}
                            type="button"
                            aria-label="隐藏评论"
                            onClick={() => setCommentStatus(comment, "hidden")}
                          >
                            <EyeOff aria-hidden size={16} />
                          </button>
                          <button
                            className="danger-button icon-button"
                            disabled={actingId === comment.id}
                            type="button"
                            aria-label="删除评论"
                            onClick={() => deleteComment(comment)}
                          >
                            {actingId === comment.id ? <Loader2 aria-hidden className="spin" size={16} /> : <Trash2 aria-hidden size={16} />}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <StudioPagePagination disabled={loading} page={page} totalPages={totalPages} onPageChange={setPage} />
          </>
        ) : (
          <div className={styles.emptyState}>
            <MessageSquare aria-hidden size={28} />
            <strong>暂无评论</strong>
            <span>公开页收到评论后会进入待审核队列。</span>
          </div>
        )}
      </section>
    </>
  );
}

function renderStatus(status: string) {
  const label = status === "published" ? "已发布" : status === "hidden" ? "已隐藏" : "待审核";
  const className = status === "published" ? styles.statusReady : status === "hidden" ? styles.statusPending : styles.statusPill;
  return <span className={className}>{label}</span>;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
