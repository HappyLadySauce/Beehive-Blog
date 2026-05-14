"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { Loader2, Pencil, Plus, Save, Trash2, Users, X } from "lucide-react";

import { humanizeApiError } from "@/lib/api/client";
import type { ListUsersResponse, UpdateUserRequest, UserItem } from "@/lib/api/types";
import { createUser, deleteUser, listUsers, updateUser } from "@/lib/api/users";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioTopbar } from "./StudioTopbar";

const defaultPageSize = 20;

const statusOptions = [
  { value: "", label: "全部状态" },
  { value: "active", label: "活跃" },
  { value: "disabled", label: "已禁用" },
  { value: "locked", label: "已锁定" },
  { value: "pending", label: "待激活" }
] as const;

const roleOptions = [
  { value: "", label: "全部角色" },
  { value: "member", label: "成员" },
  { value: "admin", label: "管理员" }
] as const;

// Dedupe identical in-flight list requests (e.g. React Strict Mode remount in dev).
// 对相同查询参数的进行中列表请求去重（例如开发环境 React Strict Mode 重复挂载）。
let usersListInflight: { key: string; promise: Promise<ListUsersResponse> } | null = null;

function usersListKey(page: number, statusFilter: string, roleFilter: string, search: string) {
  return `${page}\x1e${statusFilter}\x1e${roleFilter}\x1e${search}`;
}

function requestUsersList(page: number, statusFilter: string, roleFilter: string, search: string) {
  return listUsers({
    page,
    page_size: defaultPageSize,
    status: statusFilter || undefined,
    role: roleFilter || undefined,
    search: search || undefined
  });
}

function loadUsersList(page: number, statusFilter: string, roleFilter: string, search: string) {
  const key = usersListKey(page, statusFilter, roleFilter, search);
  if (usersListInflight?.key === key) {
    return usersListInflight.promise;
  }
  const promise = requestUsersList(page, statusFilter, roleFilter, search).finally(() => {
    if (usersListInflight?.promise === promise) {
      usersListInflight = null;
    }
  });
  usersListInflight = { key, promise };
  return promise;
}

export function StudioUsersPage() {
  const [data, setData] = useState<ListUsersResponse | null>(null);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  // Modal state
  const [showForm, setShowForm] = useState(false);
  const [editingUser, setEditingUser] = useState<UserItem | null>(null);
  const [showDelete, setShowDelete] = useState<UserItem | null>(null);

  // Form state
  const [formUsername, setFormUsername] = useState("");
  const [formPassword, setFormPassword] = useState("");
  const [formEmail, setFormEmail] = useState("");
  const [formNickname, setFormNickname] = useState("");
  const [formPhone, setFormPhone] = useState("");
  const [formRole, setFormRole] = useState("member");
  const [formStatus, setFormStatus] = useState("active");

  const fetchUsers = useCallback(async () => {
    try {
      // Always hit the network after mutations; do not share the in-flight dedupe slot.
      // 创建/更新/删除后必须重新拉取列表，不走进行中去重，避免拿到陈旧 Promise。
      const result = await requestUsersList(page, statusFilter, roleFilter, search);
      setData(result);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter, roleFilter, search]);

  useEffect(() => {
    let active = true;

    loadUsersList(page, statusFilter, roleFilter, search)
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
  }, [page, statusFilter, roleFilter, search]);

  function openCreate() {
    setEditingUser(null);
    setFormUsername("");
    setFormPassword("");
    setFormEmail("");
    setFormNickname("");
    setFormPhone("");
    setFormRole("member");
    setFormStatus("active");
    setShowForm(true);
  }

  function openEdit(user: UserItem) {
    setEditingUser(user);
    setFormUsername(user.username);
    setFormPassword("");
    setFormEmail(user.email ?? "");
    setFormNickname(user.nickname ?? "");
    setFormPhone(user.phone ?? "");
    setFormRole(user.role);
    setFormStatus(user.status);
    setShowForm(true);
  }

  function closeForm() {
    setShowForm(false);
    setEditingUser(null);
  }

  async function onSubmitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (formUsername.trim() === "") {
      setMessage({ tone: "error", text: "用户名不能为空。" });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      if (editingUser) {
        const payload: UpdateUserRequest = {};
        if (formUsername.trim() !== editingUser.username) {
          payload.username = formUsername.trim();
        }
        if (formPassword !== "") payload.password = formPassword;
        if (formEmail !== (editingUser.email ?? "")) payload.email = formEmail || null;
        if (formNickname !== (editingUser.nickname ?? "")) payload.nickname = formNickname || null;
        if (formPhone !== (editingUser.phone ?? "")) payload.phone = formPhone || null;
        if (formRole !== editingUser.role) payload.role = formRole;
        if (formStatus !== editingUser.status) payload.status = formStatus;
        await updateUser(editingUser.id, payload);
        setMessage({ tone: "success", text: "用户已更新。" });
      } else {
        await createUser({
          username: formUsername.trim(),
          password: formPassword || undefined,
          email: formEmail || undefined,
          nickname: formNickname || undefined,
          phone: formPhone || undefined,
          role: formRole,
          status: formStatus
        });
        setMessage({ tone: "success", text: "用户已创建。" });
      }
      closeForm();
      await fetchUsers();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onDeleteConfirm() {
    if (!showDelete) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteUser(showDelete.id);
      setShowDelete(null);
      setMessage({ tone: "success", text: `用户 ${showDelete.username} 已删除。` });
      await fetchUsers();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;

  return (
    <>
      <StudioTopbar
        actions={
          <button className="primary-button" type="button" onClick={openCreate}>
            <Plus aria-hidden size={18} />
            创建用户
          </button>
        }
        description="管理平台注册用户，支持创建、编辑、筛选和删除。"
        eyebrow="User management"
        title="用户"
      />

      {/* Filter bar */}
      <div className={styles.filterBar}>
        <div className={styles.searchInput}>
          <input
            aria-label="搜索用户名或邮箱"
            placeholder="搜索用户名或邮箱..."
            type="search"
            value={search}
            onChange={(event) => {
              setLoading(true);
              setMessage(null);
              setSearch(event.target.value);
              setPage(1);
            }}
          />
        </div>
        <select
          aria-label="按状态筛选"
          value={statusFilter}
          onChange={(event) => {
            setLoading(true);
            setMessage(null);
            setStatusFilter(event.target.value);
            setPage(1);
          }}
        >
          {statusOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <select
          aria-label="按角色筛选"
          value={roleFilter}
          onChange={(event) => {
            setLoading(true);
            setMessage(null);
            setRoleFilter(event.target.value);
            setPage(1);
          }}
        >
          {roleOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>

      <StudioPanel title="用户列表">
        {loading ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            <strong>正在加载用户列表...</strong>
          </div>
        ) : !data || data.items.length === 0 ? (
          <div className={styles.emptyState}>
            <Users aria-hidden size={28} />
            <strong>暂无用户</strong>
            <span>{message?.text ?? "还没有任何用户记录。"}</span>
          </div>
        ) : (
          <>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>用户名</th>
                  <th>邮箱</th>
                  <th>角色</th>
                  <th>状态</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td>
                    <td>{user.username}</td>
                    <td>{user.email ?? "—"}</td>
                    <td>
                      <span className={`${styles.statusPill} ${user.role === "admin" ? styles.statusReady : styles.statusPending}`}>
                        {user.role === "admin" ? "管理员" : "成员"}
                      </span>
                    </td>
                    <td>
                      <span className={`${styles.statusPill} ${statusPillTone(user.status)}`}>
                        {statusLabel(user.status)}
                      </span>
                    </td>
                    <td>{formatDate(user.created_at)}</td>
                    <td>
                      <div className={styles.tableActions}>
                        <button
                          aria-label={`编辑 ${user.username}`}
                          className="secondary-button"
                          type="button"
                          onClick={() => openEdit(user)}
                        >
                          <Pencil aria-hidden size={16} />
                        </button>
                        <button
                          aria-label={`删除 ${user.username}`}
                          className="secondary-button"
                          type="button"
                          onClick={() => setShowDelete(user)}
                        >
                          <Trash2 aria-hidden size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Pagination */}
            <div className={styles.pagination}>
              <button
                className="secondary-button"
                disabled={page <= 1}
                type="button"
                onClick={() => {
                  setLoading(true);
                  setMessage(null);
                  setPage((p) => Math.max(1, p - 1));
                }}
              >
                上一页
              </button>
              <span>
                第 {data.page} / {totalPages} 页（共 {data.total} 条）
              </span>
              <button
                className="secondary-button"
                disabled={page >= totalPages}
                type="button"
                onClick={() => {
                  setLoading(true);
                  setMessage(null);
                  setPage((p) => p + 1);
                }}
              >
                下一页
              </button>
            </div>
          </>
        )}

        {message ? (
          <p
            className={`${styles.message} ${message.tone === "success" ? styles.messageSuccess : styles.messageError}`}
            role={message.tone === "error" ? "alert" : "status"}
          >
            {message.text}
          </p>
        ) : null}
      </StudioPanel>

      {/* Create / Edit modal */}
      {showForm ? (
        <div className={styles.overlay} role="dialog" aria-label={editingUser ? "编辑用户" : "创建用户"}>
          <div className={styles.modal}>
            <div className={styles.panelHeader}>
              <h3>{editingUser ? "编辑用户" : "创建用户"}</h3>
              <button aria-label="关闭" className="secondary-button" type="button" onClick={closeForm}>
                <X aria-hidden size={18} />
              </button>
            </div>
            <form className={styles.formGrid} onSubmit={onSubmitForm}>
              <label className={styles.fieldFull}>
                <span>用户名</span>
                <input
                  autoComplete="username"
                  required
                  value={formUsername}
                  onChange={(event) => setFormUsername(event.target.value)}
                />
              </label>
              <label className={styles.fieldFull}>
                <span>密码{editingUser ? "（留空不修改）" : ""}</span>
                <input
                  autoComplete="new-password"
                  minLength={editingUser ? undefined : 8}
                  placeholder={editingUser ? "留空保持不变" : "至少 8 位"}
                  type="password"
                  value={formPassword}
                  onChange={(event) => setFormPassword(event.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span>邮箱</span>
                <input
                  autoComplete="email"
                  placeholder="user@example.com"
                  type="email"
                  value={formEmail}
                  onChange={(event) => setFormEmail(event.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span>昵称</span>
                <input
                  value={formNickname}
                  onChange={(event) => setFormNickname(event.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span>手机号</span>
                <input
                  autoComplete="tel"
                  value={formPhone}
                  onChange={(event) => setFormPhone(event.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span>角色</span>
                <select value={formRole} onChange={(event) => setFormRole(event.target.value)}>
                  <option value="member">成员</option>
                  <option value="admin">管理员</option>
                </select>
              </label>
              <label className={styles.field}>
                <span>状态</span>
                <select value={formStatus} onChange={(event) => setFormStatus(event.target.value)}>
                  <option value="active">活跃</option>
                  <option value="disabled">已禁用</option>
                  <option value="locked">已锁定</option>
                  <option value="pending">待激活</option>
                </select>
              </label>
              <div className={`${styles.modalActions} ${styles.fieldFull}`}>
                <button className="secondary-button" type="button" onClick={closeForm}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} type="submit">
                  {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                  {editingUser ? "保存修改" : "创建用户"}
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      {/* Delete confirmation */}
      {showDelete ? (
        <div className={styles.overlay} role="alertdialog" aria-label="确认删除">
          <div className={styles.modal}>
            <h3>确认删除用户</h3>
            <p>
              确定要删除用户 <strong>{showDelete.username}</strong>（ID: {showDelete.id}）吗？
              删除后其用户名和邮箱可被新用户复用。
            </p>
            <div className={styles.modalActions}>
              <button className="secondary-button" type="button" onClick={() => setShowDelete(null)}>
                取消
              </button>
              <button className="primary-button" disabled={saving} type="button" onClick={onDeleteConfirm}>
                {saving ? <Loader2 aria-hidden className="spin" size={18} /> : null}
                确认删除
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

function statusLabel(status: string) {
  switch (status) {
    case "active": return "活跃";
    case "disabled": return "已禁用";
    case "locked": return "已锁定";
    case "pending": return "待激活";
    default: return status;
  }
}

function statusPillTone(status: string) {
  switch (status) {
    case "active": return styles.statusReady;
    case "disabled": return styles.messageError;
    default: return styles.statusPending;
  }
}

function formatDate(iso: string) {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" });
}
