"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { CheckCircle2, Download, FileUp, Loader2, Pencil, Save, Trash2, X } from "lucide-react";

import {
  attachmentContentUrl,
  completeAttachment,
  createAttachmentCategory,
  deleteAttachmentCategory,
  deleteAttachment,
  listAttachmentCategories,
  listAttachments,
  updateAttachmentCategory,
  updateAttachment,
  uploadLocalAttachment
} from "@/lib/api/attachments";
import { useAuth } from "@/components/auth/AuthProvider";
import { humanizeApiError } from "@/lib/api/client";
import type {
  AttachmentCategoryResponse,
  AttachmentListResponse,
  AttachmentResponse,
  StorageMountResponse
} from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPanel } from "./StudioPanel";
import { StudioSelect } from "./StudioSelect";

type FileMessage = { tone: "success" | "error"; text: string } | null;

const attachmentLimit = 20;
const purposeOptions = [
  { value: "", label: "全部用途" },
  { value: "content", label: "内容" },
  { value: "avatar", label: "头像" }
];
const statusOptions = [
  { value: "", label: "全部状态" },
  { value: "active", label: "活跃" },
  { value: "archived", label: "已归档" }
];
const accessOptions = [
  { value: "private", label: "私有" },
  { value: "public", label: "公开" }
];
const categoryStatusOptions = [
  { value: "active", label: "启用" },
  { value: "disabled", label: "停用" }
];

export function StudioFileManager({ mounts }: { mounts: StorageMountResponse[] }) {
  const { claims } = useAuth();
  const [data, setData] = useState<AttachmentListResponse>({ items: [] });
  const [categories, setCategories] = useState<AttachmentCategoryResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<FileMessage>(null);
  const [purpose, setPurpose] = useState("");
  const [status, setStatus] = useState("");
  const [categoryID, setCategoryID] = useState("");
  const [cursor, setCursor] = useState("");
  const [cursorStack, setCursorStack] = useState<string[]>([]);
  const [showUpload, setShowUpload] = useState(false);
  const [showCategories, setShowCategories] = useState(false);
  const [editing, setEditing] = useState<AttachmentResponse | null>(null);
  const [deleting, setDeleting] = useState<AttachmentResponse | null>(null);
  const [completing, setCompleting] = useState<AttachmentResponse | null>(null);
  const [categoryEditing, setCategoryEditing] = useState<AttachmentCategoryResponse | null>(null);
  const [categoryDeleting, setCategoryDeleting] = useState<AttachmentCategoryResponse | null>(null);

  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadPurpose, setUploadPurpose] = useState("content");
  const [uploadAccess, setUploadAccess] = useState("private");
  const [uploadMountID, setUploadMountID] = useState("");
  const [uploadCategoryID, setUploadCategoryID] = useState("");

  const [editOriginalName, setEditOriginalName] = useState("");
  const [editAccess, setEditAccess] = useState("private");
  const [editStatus, setEditStatus] = useState("active");
  const [editCategoryIDs, setEditCategoryIDs] = useState<number[]>([]);
  const [completeETag, setCompleteETag] = useState("");
  const [completeChecksum, setCompleteChecksum] = useState("");
  const [completeSize, setCompleteSize] = useState("");
  const [categoryName, setCategoryName] = useState("");
  const [categorySlug, setCategorySlug] = useState("");
  const [categoryParentID, setCategoryParentID] = useState("");
  const [categoryDescription, setCategoryDescription] = useState("");
  const [categoryIcon, setCategoryIcon] = useState("");
  const [categorySortOrder, setCategorySortOrder] = useState("0");
  const [categoryStatus, setCategoryStatus] = useState("active");

  const mountOptions = useMemo(
    () => [
      { value: "", label: "默认存储" },
      ...mounts
        .filter((mount) => !mount.disabled)
        .map((mount) => ({ value: String(mount.id), label: `${mount.name} (${mount.mount_path})` }))
    ],
    [mounts]
  );
  const categoryOptions = useMemo(
    () => [
      { value: "", label: "全部分类" },
      ...categories.map((category) => ({ value: String(category.id), label: category.name }))
    ],
    [categories]
  );
  const categoryParentOptions = useMemo(
    () => [
      { value: "", label: "根分类" },
      ...categories
        .filter((category) => category.id !== categoryEditing?.id)
        .map((category) => ({ value: String(category.id), label: `${"— ".repeat(category.depth)}${category.name}` }))
    ],
    [categories, categoryEditing]
  );

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [attachments, categoryPayload] = await Promise.all([
        listAttachments({
          category_id: categoryID ? Number(categoryID) : undefined,
          cursor: cursor || undefined,
          limit: attachmentLimit,
          purpose: purpose || undefined,
          status: status || undefined
        }),
        listAttachmentCategories()
      ]);
      setData(attachments);
      setCategories(categoryPayload.items);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setLoading(false);
    }
  }, [categoryID, cursor, purpose, status]);

  useEffect(() => {
    let active = true;
    Promise.all([
      listAttachments({
        category_id: categoryID ? Number(categoryID) : undefined,
        cursor: cursor || undefined,
        limit: attachmentLimit,
        purpose: purpose || undefined,
        status: status || undefined
      }),
      listAttachmentCategories()
    ])
      .then(([attachments, categoryPayload]) => {
        if (!active) return;
        setData(attachments);
        setCategories(categoryPayload.items);
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
  }, [categoryID, cursor, purpose, status]);

  function resetUploadForm() {
    setUploadFile(null);
    setUploadPurpose("content");
    setUploadAccess("private");
    setUploadMountID("");
    setUploadCategoryID("");
  }

  function resetCategoryForm() {
    setCategoryEditing(null);
    setCategoryName("");
    setCategorySlug("");
    setCategoryParentID("");
    setCategoryDescription("");
    setCategoryIcon("");
    setCategorySortOrder("0");
    setCategoryStatus("active");
  }

  function openUpload() {
    resetUploadForm();
    setMessage(null);
    setShowUpload(true);
  }

  function openEdit(attachment: AttachmentResponse) {
    setEditing(attachment);
    setEditOriginalName(attachment.original_name ?? "");
    setEditAccess(attachment.access_scope);
    setEditStatus(attachment.status);
    setEditCategoryIDs(attachment.category_ids ?? []);
    setMessage(null);
  }

  function openCategoryCreate() {
    resetCategoryForm();
    setMessage(null);
    setShowCategories(true);
  }

  function openCategoryEdit(category: AttachmentCategoryResponse) {
    setCategoryEditing(category);
    setCategoryName(category.name);
    setCategorySlug(category.slug);
    setCategoryParentID(category.parent_id ? String(category.parent_id) : "");
    setCategoryDescription(category.description ?? "");
    setCategoryIcon(category.icon ?? "");
    setCategorySortOrder(String(category.sort_order));
    setCategoryStatus(category.status);
    setMessage(null);
    setShowCategories(true);
  }

  function openComplete(attachment: AttachmentResponse) {
    setCompleting(attachment);
    setCompleteETag(attachment.etag ?? "");
    setCompleteChecksum(attachment.checksum ?? "");
    setCompleteSize(String(attachment.size));
    setMessage(null);
  }

  async function onUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      if (!uploadFile) {
        setMessage({ tone: "error", text: "请选择要上传的文件。" });
        return;
      }
      if (!claims?.uid) {
        setMessage({ tone: "error", text: "当前会话缺少用户 ID，请重新登录后再上传。" });
        return;
      }
      const formData = new FormData();
      formData.set("file", uploadFile);
      formData.set("owner_user_id", String(claims.uid));
      formData.set("purpose", uploadPurpose);
      formData.set("access_scope", uploadAccess);
      if (uploadMountID) formData.set("storage_mount_id", uploadMountID);
      if (uploadCategoryID) formData.set("category_ids", uploadCategoryID);
      await uploadLocalAttachment(formData);
      setShowUpload(false);
      setMessage({ tone: "success", text: "文件已上传。" });
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onSaveEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editing) return;
    setSaving(true);
    setMessage(null);
    try {
      await updateAttachment(editing.id, {
        access_scope: editAccess,
        category_ids: editCategoryIDs,
        original_name: editOriginalName.trim() || null,
        status: editStatus
      });
      setEditing(null);
      setMessage({ tone: "success", text: "附件信息已更新。" });
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onConfirmDelete() {
    if (!deleting) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteAttachment(deleting.id);
      setMessage({ tone: "success", text: `${displayName(deleting)} 已删除。` });
      setDeleting(null);
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onComplete(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!completing) return;
    setSaving(true);
    setMessage(null);
    try {
      await completeAttachment(completing.id, {
        checksum: completeChecksum.trim() || undefined,
        etag: completeETag.trim() || undefined,
        size: completeSize.trim() ? Number.parseInt(completeSize, 10) : undefined
      });
      setCompleting(null);
      setMessage({ tone: "success", text: "远端附件已标记完成。" });
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onSaveCategory(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const sortOrder = Number.parseInt(categorySortOrder, 10);
    if (!categoryName.trim() || !categorySlug.trim()) {
      setMessage({ tone: "error", text: "分类名称和 slug 必填。" });
      return;
    }
    if (!Number.isFinite(sortOrder)) {
      setMessage({ tone: "error", text: "排序必须是有效数字。" });
      return;
    }
    setSaving(true);
    setMessage(null);
    try {
      const payload = {
        description: categoryDescription.trim() || null,
        icon: categoryIcon.trim() || null,
        name: categoryName.trim(),
        parent_id: categoryParentID ? Number(categoryParentID) : null,
        slug: categorySlug.trim(),
        sort_order: sortOrder,
        status: categoryStatus
      };
      if (categoryEditing) {
        await updateAttachmentCategory(categoryEditing.id, payload);
        setMessage({ tone: "success", text: "分类已更新。" });
      } else {
        await createAttachmentCategory(payload);
        setMessage({ tone: "success", text: "分类已创建。" });
      }
      resetCategoryForm();
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onConfirmDeleteCategory() {
    if (!categoryDeleting) return;
    setSaving(true);
    setMessage(null);
    try {
      await deleteAttachmentCategory(categoryDeleting.id);
      setCategoryDeleting(null);
      setMessage({ tone: "success", text: `${categoryDeleting.name} 已删除。` });
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  function resetFilters() {
    setLoading(true);
    setPurpose("");
    setStatus("");
    setCategoryID("");
    setCursor("");
    setCursorStack([]);
    setMessage(null);
  }

  function nextPage() {
    if (!data.next_cursor) return;
    setLoading(true);
    setCursorStack((current) => [...current, cursor]);
    setCursor(data.next_cursor);
  }

  function previousPage() {
    setLoading(true);
    setCursorStack((current) => {
      const next = [...current];
      const previous = next.pop() ?? "";
      setCursor(previous);
      return next;
    });
  }

  return (
    <>
      {message ? (
        <p className={`${styles.message} ${message.tone === "success" ? styles.messageSuccess : styles.messageError}`} role="alert">
          {message.text}
        </p>
      ) : null}

      <StudioPanel
        action={
          <div className={styles.panelActions}>
            <button className="secondary-button" type="button" onClick={openCategoryCreate}>
              <Pencil aria-hidden size={18} />
              管理分类
            </button>
            <button className="primary-button" type="button" onClick={openUpload}>
              <FileUp aria-hidden size={18} />
              上传文件
            </button>
          </div>
        }
        title="文件"
      >
        <div className={styles.filterBar}>
          <StudioSelect
            ariaLabel="按用途筛选"
            className={styles.filterSelect}
            options={purposeOptions}
            value={purpose}
            onChange={(value) => {
              setLoading(true);
              setPurpose(value);
              setCursor("");
              setCursorStack([]);
            }}
          />
          <StudioSelect
            ariaLabel="按状态筛选"
            className={styles.filterSelect}
            options={statusOptions}
            value={status}
            onChange={(value) => {
              setLoading(true);
              setStatus(value);
              setCursor("");
              setCursorStack([]);
            }}
          />
          <StudioSelect
            ariaLabel="按分类筛选"
            className={styles.filterSelect}
            options={categoryOptions}
            value={categoryID}
            onChange={(value) => {
              setLoading(true);
              setCategoryID(value);
              setCursor("");
              setCursorStack([]);
            }}
          />
          <button className="secondary-button" type="button" onClick={resetFilters}>
            重置
          </button>
        </div>

        {loading ? (
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            正在加载文件...
          </div>
        ) : data.items.length === 0 ? (
          <div className={styles.emptyState}>暂无文件。上传内容附件后会显示在这里。</div>
        ) : (
          <div className={styles.tableScroll}>
            <table className={`${styles.table} ${styles.fileTable}`}>
              <thead>
                <tr>
                  <th>文件</th>
                  <th>用途</th>
                  <th>范围</th>
                  <th>上传</th>
                  <th>存储</th>
                  <th>大小</th>
                  <th>更新</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((attachment) => (
                  <tr key={attachment.id}>
                    <td>
                      <div className={styles.fileNameCell}>
                        <strong>{displayName(attachment)}</strong>
                        <span className={styles.codePill}>{attachment.object_key}</span>
                      </div>
                    </td>
                    <td>{labelForPurpose(attachment.purpose)}</td>
                    <td>
                      <span className={styles.statusPill}>{attachment.access_scope === "public" ? "公开" : "私有"}</span>
                    </td>
                    <td>
                      <StatusBadge value={attachment.upload_status} />
                    </td>
                    <td>{mountLabel(attachment.storage_mount_id, mounts)}</td>
                    <td>{formatBytes(attachment.size)}</td>
                    <td>{formatDate(attachment.updated_at)}</td>
                    <td>
                      <div className={styles.tableActions}>
                        <a className="icon-button" href={attachmentContentUrl(attachment.id)} target="_blank" rel="noreferrer" aria-label="下载文件">
                          <Download aria-hidden size={16} />
                        </a>
                        {attachment.upload_status === "pending" ? (
                          <button className="icon-button" type="button" aria-label="标记完成" onClick={() => openComplete(attachment)}>
                            <CheckCircle2 aria-hidden size={16} />
                          </button>
                        ) : null}
                        <button className="icon-button" type="button" aria-label="编辑文件" onClick={() => openEdit(attachment)}>
                          <Pencil aria-hidden size={16} />
                        </button>
                        <button className="icon-button" type="button" aria-label="删除文件" onClick={() => setDeleting(attachment)}>
                          <Trash2 aria-hidden size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className={styles.pagination}>
          <button className="secondary-button" disabled={cursorStack.length === 0 || loading} type="button" onClick={previousPage}>
            上一页
          </button>
          <button className="secondary-button" disabled={!data.next_cursor || loading} type="button" onClick={nextPage}>
            下一页
          </button>
        </div>
      </StudioPanel>

      {showUpload &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modalWide} role="dialog">
              <div className={styles.modalHeader}>
                <div>
                  <h3>上传文件</h3>
                  <p>未选择存储实例时，后端会使用默认启用存储。</p>
                </div>
                <button aria-label="关闭" className="icon-button" type="button" onClick={() => setShowUpload(false)}>
                  <X aria-hidden size={18} />
                </button>
              </div>
              <form className={styles.formGrid} id="attachment-upload-form" onSubmit={onUpload}>
                <label className={styles.field}>
                  <span>用途</span>
                  <StudioSelect ariaLabel="用途" options={purposeOptions.filter((option) => option.value)} value={uploadPurpose} onChange={setUploadPurpose} />
                </label>
                <label className={styles.field}>
                  <span>访问范围</span>
                  <StudioSelect ariaLabel="访问范围" options={accessOptions} value={uploadAccess} onChange={setUploadAccess} />
                </label>
                <label className={styles.field}>
                  <span>存储实例</span>
                  <StudioSelect ariaLabel="存储实例" options={mountOptions} value={uploadMountID} onChange={setUploadMountID} />
                </label>
                <label className={styles.field}>
                  <span>分类</span>
                  <StudioSelect ariaLabel="分类" options={categoryOptions} value={uploadCategoryID} onChange={setUploadCategoryID} />
                </label>

                <label className={styles.fieldFull}>
                  <span>文件</span>
                  <input aria-label="文件" type="file" onChange={(event) => setUploadFile(event.target.files?.[0] ?? null)} />
                </label>
              </form>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setShowUpload(false)}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} form="attachment-upload-form" type="submit">
                  {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                  上传
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}

      {showCategories &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modalWide} role="dialog">
              <div className={styles.modalHeader}>
                <div>
                  <h3>管理文件分类</h3>
                  <p>分类用于筛选和绑定附件；停用分类不会删除已有关联。</p>
                </div>
                <button
                  aria-label="关闭"
                  className="icon-button"
                  type="button"
                  onClick={() => {
                    setShowCategories(false);
                    resetCategoryForm();
                  }}
                >
                  <X aria-hidden size={18} />
                </button>
              </div>

              <div className={styles.categoryManager}>
                <form className={styles.formGrid} id="attachment-category-form" onSubmit={onSaveCategory}>
                  <label className={styles.field}>
                    <span>名称</span>
                    <input aria-label="分类名称" value={categoryName} onChange={(event) => setCategoryName(event.target.value)} />
                  </label>
                  <label className={styles.field}>
                    <span>Slug</span>
                    <input aria-label="分类 slug" value={categorySlug} onChange={(event) => setCategorySlug(event.target.value)} />
                  </label>
                  <label className={styles.field}>
                    <span>父分类</span>
                    <StudioSelect ariaLabel="父分类" options={categoryParentOptions} value={categoryParentID} onChange={setCategoryParentID} />
                  </label>
                  <label className={styles.field}>
                    <span>状态</span>
                    <StudioSelect ariaLabel="分类状态" options={categoryStatusOptions} value={categoryStatus} onChange={setCategoryStatus} />
                  </label>
                  <label className={styles.field}>
                    <span>图标</span>
                    <input aria-label="分类图标" value={categoryIcon} onChange={(event) => setCategoryIcon(event.target.value)} />
                  </label>
                  <label className={styles.field}>
                    <span>排序</span>
                    <input aria-label="分类排序" type="number" value={categorySortOrder} onChange={(event) => setCategorySortOrder(event.target.value)} />
                  </label>
                  <label className={styles.fieldFull}>
                    <span>描述</span>
                    <textarea
                      aria-label="分类描述"
                      className={styles.textarea}
                      value={categoryDescription}
                      onChange={(event) => setCategoryDescription(event.target.value)}
                    />
                  </label>
                </form>

                <div className={styles.categoryActions}>
                  <button className="secondary-button" type="button" onClick={resetCategoryForm}>
                    新建分类
                  </button>
                  <button className="primary-button" disabled={saving} form="attachment-category-form" type="submit">
                    {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                    {categoryEditing ? "保存分类" : "创建分类"}
                  </button>
                </div>

                <div className={styles.tableScroll}>
                  <table className={styles.table}>
                    <thead>
                      <tr>
                        <th>分类</th>
                        <th>Slug</th>
                        <th>状态</th>
                        <th>排序</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {categories.length === 0 ? (
                        <tr>
                          <td colSpan={5}>暂无分类。</td>
                        </tr>
                      ) : (
                        categories.map((category) => (
                          <tr key={category.id}>
                            <td>
                              <div className={styles.fileNameCell}>
                                <strong>{`${"— ".repeat(category.depth)}${category.name}`}</strong>
                                <span className={styles.codePill}>{category.path}</span>
                              </div>
                            </td>
                            <td>{category.slug}</td>
                            <td>
                              <StatusBadge value={category.status} />
                            </td>
                            <td>{category.sort_order}</td>
                            <td>
                              <div className={styles.tableActions}>
                                <button className="icon-button" type="button" aria-label={`编辑分类 ${category.name}`} onClick={() => openCategoryEdit(category)}>
                                  <Pencil aria-hidden size={16} />
                                </button>
                                <button className="icon-button" type="button" aria-label={`删除分类 ${category.name}`} onClick={() => setCategoryDeleting(category)}>
                                  <Trash2 aria-hidden size={16} />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>,
          document.body
        )}

      {editing &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modalWide} role="dialog">
              <div className={styles.modalHeader}>
                <div>
                  <h3>编辑文件</h3>
                  <p>{displayName(editing)}</p>
                </div>
                <button aria-label="关闭" className="icon-button" type="button" onClick={() => setEditing(null)}>
                  <X aria-hidden size={18} />
                </button>
              </div>
              <form className={styles.formGrid} id="attachment-edit-form" onSubmit={onSaveEdit}>
                <label className={styles.fieldFull}>
                  <span>展示名称</span>
                  <input aria-label="展示名称" value={editOriginalName} onChange={(event) => setEditOriginalName(event.target.value)} />
                </label>
                <label className={styles.field}>
                  <span>访问范围</span>
                  <StudioSelect ariaLabel="编辑访问范围" options={accessOptions} value={editAccess} onChange={setEditAccess} />
                </label>
                <label className={styles.field}>
                  <span>状态</span>
                  <StudioSelect ariaLabel="编辑状态" options={statusOptions.filter((option) => option.value)} value={editStatus} onChange={setEditStatus} />
                </label>
                <div className={styles.fieldFull}>
                  <span className={styles.subsectionTitle}>分类</span>
                  <div className={styles.checkboxGrid}>
                    {categories.length === 0 ? (
                      <span className={styles.muted}>暂无分类。</span>
                    ) : (
                      categories.map((category) => (
                        <label className={styles.inlineCheck} key={category.id}>
                          <input
                            checked={editCategoryIDs.includes(category.id)}
                            type="checkbox"
                            onChange={(event) => {
                              setEditCategoryIDs((current) =>
                                event.target.checked ? [...current, category.id] : current.filter((id) => id !== category.id)
                              );
                            }}
                          />
                          {category.name}
                        </label>
                      ))
                    )}
                  </div>
                </div>
                <div className={styles.fieldFull}>
                  <span className={styles.subsectionTitle}>对象信息</span>
                  <pre className={styles.codeBlock}>{JSON.stringify({ id: editing.id, object_key: editing.object_key, mime_type: editing.mime_type }, null, 2)}</pre>
                </div>
              </form>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setEditing(null)}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} form="attachment-edit-form" type="submit">
                  {saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
                  保存
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}

      {completing &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modal} role="dialog">
              <h3>标记远端附件完成</h3>
              <form className={styles.formGrid} id="attachment-complete-form" onSubmit={onComplete}>
                <label className={styles.fieldFull}>
                  <span>ETag</span>
                  <input aria-label="ETag" value={completeETag} onChange={(event) => setCompleteETag(event.target.value)} />
                </label>
                <label className={styles.fieldFull}>
                  <span>Checksum</span>
                  <input aria-label="Checksum" value={completeChecksum} onChange={(event) => setCompleteChecksum(event.target.value)} />
                </label>
                <label className={styles.fieldFull}>
                  <span>大小（字节）</span>
                  <input aria-label="完成大小" type="number" value={completeSize} onChange={(event) => setCompleteSize(event.target.value)} />
                </label>
              </form>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setCompleting(null)}>
                  取消
                </button>
                <button className="primary-button" disabled={saving} form="attachment-complete-form" type="submit">
                  完成
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}

      {deleting &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modal} role="dialog">
              <h3>删除文件</h3>
              <p>确认删除 {displayName(deleting)}？后端会软删附件，并尽力清理真实对象。</p>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setDeleting(null)}>
                  取消
                </button>
                <button className="danger-button" disabled={saving} type="button" onClick={onConfirmDelete}>
                  删除
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}

      {categoryDeleting &&
        createPortal(
          <div className={styles.overlay} role="presentation">
            <div aria-modal="true" className={styles.modal} role="dialog">
              <h3>删除分类</h3>
              <p>确认删除 {categoryDeleting.name}？后端会软删分类，已绑定文件不会被删除。</p>
              <div className={styles.modalActions}>
                <button className="secondary-button" type="button" onClick={() => setCategoryDeleting(null)}>
                  取消
                </button>
                <button className="danger-button" disabled={saving} type="button" onClick={onConfirmDeleteCategory}>
                  删除
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}
    </>
  );
}

function StatusBadge({ value }: { value: string }) {
  if (value === "ready") return <span className={styles.statusReady}>ready</span>;
  if (value === "pending") return <span className={styles.statusPending}>pending</span>;
  return <span className={styles.statusPill}>{value}</span>;
}

function displayName(attachment: AttachmentResponse) {
  return attachment.original_name || attachment.filename;
}

function labelForPurpose(value: string) {
  if (value === "content") return "内容";
  if (value === "avatar") return "头像";
  return value;
}

function mountLabel(id: number, mounts: StorageMountResponse[]) {
  const mount = mounts.find((item) => item.id === id);
  return mount ? mount.name : `#${id}`;
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`;
  return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit"
  }).format(new Date(value));
}
