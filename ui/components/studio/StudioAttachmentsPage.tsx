"use client";

import { ChangeEvent, DragEvent, FormEvent, KeyboardEvent, MouseEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import Image from "next/image";
import {
  CheckCircle2,
  Download,
  FileImage,
  FileUp,
  Grid3X3,
  List,
  Loader2,
  MoreHorizontal,
  Pencil,
  Plus,
  RefreshCw,
  Save,
  Search,
  Trash2,
  X
} from "lucide-react";

import { useAuth } from "@/components/auth/AuthProvider";
import {
  attachmentContentUrl,
  completeAttachment,
  createAttachmentCategory,
  deleteAttachment,
  deleteAttachmentCategory,
  getAttachmentReferences,
  listAttachmentCategories,
  listAttachmentReferences,
  listAttachments,
  updateAttachment,
  updateAttachmentCategory,
  uploadLocalAttachmentsBatch
} from "@/lib/api/attachments";
import { humanizeApiError } from "@/lib/api/client";
import { listStorageMounts } from "@/lib/api/storage";
import type {
  AttachmentCategoryListResponse,
  AttachmentCategoryResponse,
  AttachmentListResponse,
  AttachmentReferenceListResponse,
  AttachmentReferenceResponse,
  AttachmentResponse,
  StorageMountListResponse,
  StorageMountResponse
} from "@/lib/api/types";
import styles from "./Studio.module.css";
import { StudioPagePagination } from "./StudioPagePagination";
import { StudioSelect } from "./StudioSelect";
import { ToastMessage } from "@/components/toast/ToastProvider";
import { StudioTopbar } from "./StudioTopbar";

type Message = { tone: "success" | "error"; text: string } | null;
type ViewMode = "list" | "grid";

const attachmentPageSize = 10;
const attachmentPagesPerFetch = 10;
const attachmentFetchSize = attachmentPageSize * attachmentPagesPerFetch;

function batchStartPageFor(pageNumber: number) {
  return Math.floor((pageNumber - 1) / attachmentPagesPerFetch) * attachmentPagesPerFetch + 1;
}

function stopRowClick(event: MouseEvent | KeyboardEvent) {
  event.stopPropagation();
}

function isImageMime(mimeType: string) {
  return mimeType.startsWith("image/");
}

// listPageForBatchStart maps UI batch start to API page (offset uses page_size as fetch window).
// listPageForBatchStart 将 UI 批次起始页映射为 API 的 page（偏移按 page_size 窗口计算）。
function listPageForBatchStart(batchStartPage: number) {
  const offset = (batchStartPage - 1) * attachmentPageSize;
  return Math.floor(offset / attachmentFetchSize) + 1;
}
type AttachmentsDataPayload = {
  attachments: AttachmentListResponse;
  categoryPayload: AttachmentCategoryListResponse;
  mountPayload: StorageMountListResponse;
  referencePayload: AttachmentReferenceListResponse;
};

// Dedupe identical in-flight list requests (e.g. React Strict Mode remount in dev).
// 对相同查询参数的进行中列表请求去重（例如开发环境 React Strict Mode 重复挂载）。
let attachmentsDataInflight: { key: string; promise: Promise<AttachmentsDataPayload> } | null = null;

function attachmentsDataKey(batchStartPage: number, listFilters: Record<string, unknown>) {
  return `${batchStartPage}\x1e${JSON.stringify(listFilters)}`;
}

function requestAttachmentsData(batchStartPage: number, listFilters: Record<string, unknown>) {
  return Promise.all([
    listAttachments({
      ...listFilters,
      page: listPageForBatchStart(batchStartPage),
      page_size: attachmentFetchSize
    }),
    listAttachmentCategories(),
    listStorageMounts(),
    listAttachmentReferences()
  ]).then(([attachments, categoryPayload, mountPayload, referencePayload]) => ({
    attachments,
    categoryPayload,
    mountPayload,
    referencePayload
  }));
}

function loadAttachmentsData(batchStartPage: number, listFilters: Record<string, unknown>) {
  const key = attachmentsDataKey(batchStartPage, listFilters);
  if (attachmentsDataInflight?.key === key) {
    return attachmentsDataInflight.promise;
  }
  const promise = requestAttachmentsData(batchStartPage, listFilters).finally(() => {
    if (attachmentsDataInflight?.promise === promise) {
      attachmentsDataInflight = null;
    }
  });
  attachmentsDataInflight = { key, promise };
  return promise;
}

const purposeOptions = [
  { value: "", label: "类型：全部" },
  { value: "content", label: "内容" },
  { value: "avatar", label: "头像" },
  { value: "system", label: "系统" },
  { value: "other", label: "其他" }
];
const referenceOptions = [
  { value: "", label: "引用：全部" },
  { value: "referenced", label: "有引用" },
  { value: "orphan", label: "孤儿附件" }
];
const statusOptions = [
  { value: "", label: "状态：全部" },
  { value: "active", label: "活跃" },
  { value: "archived", label: "已归档" }
];
const accessOptions = [
  { value: "private", label: "私有" },
  { value: "public", label: "公开" }
];
const sortOptions = [
  { value: "default", label: "排序：默认" },
  { value: "name", label: "名称" },
  { value: "size", label: "大小" }
];
const categoryStatusOptions = [
  { value: "active", label: "启用" },
  { value: "disabled", label: "停用" }
];

export function StudioAttachmentsPage() {
  const { claims } = useAuth();
  const [categories, setCategories] = useState<AttachmentCategoryResponse[]>([]);
  const [mounts, setMounts] = useState<StorageMountResponse[]>([]);
  const [references, setReferences] = useState<AttachmentReferenceResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("list");
  const [search, setSearch] = useState("");
  const [purpose, setPurpose] = useState("");
  const [status, setStatus] = useState("");
  const [referenceStatus, setReferenceStatus] = useState("");
  const [categoryID, setCategoryID] = useState("");
  const [sort, setSort] = useState("default");
  const [page, setPage] = useState(1);
  const [batchStartPage, setBatchStartPage] = useState(1);
  const [batchItems, setBatchItems] = useState<AttachmentResponse[]>([]);
  const [total, setTotal] = useState(0);
  const [showUpload, setShowUpload] = useState(false);
  const [showCategories, setShowCategories] = useState(false);
  const [editing, setEditing] = useState<AttachmentResponse | null>(null);
  const [deleting, setDeleting] = useState<AttachmentResponse | null>(null);
  const [completing, setCompleting] = useState<AttachmentResponse | null>(null);
  const [referenceTarget, setReferenceTarget] = useState<AttachmentResponse | null>(null);
  const [referenceDetail, setReferenceDetail] = useState<AttachmentReferenceResponse[]>([]);
  const [categoryEditing, setCategoryEditing] = useState<AttachmentCategoryResponse | null>(null);
  const [categoryDeleting, setCategoryDeleting] = useState<AttachmentCategoryResponse | null>(null);
  const [selectedAttachmentIDs, setSelectedAttachmentIDs] = useState<number[]>([]);
  const [bulkEditing, setBulkEditing] = useState(false);
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const [uploadFiles, setUploadFiles] = useState<File[]>([]);
  const [uploadPurpose, setUploadPurpose] = useState("content");
  const [uploadAccess, setUploadAccess] = useState("public");
  const [uploadMountID, setUploadMountID] = useState("");
  const [uploadCategoryID, setUploadCategoryID] = useState("");
  const [editOriginalName, setEditOriginalName] = useState("");
  const [editAccess, setEditAccess] = useState("private");
  const [editStatus, setEditStatus] = useState("active");
  const [editCategoryIDs, setEditCategoryIDs] = useState<number[]>([]);
  const [bulkAccess, setBulkAccess] = useState("private");
  const [bulkStatus, setBulkStatus] = useState("active");
  const [bulkCategoryIDs, setBulkCategoryIDs] = useState<number[]>([]);
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

  const categoryOptions = useMemo(
    () => [
      { value: "", label: "未分组" },
      ...categories.map((category) => ({ value: String(category.id), label: category.name }))
    ],
    [categories]
  );
  const mountOptions = useMemo(
    () => [
      { value: "", label: "默认存储" },
      ...mounts
        .filter((mount) => !mount.disabled)
        .map((mount) => ({ value: String(mount.id), label: `${mount.name} (${mount.mount_path})` }))
    ],
    [mounts]
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
  const referencesByAttachment = useMemo(() => {
    const map = new Map<number, AttachmentReferenceResponse[]>();
    for (const item of references) {
      const current = map.get(item.attachment_id) ?? [];
      current.push(item);
      map.set(item.attachment_id, current);
    }
    return map;
  }, [references]);
  const totalPages = useMemo(() => Math.max(1, Math.ceil(total / attachmentPageSize)), [total]);
  const pageItems = useMemo(() => {
    const start = (page - batchStartPage) * attachmentPageSize;
    return batchItems.slice(start, start + attachmentPageSize);
  }, [batchItems, batchStartPage, page]);
  const visibleItems = useMemo(() => {
    const items = [...pageItems];
    if (sort === "name") items.sort((a, b) => displayName(a).localeCompare(displayName(b), "zh-CN"));
    if (sort === "size") items.sort((a, b) => b.size - a.size);
    return items;
  }, [pageItems, sort]);
  const selectedAttachments = useMemo(
    () => batchItems.filter((attachment) => selectedAttachmentIDs.includes(attachment.id)),
    [batchItems, selectedAttachmentIDs]
  );
  const selectedReferencedCount = useMemo(
    () => selectedAttachmentIDs.filter((id) => hasAttachmentReferences(id, referencesByAttachment)).length,
    [referencesByAttachment, selectedAttachmentIDs]
  );
  const visibleAttachmentIDs = useMemo(() => visibleItems.map((attachment) => attachment.id), [visibleItems]);
  const allVisibleSelected = visibleAttachmentIDs.length > 0 && visibleAttachmentIDs.every((id) => selectedAttachmentIDs.includes(id));
  const unassignedCount = batchItems.filter((attachment) => (attachment.category_ids ?? []).length === 0).length;

  const listFilters = useMemo(
    () => ({
      category_id: categoryID && categoryID !== "unassigned" ? Number(categoryID) : undefined,
      category_mode: categoryID === "unassigned" ? ("unassigned" as const) : undefined,
      purpose: purpose || undefined,
      reference_status: referenceStatus || undefined,
      search: search.trim() || undefined,
      status: status || undefined
    }),
    [categoryID, purpose, referenceStatus, search, status]
  );

  const requestData = useCallback(async () => {
    const [attachments, categoryPayload, mountPayload, referencePayload] = await Promise.all([
      listAttachments({
        ...listFilters,
        page: listPageForBatchStart(batchStartPage),
        page_size: attachmentFetchSize
      }),
      listAttachmentCategories(),
      listStorageMounts(),
      listAttachmentReferences()
    ]);
    return { attachments, categoryPayload, mountPayload, referencePayload };
  }, [batchStartPage, listFilters]);

  // loadData always fetches fresh data; used after mutations.
  // loadData 始终拉取最新数据；用于创建/更新/删除后刷新。
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const { attachments, categoryPayload, mountPayload, referencePayload } = await requestData();
      setBatchItems(attachments.items);
      setTotal(attachments.total ?? attachments.items.length);
      setCategories(categoryPayload.items);
      setMounts(mountPayload.items);
      setReferences(referencePayload.items);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setLoading(false);
    }
  }, [requestData]);

  useEffect(() => {
    let active = true;
    loadAttachmentsData(batchStartPage, listFilters)
      .then(({ attachments, categoryPayload, mountPayload, referencePayload }) => {
        if (!active) return;
        setBatchItems(attachments.items);
        setTotal(attachments.total ?? attachments.items.length);
        setCategories(categoryPayload.items);
        setMounts(mountPayload.items);
        setReferences(referencePayload.items);
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
  }, [batchStartPage, listFilters]);

  function resetPage() {
    setPage(1);
    setBatchStartPage(1);
  }

  function goToPage(nextPage: number) {
    if (nextPage < 1 || nextPage > totalPages) return;
    const nextBatchStart = batchStartPageFor(nextPage);
    setPage(nextPage);
    if (nextBatchStart !== batchStartPage) {
      setBatchStartPage(nextBatchStart);
    }
  }

  function resetUploadForm() {
    setUploadFiles([]);
    setUploadPurpose("content");
    setUploadAccess("public");
    setUploadMountID("");
    setUploadCategoryID(categoryID && categoryID !== "unassigned" ? categoryID : "");
  }

  function addUploadFiles(incoming: File[]) {
    const valid = incoming.filter((file) => file.name);
    if (valid.length === 0) return;
    setUploadFiles((current) => {
      const seen = new Set(current.map(uploadFileKey));
      const next = [...current];
      for (const file of valid) {
        const key = uploadFileKey(file);
        if (seen.has(key)) continue;
        seen.add(key);
        next.push(file);
      }
      return next;
    });
  }

  function removeUploadFile(index: number) {
    setUploadFiles((current) => current.filter((_, itemIndex) => itemIndex !== index));
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

  function openEdit(attachment: AttachmentResponse) {
    setEditing(attachment);
    setEditOriginalName(attachment.original_name ?? "");
    setEditAccess(attachment.access_scope);
    setEditStatus(attachment.status);
    setEditCategoryIDs(attachment.category_ids ?? []);
    setMessage(null);
  }

  function toggleAttachmentSelection(id: number, checked: boolean) {
    setSelectedAttachmentIDs((current) => (checked ? Array.from(new Set([...current, id])) : current.filter((item) => item !== id)));
  }

  function toggleVisibleSelection(checked: boolean) {
    setSelectedAttachmentIDs((current) => {
      if (!checked) {
        return current.filter((id) => !visibleAttachmentIDs.includes(id));
      }
      return Array.from(new Set([...current, ...visibleAttachmentIDs]));
    });
  }

  function openBulkEdit() {
    if (selectedAttachments.length === 0) return;
    const first = selectedAttachments[0];
    setBulkAccess(first.access_scope);
    setBulkStatus(first.status);
    setBulkCategoryIDs(first.category_ids ?? []);
    setBulkEditing(true);
    setMessage(null);
  }

  function openBulkDelete() {
    if (selectedAttachmentIDs.length === 0) return;
    setBulkDeleting(true);
    setMessage(null);
  }

  function openComplete(attachment: AttachmentResponse) {
    setCompleting(attachment);
    setCompleteETag(attachment.etag ?? "");
    setCompleteChecksum(attachment.checksum ?? "");
    setCompleteSize(String(attachment.size));
    setMessage(null);
  }

  async function openReferences(attachment: AttachmentResponse) {
    setReferenceTarget(attachment);
    setReferenceDetail([]);
    setMessage(null);
    try {
      const result = await getAttachmentReferences(attachment.id);
      setReferenceDetail(result.items);
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    }
  }

  async function onUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      if (uploadFiles.length === 0) {
        setMessage({ tone: "error", text: "请至少添加一个文件。" });
        return;
      }
      if (!claims?.uid) {
        setMessage({ tone: "error", text: "当前会话缺少用户 ID，请重新登录后再上传。" });
        return;
      }
      const formData = new FormData();
      for (const file of uploadFiles) {
        formData.append("files", file);
      }
      formData.set("owner_user_id", String(claims.uid));
      formData.set("purpose", uploadPurpose);
      formData.set("access_scope", uploadAccess);
      if (uploadMountID) formData.set("storage_mount_id", uploadMountID);
      if (uploadCategoryID) formData.set("category_ids", uploadCategoryID);

      const result = await uploadLocalAttachmentsBatch(formData);

      if (result.failed > 0) {
        const errors = result.items
          .filter((item) => item.error)
          .map((item) => `${item.filename}: ${item.error}`)
          .join("; ");
        if (result.uploaded > 0) {
          setMessage({ tone: "error", text: `已上传 ${result.uploaded} 个，${result.failed} 个失败：${errors}` });
        } else {
          setMessage({ tone: "error", text: errors });
        }
        if (result.uploaded > 0) await loadData();
        return;
      }

      setShowUpload(false);
      setMessage({
        tone: "success",
        text: uploadFiles.length === 1 ? "附件已上传。" : `已上传 ${uploadFiles.length} 个附件。`
      });
      await loadData();
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

  async function onSaveBulkEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (selectedAttachmentIDs.length === 0) return;
    setSaving(true);
    setMessage(null);
    try {
      await Promise.all(
        selectedAttachmentIDs.map((id) =>
          updateAttachment(id, {
            access_scope: bulkAccess,
            category_ids: bulkCategoryIDs,
            status: bulkStatus
          })
        )
      );
      setBulkEditing(false);
      setSelectedAttachmentIDs([]);
      setMessage({ tone: "success", text: "已批量更新附件。" });
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onConfirmDelete() {
    if (!deleting) return;
    const force = hasAttachmentReferences(deleting.id, referencesByAttachment);
    setSaving(true);
    setMessage(null);
    try {
      await deleteAttachment(deleting.id, { force });
      setMessage({ tone: "success", text: `${displayName(deleting)} 已删除。` });
      setDeleting(null);
      await loadData();
    } catch (error) {
      setMessage({ tone: "error", text: humanizeApiError(error) });
    } finally {
      setSaving(false);
    }
  }

  async function onConfirmBulkDelete() {
    if (selectedAttachmentIDs.length === 0) return;
    setSaving(true);
    setMessage(null);
    const ids = [...selectedAttachmentIDs];
    let deleted = 0;
    try {
      for (const id of ids) {
        try {
          await deleteAttachment(id, { force: hasAttachmentReferences(id, referencesByAttachment) });
          deleted += 1;
        } catch (error) {
          const detail = humanizeApiError(error);
          if (deleted > 0) {
            setMessage({ tone: "error", text: `已删除 ${deleted} 个，第 ${deleted + 1} 个失败：${detail}` });
            setBulkDeleting(false);
            setSelectedAttachmentIDs(ids.slice(deleted));
            await loadData();
          } else {
            setMessage({ tone: "error", text: detail });
          }
          return;
        }
      }
      setBulkDeleting(false);
      setSelectedAttachmentIDs([]);
      setMessage({
        tone: "success",
        text: ids.length === 1 ? "附件已删除。" : `已删除 ${ids.length} 个附件。`
      });
      await loadData();
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
        setMessage({ tone: "success", text: "分组已更新。" });
      } else {
        await createAttachmentCategory(payload);
        setMessage({ tone: "success", text: "分组已创建。" });
      }
      setShowCategories(false);
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
    setSearch("");
    setPurpose("");
    setStatus("");
    setReferenceStatus("");
    setCategoryID("");
    resetPage();
    setMessage(null);
  }

  return (
    <>
      <StudioTopbar
        actions={
          <>
            <button className="secondary-button" type="button" onClick={() => setViewMode((value) => (value === "list" ? "grid" : "list"))}>
              {viewMode === "list" ? <Grid3X3 aria-hidden size={18} /> : <List aria-hidden size={18} />}
              缩略图
            </button>
            <button className="primary-button" type="button" onClick={openUpload}>
              <FileUp aria-hidden size={18} />
              上传
            </button>
          </>
        }
        description="管理附件、分组、引用关系和上传入口；存储实例与驱动配置保留在存储管理页。"
        eyebrow="Attachment library"
        title="附件库"
      />

      <ToastMessage message={message} />

      <section className={styles.attachmentShell}>
        <div className={styles.attachmentToolbar}>
          <label className={styles.searchInput}>
            <Search aria-hidden size={18} />
            <input
              aria-label="搜索附件"
              placeholder="输入关键词搜索"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                resetPage();
              }}
            />
          </label>
          <StudioSelect ariaLabel="按类型筛选" className={styles.filterSelect} options={purposeOptions} value={purpose} onChange={(value) => {
            setPurpose(value);
            resetPage();
          }} />
          <StudioSelect ariaLabel="按引用筛选" className={styles.filterSelect} options={referenceOptions} value={referenceStatus} onChange={(value) => {
            setReferenceStatus(value);
            resetPage();
          }} />
          <StudioSelect ariaLabel="按状态筛选" className={styles.filterSelect} options={statusOptions} value={status} onChange={(value) => {
            setStatus(value);
            resetPage();
          }} />
          <StudioSelect ariaLabel="排序" className={styles.filterSelect} options={sortOptions} value={sort} onChange={setSort} />
          <button className="secondary-button" type="button" onClick={resetFilters}>
            重置
          </button>
          <button className="icon-button" type="button" aria-label="刷新附件" onClick={() => void loadData()}>
            <RefreshCw aria-hidden size={18} />
          </button>
        </div>
        {selectedAttachmentIDs.length > 0 ? (
          <div className={styles.selectionBar}>
            <span>已选择 {selectedAttachmentIDs.length} 个附件</span>
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

        <div className={styles.categoryStrip}>
          <CategoryCard active={!categoryID} count={total || batchItems.length} title="全部" onClick={() => {
            setCategoryID("");
            resetPage();
          }} />
          <CategoryCard active={categoryID === "unassigned"} count={unassignedCount} title="未分组" onClick={() => {
            setCategoryID("unassigned");
            resetPage();
          }} />
          {categories.map((category) => (
            <CategoryCard
              active={categoryID === String(category.id)}
              key={category.id}
              title={category.name}
              onEdit={() => openCategoryEdit(category)}
              onClick={() => {
                setCategoryID(String(category.id));
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
          <div className={styles.emptyState} role="status">
            <Loader2 aria-hidden className="spin" size={24} />
            正在加载附件...
          </div>
        ) : visibleItems.length === 0 ? (
          viewMode === "list" ? (
            <div className={styles.attachmentList}>
              <AttachmentListHeader checked={allVisibleSelected} onChange={toggleVisibleSelection} />
              <div className={styles.emptyState}>暂无附件。上传内容附件后会显示在这里。</div>
            </div>
          ) : (
            <div className={styles.emptyState}>暂无附件。上传内容附件后会显示在这里。</div>
          )
        ) : viewMode === "grid" ? (
          <div className={styles.attachmentGrid}>
            {visibleItems.map((attachment) => (
              <article
                aria-label={`编辑附件 ${displayName(attachment)}`}
                className={`${styles.attachmentTile} ${styles.attachmentTileClickable}`}
                key={attachment.id}
                tabIndex={0}
                onClick={() => openEdit(attachment)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    openEdit(attachment);
                  }
                }}
              >
                <label
                  className={styles.tileCheck}
                  aria-label={`选择附件 ${displayName(attachment)}`}
                  onClick={stopRowClick}
                  onKeyDown={stopRowClick}
                >
                  <input
                    checked={selectedAttachmentIDs.includes(attachment.id)}
                    type="checkbox"
                    onChange={(event) => toggleAttachmentSelection(attachment.id, event.target.checked)}
                  />
                </label>
                <div className={styles.attachmentThumb}>
                  <FileImage aria-hidden size={34} />
                </div>
                <div>
                  <strong>{displayName(attachment)}</strong>
                  <span>{attachment.mime_type} · {formatBytes(attachment.size)}</span>
                </div>
                <div className={styles.tileActions} onClick={stopRowClick} onKeyDown={stopRowClick}>
                  <button className="icon-button" type="button" aria-label={`查看引用 ${displayName(attachment)}`} onClick={() => void openReferences(attachment)}>
                    <MoreHorizontal aria-hidden size={16} />
                  </button>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className={styles.attachmentList}>
            <AttachmentListHeader checked={allVisibleSelected} onChange={toggleVisibleSelection} />
            {visibleItems.map((attachment) => {
              const refs = referencesByAttachment.get(attachment.id) ?? [];
              return (
                <article
                  aria-label={`编辑附件 ${displayName(attachment)}`}
                  className={`${styles.attachmentRow} ${styles.attachmentRowClickable}`}
                  key={attachment.id}
                  tabIndex={0}
                  onClick={() => openEdit(attachment)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      openEdit(attachment);
                    }
                  }}
                >
                  <label
                    className={styles.selectCell}
                    aria-label={`选择附件 ${displayName(attachment)}`}
                    onClick={stopRowClick}
                    onKeyDown={stopRowClick}
                  >
                    <input
                      checked={selectedAttachmentIDs.includes(attachment.id)}
                      type="checkbox"
                      onChange={(event) => toggleAttachmentSelection(attachment.id, event.target.checked)}
                    />
                  </label>
                  <div className={styles.attachmentThumbSmall}>
                    <FileImage aria-hidden size={22} />
                  </div>
                  <div className={styles.attachmentMain}>
                    <strong>{displayName(attachment)}</strong>
                    <span>{attachment.mime_type} · {formatBytes(attachment.size)} · {attachment.object_key}</span>
                  </div>
                  <span className={styles.compactCell}>{mountLabel(attachment.storage_mount_id, mounts)}</span>
                  <span className={styles.compactCell}>{ownerLabel(attachment)}</span>
                  <button
                    className={refs.length > 0 ? styles.statusReady : styles.statusPending}
                    type="button"
                    onClick={(event) => {
                      stopRowClick(event);
                      void openReferences(attachment);
                    }}
                  >
                    {refs.length > 0 ? `${refs.length} 引用` : "孤儿"}
                  </button>
                  <span>{formatDate(attachment.updated_at)}</span>
                  <div className={styles.tableActions} onClick={stopRowClick} onKeyDown={stopRowClick}>
                    <a
                      className="icon-button"
                      href={attachmentContentUrl(attachment.id)}
                      target="_blank"
                      rel="noreferrer"
                      aria-label="下载附件"
                      onClick={stopRowClick}
                    >
                      <Download aria-hidden size={16} />
                    </a>
                    {attachment.upload_status === "pending" ? (
                      <button className="icon-button" type="button" aria-label="标记完成" onClick={() => openComplete(attachment)}>
                        <CheckCircle2 aria-hidden size={16} />
                      </button>
                    ) : null}
                    <button className="icon-button" type="button" aria-label="编辑附件" onClick={() => openEdit(attachment)}>
                      <Pencil aria-hidden size={16} />
                    </button>
                    <button className="icon-button" type="button" aria-label="删除附件" onClick={() => setDeleting(attachment)}>
                      <Trash2 aria-hidden size={16} />
                    </button>
                  </div>
                </article>
              );
            })}
          </div>
        )}

        {!loading ? (
          <StudioPagePagination disabled={loading} page={page} totalPages={totalPages} onPageChange={goToPage} />
        ) : null}
      </section>

      {showUpload ? (
        <AttachmentUploadModal
          access={uploadAccess}
          categoryID={uploadCategoryID}
          categoryOptions={categoryOptions}
          files={uploadFiles}
          mountID={uploadMountID}
          mountOptions={mountOptions}
          purpose={uploadPurpose}
          saving={saving}
          onAccess={setUploadAccess}
          onAddFiles={addUploadFiles}
          onCategory={setUploadCategoryID}
          onClose={() => setShowUpload(false)}
          onMount={setUploadMountID}
          onPurpose={setUploadPurpose}
          onRemoveFile={removeUploadFile}
          onSubmit={onUpload}
        />
      ) : null}

      {showCategories ? (
        <CategoryModal
          categoryDescription={categoryDescription}
          categoryEditing={categoryEditing}
          categoryIcon={categoryIcon}
          categoryName={categoryName}
          categoryParentID={categoryParentID}
          categoryParentOptions={categoryParentOptions}
          categorySlug={categorySlug}
          categorySortOrder={categorySortOrder}
          categoryStatus={categoryStatus}
          saving={saving}
          onClose={() => {
            setShowCategories(false);
            resetCategoryForm();
          }}
          onDescription={setCategoryDescription}
          onDelete={
            categoryEditing
              ? () => {
                  setShowCategories(false);
                  setCategoryDeleting(categoryEditing);
                }
              : undefined
          }
          onIcon={setCategoryIcon}
          onName={setCategoryName}
          onParent={setCategoryParentID}
          onSlug={setCategorySlug}
          onSortOrder={setCategorySortOrder}
          onStatus={setCategoryStatus}
          onSubmit={onSaveCategory}
        />
      ) : null}

      {editing ? (
        <EditAttachmentModal
          access={editAccess}
          attachment={editing}
          categoryIDs={editCategoryIDs}
          categories={categories}
          name={editOriginalName}
          saving={saving}
          status={editStatus}
          onAccess={setEditAccess}
          onCategoryIDs={setEditCategoryIDs}
          onClose={() => setEditing(null)}
          onName={setEditOriginalName}
          onStatus={setEditStatus}
          onSubmit={onSaveEdit}
        />
      ) : null}

      {bulkEditing ? (
        <BulkAttachmentEditModal
          access={bulkAccess}
          categoryIDs={bulkCategoryIDs}
          categories={categories}
          count={selectedAttachmentIDs.length}
          saving={saving}
          status={bulkStatus}
          onAccess={setBulkAccess}
          onCategoryIDs={setBulkCategoryIDs}
          onClose={() => setBulkEditing(false)}
          onStatus={setBulkStatus}
          onSubmit={onSaveBulkEdit}
        />
      ) : null}

      {completing ? (
        <CompleteModal
          checksum={completeChecksum}
          etag={completeETag}
          saving={saving}
          size={completeSize}
          onChecksum={setCompleteChecksum}
          onClose={() => setCompleting(null)}
          onETag={setCompleteETag}
          onSize={setCompleteSize}
          onSubmit={onComplete}
        />
      ) : null}

      {referenceTarget ? (
        <ReferencesModal attachment={referenceTarget} references={referenceDetail} onClose={() => setReferenceTarget(null)} />
      ) : null}

      {deleting ? (
        <ConfirmModal
          body={deleteConfirmBody(deleting, referencesByAttachment)}
          danger
          saving={saving}
          title="删除附件"
          onCancel={() => setDeleting(null)}
          onConfirm={onConfirmDelete}
        />
      ) : null}

      {bulkDeleting ? (
        <ConfirmModal
          body={bulkDeleteConfirmBody(selectedAttachmentIDs.length, selectedReferencedCount)}
          danger
          saving={saving}
          title="批量删除附件"
          onCancel={() => setBulkDeleting(false)}
          onConfirm={onConfirmBulkDelete}
        />
      ) : null}

      {categoryDeleting ? (
        <ConfirmModal
          body={`确认删除 ${categoryDeleting.name}？后端会软删分组，已绑定附件不会被删除。`}
          danger
          saving={saving}
          title="删除分组"
          onCancel={() => setCategoryDeleting(null)}
          onConfirm={onConfirmDeleteCategory}
        />
      ) : null}
    </>
  );
}

function AttachmentListHeader(props: {
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <div className={styles.attachmentListHeader}>
      <label className={styles.selectCell} aria-label="选择当前页附件">
        <input checked={props.checked} type="checkbox" onChange={(event) => props.onChange(event.target.checked)} />
      </label>
      <span>预览</span>
      <span>名称 / 类型 / 路径</span>
      <span>存储实例</span>
      <span>上传者</span>
      <span>引用</span>
      <span>更新时间</span>
      <span>操作</span>
    </div>
  );
}

function CategoryCard({
  active,
  count,
  disabled,
  title,
  onClick,
  onEdit
}: {
  active?: boolean;
  count?: number;
  disabled?: boolean;
  title: string;
  onClick?: () => void;
  onEdit?: () => void;
}) {
  return (
    <button className={`${styles.categoryCard} ${active ? styles.categoryCardActive : ""}`} disabled={disabled} type="button" onClick={onClick}>
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

function AttachmentUploadModal(props: {
  access: string;
  categoryID: string;
  categoryOptions: { value: string; label: string }[];
  files: File[];
  mountID: string;
  mountOptions: { value: string; label: string }[];
  purpose: string;
  saving: boolean;
  onAccess: (value: string) => void;
  onAddFiles: (files: File[]) => void;
  onCategory: (value: string) => void;
  onClose: () => void;
  onMount: (value: string) => void;
  onPurpose: (value: string) => void;
  onRemoveFile: (index: number) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  function openFilePicker() {
    fileInputRef.current?.click();
  }

  function appendFiles(fileList: FileList | null) {
    if (!fileList?.length) return;
    props.onAddFiles(Array.from(fileList));
  }

  function handleFileInputChange(event: ChangeEvent<HTMLInputElement>) {
    appendFiles(event.target.files);
    event.target.value = "";
  }

  function handleDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    appendFiles(event.dataTransfer.files);
  }

  const uploadLabel = props.files.length > 0 ? `上传 ${props.files.length} 个附件` : "上传附件";

  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modalTall} role="dialog">
        <div className={styles.modalHeader}>
          <div>
            <h3>上传附件</h3>
            <p>未选择存储实例时，后端会使用默认启用存储。</p>
          </div>
          <button aria-label="关闭" className="icon-button" type="button" onClick={props.onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <form className={styles.formGrid} id="attachment-upload-form" onSubmit={props.onSubmit}>
          <label className={styles.field}>
            <span>类型</span>
            <StudioSelect ariaLabel="用途" options={purposeOptions.filter((option) => option.value)} value={props.purpose} onChange={props.onPurpose} />
          </label>
          <label className={styles.field}>
            <span>访问范围</span>
            <StudioSelect ariaLabel="访问范围" options={accessOptions} value={props.access} onChange={props.onAccess} />
          </label>
          <label className={styles.field}>
            <span>存储实例</span>
            <StudioSelect ariaLabel="存储实例" options={props.mountOptions} value={props.mountID} onChange={props.onMount} />
          </label>
          <label className={styles.field}>
            <span>分组</span>
            <StudioSelect ariaLabel="分组" options={props.categoryOptions} value={props.categoryID} onChange={props.onCategory} />
          </label>
          <div className={styles.uploadSection}>
            <div
              className={styles.uploadDropzone}
              role="button"
              tabIndex={0}
              onClick={openFilePicker}
              onDragOver={(event) => event.preventDefault()}
              onDrop={handleDrop}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  openFilePicker();
                }
              }}
            >
              <input
                ref={fileInputRef}
                className={styles.fileInputHidden}
                aria-label="文件"
                multiple
                type="file"
                onChange={handleFileInputChange}
              />
              <FileUp aria-hidden size={32} />
              <strong>
                拖拽文件到这里，或者
                <button
                  className={styles.uploadBrowseButton}
                  type="button"
                  onClick={(event) => {
                    event.stopPropagation();
                    openFilePicker();
                  }}
                >
                  浏览文件
                </button>
              </strong>
              <span>本地上传，可多选</span>
            </div>
            <div className={styles.uploadFileList} aria-label="待上传文件列表">
              {props.files.length === 0 ? (
                <p className={styles.uploadFileListEmpty}>尚未添加文件</p>
              ) : (
                props.files.map((file, index) => (
                  <div key={`${file.name}-${file.size}-${file.lastModified}-${index}`} className={styles.uploadFileRow}>
                    <div className={styles.uploadFileRowMeta}>
                      <strong>{file.name}</strong>
                      <span>{formatBytes(file.size)}</span>
                    </div>
                    <button
                      aria-label={`移除 ${file.name}`}
                      className="icon-button"
                      type="button"
                      onClick={() => props.onRemoveFile(index)}
                    >
                      <Trash2 aria-hidden size={16} />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </form>
        <div className={styles.modalActions}>
          <button className="secondary-button" type="button" onClick={props.onClose}>
            取消
          </button>
          <button
            className="primary-button"
            disabled={props.saving || props.files.length === 0}
            form="attachment-upload-form"
            type="submit"
          >
            {props.saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            {uploadLabel}
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function CategoryModal(props: {
  categoryDescription: string;
  categoryEditing: AttachmentCategoryResponse | null;
  categoryIcon: string;
  categoryName: string;
  categoryParentID: string;
  categoryParentOptions: { value: string; label: string }[];
  categorySlug: string;
  categorySortOrder: string;
  categoryStatus: string;
  saving: boolean;
  onClose: () => void;
  onDelete?: () => void;
  onDescription: (value: string) => void;
  onIcon: (value: string) => void;
  onName: (value: string) => void;
  onParent: (value: string) => void;
  onSlug: (value: string) => void;
  onSortOrder: (value: string) => void;
  onStatus: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modalWide} role="dialog">
        <div className={styles.modalHeader}>
          <div>
            <h3>管理附件分组</h3>
            <p>分组用于筛选和绑定附件；停用分组不会删除已有关联。</p>
          </div>
          <button aria-label="关闭" className="icon-button" type="button" onClick={props.onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <form className={styles.formGrid} id="attachment-category-form" onSubmit={props.onSubmit}>
          <label className={styles.field}>
            <span>名称</span>
            <input aria-label="分类名称" value={props.categoryName} onChange={(event) => props.onName(event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Slug</span>
            <input aria-label="分类 slug" value={props.categorySlug} onChange={(event) => props.onSlug(event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>父分组</span>
            <StudioSelect ariaLabel="父分类" options={props.categoryParentOptions} value={props.categoryParentID} onChange={props.onParent} />
          </label>
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="分类状态" options={categoryStatusOptions} value={props.categoryStatus} onChange={props.onStatus} />
          </label>
          <label className={styles.field}>
            <span>图标</span>
            <input aria-label="分类图标" value={props.categoryIcon} onChange={(event) => props.onIcon(event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>排序</span>
            <input aria-label="分类排序" type="number" value={props.categorySortOrder} onChange={(event) => props.onSortOrder(event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>描述</span>
            <textarea aria-label="分类描述" className={styles.textarea} value={props.categoryDescription} onChange={(event) => props.onDescription(event.target.value)} />
          </label>
        </form>
        <div className={styles.modalActions}>
          {props.onDelete ? (
            <button className="danger-button" type="button" onClick={props.onDelete}>
              删除分组
            </button>
          ) : null}
          <button className="primary-button" disabled={props.saving} form="attachment-category-form" type="submit">
            {props.saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            {props.categoryEditing ? "保存分组" : "创建分组"}
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function EditAttachmentModal(props: {
  access: string;
  attachment: AttachmentResponse;
  categoryIDs: number[];
  categories: AttachmentCategoryResponse[];
  name: string;
  saving: boolean;
  status: string;
  onAccess: (value: string) => void;
  onCategoryIDs: (value: number[]) => void;
  onClose: () => void;
  onName: (value: string) => void;
  onStatus: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  const attachmentName = displayName(props.attachment);
  const showImagePreview = isImageMime(props.attachment.mime_type);
  const imageSrc = attachmentContentUrl(props.attachment.id);
  const [imageZoomOpen, setImageZoomOpen] = useState(false);

  useEffect(() => {
    if (!imageZoomOpen) {
      return;
    }

    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === "Escape") {
        setImageZoomOpen(false);
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [imageZoomOpen]);

  return createPortal(
    <>
      <div className={styles.overlay} role="presentation" onClick={props.onClose}>
        <div aria-modal="true" className={styles.modalWide} role="dialog" onClick={stopRowClick}>
          <div className={styles.modalHeader}>
            <div>
              <h3>编辑附件</h3>
            {showImagePreview ? (
              <div className={styles.editAttachmentPreview}>
                <button
                  aria-label={`放大查看 ${attachmentName}`}
                  className={styles.editAttachmentPreviewButton}
                  type="button"
                  onClick={() => setImageZoomOpen(true)}
                >
                  <Image
                    alt={attachmentName}
                    className={styles.editAttachmentPreviewImage}
                    height={240}
                    src={imageSrc}
                    unoptimized
                    width={400}
                  />
                </button>
              </div>
            ) : (
              <p>{attachmentName}</p>
            )}
          </div>
          <button aria-label="关闭" className="icon-button" type="button" onClick={props.onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <form className={styles.formGrid} id="attachment-edit-form" onSubmit={props.onSubmit}>
          <label className={styles.fieldFull}>
            <span>展示名称</span>
            <input aria-label="展示名称" value={props.name} onChange={(event) => props.onName(event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>访问范围</span>
            <StudioSelect ariaLabel="编辑访问范围" options={accessOptions} value={props.access} onChange={props.onAccess} />
          </label>
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="编辑状态" options={statusOptions.filter((option) => option.value)} value={props.status} onChange={props.onStatus} />
          </label>
          <div className={styles.fieldFull}>
            <span className={styles.subsectionTitle}>分组</span>
            <div className={styles.checkboxGrid}>
              {props.categories.length === 0 ? (
                <span className={styles.muted}>暂无分组。</span>
              ) : (
                props.categories.map((category) => (
                  <label className={styles.inlineCheck} key={category.id}>
                    <input
                      checked={props.categoryIDs.includes(category.id)}
                      type="checkbox"
                      onChange={(event) => {
                        props.onCategoryIDs(
                          event.target.checked ? [...props.categoryIDs, category.id] : props.categoryIDs.filter((id) => id !== category.id)
                        );
                      }}
                    />
                    {category.name}
                  </label>
                ))
              )}
            </div>
          </div>
        </form>
        <div className={styles.modalActions}>
          <button className="secondary-button" type="button" onClick={props.onClose}>
            取消
          </button>
          <button className="primary-button" disabled={props.saving} form="attachment-edit-form" type="submit">
            保存
          </button>
        </div>
        </div>
      </div>
      {imageZoomOpen && showImagePreview ? (
        <div className={styles.imageZoomOverlay} role="presentation" onClick={() => setImageZoomOpen(false)}>
          <button
            aria-label="关闭放大预览"
            className={styles.imageZoomClose}
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              setImageZoomOpen(false);
            }}
          >
            <X aria-hidden size={22} strokeWidth={2.5} />
          </button>
          <div
            aria-label="图片放大预览"
            aria-modal="true"
            className={styles.imageZoomDialog}
            role="dialog"
            onClick={(event) => event.stopPropagation()}
          >
            <Image
              alt={attachmentName}
              className={styles.imageZoomImage}
              height={1080}
              src={imageSrc}
              unoptimized
              width={1920}
            />
          </div>
        </div>
      ) : null}
    </>,
    document.body
  );
}

function BulkAttachmentEditModal(props: {
  access: string;
  categoryIDs: number[];
  categories: AttachmentCategoryResponse[];
  count: number;
  saving: boolean;
  status: string;
  onAccess: (value: string) => void;
  onCategoryIDs: (value: number[]) => void;
  onClose: () => void;
  onStatus: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modalWide} role="dialog">
        <div className={styles.modalHeader}>
          <div>
            <h3>批量编辑附件</h3>
            <p>将对已选择的 {props.count} 个附件统一更新访问范围、状态和分组。</p>
          </div>
          <button aria-label="关闭" className="icon-button" type="button" onClick={props.onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        <form className={styles.formGrid} id="attachment-bulk-edit-form" onSubmit={props.onSubmit}>
          <label className={styles.field}>
            <span>访问范围</span>
            <StudioSelect ariaLabel="批量访问范围" options={accessOptions} value={props.access} onChange={props.onAccess} />
          </label>
          <label className={styles.field}>
            <span>状态</span>
            <StudioSelect ariaLabel="批量状态" options={statusOptions.filter((option) => option.value)} value={props.status} onChange={props.onStatus} />
          </label>
          <div className={styles.fieldFull}>
            <span className={styles.subsectionTitle}>分组</span>
            <div className={styles.checkboxGrid}>
              {props.categories.length === 0 ? (
                <span className={styles.muted}>暂无分组。</span>
              ) : (
                props.categories.map((category) => (
                  <label className={styles.inlineCheck} key={category.id}>
                    <input
                      checked={props.categoryIDs.includes(category.id)}
                      type="checkbox"
                      onChange={(event) => {
                        props.onCategoryIDs(
                          event.target.checked ? [...props.categoryIDs, category.id] : props.categoryIDs.filter((id) => id !== category.id)
                        );
                      }}
                    />
                    {category.name}
                  </label>
                ))
              )}
            </div>
          </div>
        </form>
        <div className={styles.modalActions}>
          <button className="secondary-button" type="button" onClick={props.onClose}>
            取消
          </button>
          <button className="primary-button" disabled={props.saving} form="attachment-bulk-edit-form" type="submit">
            {props.saving ? <Loader2 aria-hidden className="spin" size={18} /> : <Save aria-hidden size={18} />}
            保存批量修改
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function CompleteModal(props: {
  checksum: string;
  etag: string;
  saving: boolean;
  size: string;
  onChecksum: (value: string) => void;
  onClose: () => void;
  onETag: (value: string) => void;
  onSize: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modal} role="dialog">
        <h3>标记远端附件完成</h3>
        <form className={styles.formGrid} id="attachment-complete-form" onSubmit={props.onSubmit}>
          <label className={styles.fieldFull}>
            <span>ETag</span>
            <input aria-label="ETag" value={props.etag} onChange={(event) => props.onETag(event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>Checksum</span>
            <input aria-label="Checksum" value={props.checksum} onChange={(event) => props.onChecksum(event.target.value)} />
          </label>
          <label className={styles.fieldFull}>
            <span>大小（字节）</span>
            <input aria-label="完成大小" type="number" value={props.size} onChange={(event) => props.onSize(event.target.value)} />
          </label>
        </form>
        <div className={styles.modalActions}>
          <button className="secondary-button" type="button" onClick={props.onClose}>
            取消
          </button>
          <button className="primary-button" disabled={props.saving} form="attachment-complete-form" type="submit">
            完成
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function ReferencesModal(props: { attachment: AttachmentResponse; references: AttachmentReferenceResponse[]; onClose: () => void }) {
  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modal} role="dialog">
        <div className={styles.modalHeader}>
          <div>
            <h3>附件引用</h3>
            <p>{displayName(props.attachment)}</p>
          </div>
          <button aria-label="关闭" className="icon-button" type="button" onClick={props.onClose}>
            <X aria-hidden size={18} />
          </button>
        </div>
        {props.references.length === 0 ? (
          <div className={styles.emptyState}>暂无引用。</div>
        ) : (
          <div className={styles.referenceList}>
            {props.references.map((item) => (
              <div className={styles.referenceItem} key={`${item.source_type}-${item.source_id}-${item.relation}`}>
                <strong>{item.source_title}</strong>
                <span>{item.source_type} · {item.relation} · {item.status}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>,
    document.body
  );
}

function ConfirmModal(props: {
  body: string;
  danger?: boolean;
  saving: boolean;
  title: string;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return createPortal(
    <div className={styles.overlay} role="presentation">
      <div aria-modal="true" className={styles.modal} role="dialog">
        <h3>{props.title}</h3>
        <p>{props.body}</p>
        <div className={styles.modalActions}>
          <button className="secondary-button" type="button" onClick={props.onCancel}>
            取消
          </button>
          <button className={props.danger ? "danger-button" : "primary-button"} disabled={props.saving} type="button" onClick={props.onConfirm}>
            确认
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

function displayName(attachment: AttachmentResponse) {
  return attachment.original_name || attachment.filename;
}

function mountLabel(id: number, mounts: StorageMountResponse[]) {
  const mount = mounts.find((item) => item.id === id);
  return mount ? mount.name : `#${id}`;
}

function ownerLabel(attachment: AttachmentResponse) {
  return attachment.owner_username ?? (attachment.owner_user_id ? `#${attachment.owner_user_id}` : "-");
}

function hasAttachmentReferences(id: number, referencesByAttachment: Map<number, AttachmentReferenceResponse[]>) {
  return (referencesByAttachment.get(id)?.length ?? 0) > 0;
}

function deleteConfirmBody(attachment: AttachmentResponse, referencesByAttachment: Map<number, AttachmentReferenceResponse[]>) {
  const references = referencesByAttachment.get(attachment.id) ?? [];
  if (references.length === 0) {
    return `确认删除 ${displayName(attachment)}？没有业务引用的附件会直接删除。`;
  }
  return `确认删除 ${displayName(attachment)}？当前附件被 ${references.length} 处业务对象引用；确认后会删除附件，并将用户头像等引用切换为默认头像。`;
}

function bulkDeleteConfirmBody(total: number, referencedCount: number) {
  if (referencedCount === 0) {
    return `确认删除已选择的 ${total} 个附件？没有业务引用的附件会直接删除。`;
  }
  return `确认删除已选择的 ${total} 个附件？其中 ${referencedCount} 个附件仍被业务对象引用；确认后会删除附件，并将用户头像等引用切换为默认头像。`;
}

function uploadFileKey(file: File) {
  return `${file.name}:${file.size}:${file.lastModified}`;
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
