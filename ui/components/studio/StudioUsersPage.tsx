"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { FileUp, Loader2, Pencil, Plus, Save, Trash2, Users, X } from "lucide-react";

import { attachmentContentUrl, uploadLocalAttachment } from "@/lib/api/attachments";
import { humanizeApiError } from "@/lib/api/client";
import type { ListUsersResponse, UpdateUserRequest, UserItem } from "@/lib/api/types";
import { createUser, deleteUser, listUsers, updateUser } from "@/lib/api/users";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioSelect } from "./StudioSelect";
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
  const [selectedUserIDs, setSelectedUserIDs] = useState<number[]>([]);
  const [showBulkEdit, setShowBulkEdit] = useState(false);
  const [showBulkDelete, setShowBulkDelete] = useState(false);

  // Form state
  const [formUsername, setFormUsername] = useState("");
  const [formPassword, setFormPassword] = useState("");
  const [formEmail, setFormEmail] = useState("");
  const [formNickname, setFormNickname] = useState("");
  const [formPhone, setFormPhone] = useState("");
  const [formRole, setFormRole] = useState("member");
  const [formStatus, setFormStatus] = useState("active");
  const [formAvatarAttachmentID, setFormAvatarAttachmentID] = useState<number | null>(null);
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [bulkRole, setBulkRole] = useState("member");
  const [bulkStatus, setBulkStatus] = useState("active");

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
    setFormAvatarAttachmentID(null);
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
    setFormAvatarAttachmentID(user.avatar_attachment_id ?? null);
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
        if (formAvatarAttachmentID !== (editingUser.avatar_attachment_id ?? null)) {
          payload.avatar_attachment_id = formAvatarAttachmentID ?? 0;
        }
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

  async function onConfirmBulkDelete() {
    if (selectedUserIDs.length === 0) return;
    setSaving(true);
    setMessage(null);
    const ids = [...selectedUserIDs];
    let deleted = 0;
    try {
      for (const id of ids) {
        try {
          await deleteUser(id);
          deleted += 1;
        } catch (error) {
          const detail = humanizeApiError(error);
          if (deleted > 0) {
            setMessage({ tone: "error", text: `已删除 ${deleted} 个，第 ${deleted + 1} 个失败：${detail}` });
            setShowBulkDelete(false);
            setSelectedUserIDs(ids.slice(deleted));
            await fetchUsers();
          } else {
            setMessage({ tone: "error", text: detail });
          }
          return;
        }
      }
      setShowBulkDelete(false);
      setSelectedUserIDs([]);
      setMessage({
        tone: "success",
        text: ids.length === 1 ? `用户已删除。` : `已删除 ${ids.length} 个用户。`
      });
      await fetchUsers();
    } finally {
      setSaving(false);
    }
  }

  async function onAvatarFile(file: File | null) {
    if (!file || !editingUser) return;
    setAvatarUploading(true);
    setMessage(null);
    try {
      const formData = new FormData();
      formData.set("file", file);
      formData.set("owner_user_id", String(editingUser.id));
      formData.set("purpose", "avatar");
      formData.set("access_scope", "public");
      const uploaded = await uploadLocalAttachment(formData);
      setFormAvatarAttachmentID(uploaded.id);
      setMessage({ tone: "success", text: "头像已上传，保存用户后生效。" });
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setAvatarUploading(false);
    }
  }

  function toggleUserSelection(id: number, checked: boolean) {
    setSelectedUserIDs((current) => (checked ? Array.from(new Set([...current, id])) : current.filter((item) => item !== id)));
  }

  function toggleVisibleUsers(checked: boolean) {
    const ids = data?.items.map((user) => user.id) ?? [];
    setSelectedUserIDs((current) => {
      if (!checked) return current.filter((id) => !ids.includes(id));
      return Array.from(new Set([...current, ...ids]));
    });
  }

  function openBulkEdit() {
    const selected = data?.items.find((user) => selectedUserIDs.includes(user.id));
    if (!selected) return;
    setBulkRole(selected.role);
    setBulkStatus(selected.status);
    setShowBulkEdit(true);
    setMessage(null);
  }

  function openBulkDelete() {
    if (selectedUserIDs.length === 0) return;
    setShowBulkDelete(true);
    setMessage(null);
  }

  async function onBulkSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (selectedUserIDs.length === 0) return;
    setSaving(true);
    setMessage(null);
    try {
      await Promise.all(selectedUserIDs.map((id) => updateUser(id, { role: bulkRole, status: bulkStatus })));
      setSelectedUserIDs([]);
      setShowBulkEdit(false);
      setMessage({ tone: "success", text: "已批量更新用户。" });
      await fetchUsers();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;
  const visibleUserIDs = data?.items.map((user) => user.id) ?? [];
  const allVisibleUsersSelected = visibleUserIDs.length > 0 && visibleUserIDs.every((id) => selectedUserIDs.includes(id));

  return (
    <>
      <StudioTopbar
        description="管理平台注册用户，支持创建、编辑、筛选和删除。"
        eyebrow="User management"
        title="用户"
      />

      {/* Filter bar */}
      <div className={`${styles.filterBar} ${styles.toolbarPanel}`}>
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
        <StudioSelect
          ariaLabel="按状态筛选"
          className={styles.filterSelect}
          options={statusOptions}
          value={statusFilter}
          onChange={(value) => {
            setLoading(true);
            setMessage(null);
            setStatusFilter(value);
            setPage(1);
          }}
        />
        <StudioSelect
          ariaLabel="按角色筛选"
          className={styles.filterSelect}
          options={roleOptions}
          value={roleFilter}
          onChange={(value) => {
            setLoading(true);
            setMessage(null);
            setRoleFilter(value);
            setPage(1);
          }}
        />
        <button className={`primary-button ${styles.filterAction}`} type="button" onClick={openCreate}>
          <Plus aria-hidden size={18} />
          创建用户
        </button>
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
            {selectedUserIDs.length > 0 ? (
              <div className={styles.selectionBar}>
                <span>已选择 {selectedUserIDs.length} 个用户</span>
                <button className="primary-button" type="button" onClick={openBulkEdit}>
                  <Pencil aria-hidden size={16} />
                  编辑已选
                </button>
                <button className="danger-button" type="button" onClick={openBulkDelete}>
                  <Trash2 aria-hidden size={16} />
                  批量删除
                </button>
              </div>
            ) : null}
            <div className={styles.tableScroll}>
              <table className={`${styles.table} ${styles.userTable}`}>
                <colgroup>
                  <col className={styles.userSelectColumn} />
                  <col className={styles.userIdColumn} />
                  <col className={styles.userNameColumn} />
                  <col className={styles.userEmailColumn} />
                  <col className={styles.userRoleColumn} />
                  <col className={styles.userStatusColumn} />
                  <col className={styles.userCreatedColumn} />
                  <col className={styles.userActionsColumn} />
                </colgroup>
                <thead>
                  <tr>
                    <th>
                      <label className={styles.selectCell} aria-label="选择当前页用户">
                        <input checked={allVisibleUsersSelected} type="checkbox" onChange={(event) => toggleVisibleUsers(event.target.checked)} />
                      </label>
                    </th>
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
                      <td>
                        <label className={styles.selectCell} aria-label={`选择用户 ${user.username}`}>
                          <input
                            checked={selectedUserIDs.includes(user.id)}
                            type="checkbox"
                            onChange={(event) => toggleUserSelection(user.id, event.target.checked)}
                          />
                        </label>
                      </td>
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
            </div>

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
              {editingUser ? (
                <div className={`${styles.avatarEditor} ${styles.fieldFull}`}>
                  <div className={styles.avatarPreview}>
                    {formAvatarAttachmentID ? (
                      <Image
                        alt={`${editingUser.username} 头像`}
                        height={72}
                        src={attachmentContentUrl(formAvatarAttachmentID)}
                        unoptimized
                        width={72}
                      />
                    ) : (
                      <span>{editingUser.username.slice(0, 2).toUpperCase()}</span>
                    )}
                  </div>
                  <div className={styles.avatarControls}>
                    <strong>头像</strong>
                    <span>上传头像后点击保存修改完成绑定。</span>
                    <div className={styles.metaRow}>
                      <label className="secondary-button">
                        {avatarUploading ? <Loader2 aria-hidden className="spin" size={18} /> : <FileUp aria-hidden size={18} />}
                        上传头像
                        <input
                          accept="image/*"
                          className={styles.fileInputHidden}
                          type="file"
                          onChange={(event) => {
                            void onAvatarFile(event.target.files?.[0] ?? null);
                            event.currentTarget.value = "";
                          }}
                        />
                      </label>
                      {formAvatarAttachmentID ? (
                        <button className="secondary-button" type="button" onClick={() => setFormAvatarAttachmentID(null)}>
                          清除头像
                        </button>
                      ) : null}
                    </div>
                  </div>
                </div>
              ) : null}
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
                <StudioSelect ariaLabel="角色" options={roleOptions.slice(1)} value={formRole} onChange={setFormRole} />
              </label>
              <label className={styles.field}>
                <span>状态</span>
                <StudioSelect ariaLabel="状态" options={statusOptions.slice(1)} value={formStatus} onChange={setFormStatus} />
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

      {showBulkEdit ? (
        <div className={styles.overlay} role="dialog" aria-label="批量编辑用户">
          <div className={styles.modal}>
            <div className={styles.panelHeader}>
              <h3>批量编辑用户</h3>
              <button aria-label="关闭" className="secondary-button" type="button" onClick={() => setShowBulkEdit(false)}>
                <X aria-hidden size={18} />
              </button>
            </div>
            <form className={styles.formGrid} onSubmit={onBulkSubmit}>
              <p className={styles.fieldFull}>将对已选择的 {selectedUserIDs.length} 个用户统一更新角色和状态。</p>
              <label className={styles.field}>
                <span>角色</span>
                <StudioSelect ariaLabel="批量角色" options={roleOptions.slice(1)} value={bulkRole} onChange={setBulkRole} />
              </label>
              <label className={styles.field}>
                <span>状态</span>
                <StudioSelect ariaLabel="批量状态" options={statusOptions.slice(1)} value={bulkStatus} onChange={setBulkStatus} />
              </label>
              <div className={`${styles.modalActions} ${styles.fieldFull}`}>
                <button className="secondary-button" type="button" onClick={() => setShowBulkEdit(false)}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} type="submit">
                  {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                  保存批量修改
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      {showBulkDelete ? (
        <div className={styles.overlay} role="alertdialog" aria-label="确认批量删除">
          <div className={styles.modal}>
            <h3>批量删除用户</h3>
            <p>
              确认删除已选择的 {selectedUserIDs.length} 个用户？删除后其用户名和邮箱可被新用户复用。
            </p>
            <div className={styles.modalActions}>
              <button className="secondary-button" type="button" onClick={() => setShowBulkDelete(false)}>
                取消
              </button>
              <button className="danger-button" disabled={saving} type="button" onClick={onConfirmBulkDelete}>
                {saving ? <Loader2 aria-hidden className="spin" size={18} /> : null}
                确认删除
              </button>
            </div>
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
