import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioStoragePage } from "./StudioStoragePage";

const listFileDrivers = vi.hoisted(() => vi.fn());
const listStorageMounts = vi.hoisted(() => vi.fn());
const createStorageMount = vi.hoisted(() => vi.fn());
const updateStorageMount = vi.hoisted(() => vi.fn());
const enableStorageMount = vi.hoisted(() => vi.fn());
const disableStorageMount = vi.hoisted(() => vi.fn());
const checkStorageMount = vi.hoisted(() => vi.fn());
const deleteStorageMount = vi.hoisted(() => vi.fn());
const listAttachments = vi.hoisted(() => vi.fn());
const listAttachmentCategories = vi.hoisted(() => vi.fn());
const uploadLocalAttachment = vi.hoisted(() => vi.fn());
const updateAttachment = vi.hoisted(() => vi.fn());
const deleteAttachment = vi.hoisted(() => vi.fn());
const completeAttachment = vi.hoisted(() => vi.fn());
const createAttachmentCategory = vi.hoisted(() => vi.fn());
const updateAttachmentCategory = vi.hoisted(() => vi.fn());
const deleteAttachmentCategory = vi.hoisted(() => vi.fn());

vi.mock("@/components/auth/AuthProvider", () => ({
  useAuth: () => ({
    claims: { uid: 1, role: "admin" }
  })
}));

vi.mock("@/lib/api/storage", () => ({
  checkStorageMount,
  createStorageMount,
  deleteStorageMount,
  disableStorageMount,
  enableStorageMount,
  listFileDrivers,
  listStorageMounts,
  updateStorageMount
}));

vi.mock("@/lib/api/attachments", () => ({
  attachmentContentUrl: (id: number) => `/api/bff/attachments/${id}/content`,
  completeAttachment,
  createAttachmentCategory,
  deleteAttachmentCategory,
  deleteAttachment,
  listAttachmentCategories,
  listAttachments,
  updateAttachmentCategory,
  updateAttachment,
  uploadLocalAttachment
}));

const drivers = {
  items: [
    {
      id: 1,
      name: "local",
      display_name: "Local Storage",
      description: "Server-local filesystem storage.",
      config_schema: { required: ["root"] },
      capabilities: { upload: true, download: true, delete: true },
      status: "active",
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

const mounts = {
  items: [
    {
      id: 10,
      driver_name: "local",
      mount_path: "/local",
      name: "Local Storage",
      remark: null,
      config: { root: "data/attachments" },
      order_index: 0,
      is_default: true,
      disabled: false,
      status: "work",
      last_checked_at: null,
      last_error: null,
      created_by: null,
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

const attachments = {
  items: [
    {
      id: 99,
      purpose: "content",
      filename: "note.md",
      original_name: "Note.md",
      mime_type: "text/markdown",
      size: 128,
      storage_mount_id: 10,
      object_key: "content/note.md",
      access_scope: "private",
      upload_status: "ready",
      status: "active",
      category_ids: [7],
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

const categories = {
  items: [
    {
      id: 7,
      name: "文章素材",
      slug: "posts",
      path: "/posts",
      depth: 0,
      sort_order: 0,
      status: "active",
      created_at: "2026-05-15T00:00:00Z",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

describe("StudioStoragePage", () => {
  beforeEach(() => {
    listFileDrivers.mockReset();
    listStorageMounts.mockReset();
    createStorageMount.mockReset();
    updateStorageMount.mockReset();
    enableStorageMount.mockReset();
    disableStorageMount.mockReset();
    checkStorageMount.mockReset();
    deleteStorageMount.mockReset();
    listAttachments.mockReset();
    listAttachmentCategories.mockReset();
    uploadLocalAttachment.mockReset();
    updateAttachment.mockReset();
    deleteAttachment.mockReset();
    completeAttachment.mockReset();
    createAttachmentCategory.mockReset();
    updateAttachmentCategory.mockReset();
    deleteAttachmentCategory.mockReset();

    listFileDrivers.mockResolvedValue(drivers);
    listStorageMounts.mockResolvedValue(mounts);
    listAttachments.mockResolvedValue(attachments);
    listAttachmentCategories.mockResolvedValue(categories);
    uploadLocalAttachment.mockResolvedValue(attachments.items[0]);
    updateAttachment.mockResolvedValue(attachments.items[0]);
    deleteAttachment.mockResolvedValue({});
    completeAttachment.mockResolvedValue(attachments.items[0]);
    createAttachmentCategory.mockResolvedValue(categories.items[0]);
    updateAttachmentCategory.mockResolvedValue(categories.items[0]);
    deleteAttachmentCategory.mockResolvedValue({});
    createStorageMount.mockResolvedValue(mounts.items[0]);
    updateStorageMount.mockResolvedValue(mounts.items[0]);
    enableStorageMount.mockResolvedValue(mounts.items[0]);
    disableStorageMount.mockResolvedValue({ ...mounts.items[0], disabled: true });
    checkStorageMount.mockResolvedValue({ status: "work", checked: "2026-05-15T00:00:00Z" });
    deleteStorageMount.mockResolvedValue({});
  });

  it("loads files as the default storage workspace segment", async () => {
    render(<StudioStoragePage />);

    expect(screen.getByText("正在加载文件...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());
    expect(screen.getByText("content/note.md")).toBeInTheDocument();
    expect(screen.getByText("Local Storage")).toBeInTheDocument();
  });

  it("loads drivers and storage mounts from the storage mount segment", async () => {
    render(<StudioStoragePage />);

    fireEvent.click(screen.getByRole("button", { name: "存储实例" }));
    await waitFor(() => expect(screen.getByText("/local")).toBeInTheDocument());
    expect(screen.getByText("默认")).toBeInTheDocument();
  });

  it("creates a local storage mount with JSON config", async () => {
    listStorageMounts.mockResolvedValueOnce({ items: [] });
    render(<StudioStoragePage />);
    fireEvent.click(screen.getByRole("button", { name: "存储实例" }));
    await waitFor(() => expect(screen.getByRole("button", { name: "添加存储" })).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "添加存储" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "Uploads" } });
    fireEvent.change(screen.getByLabelText("挂载路径"), { target: { value: "/uploads" } });
    fireEvent.change(screen.getByLabelText("驱动配置 JSON"), { target: { value: '{ "root": "data/uploads" }' } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createStorageMount).toHaveBeenCalled());
    expect(createStorageMount.mock.calls[0][0]).toMatchObject({
      config: { root: "data/uploads" },
      driver_name: "local",
      mount_path: "/uploads",
      name: "Uploads"
    });
  });

  it("runs health checks and disables mounts", async () => {
    render(<StudioStoragePage />);
    fireEvent.click(screen.getByRole("button", { name: "存储实例" }));
    await waitFor(() => expect(screen.getByText("/local")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "检查" }));
    await waitFor(() => expect(checkStorageMount).toHaveBeenCalledWith(10));

    fireEvent.click(screen.getByRole("button", { name: "禁用" }));
    await waitFor(() => expect(disableStorageMount).toHaveBeenCalledWith(10));
  });

  it("uploads a local attachment from the file segment", async () => {
    render(<StudioStoragePage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "上传文件" }));
    const file = new File(["hello"], "hello.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("文件", { selector: "input" }), { target: { files: [file] } });
    fireEvent.click(screen.getByRole("button", { name: "上传" }));

    await waitFor(() => expect(uploadLocalAttachment).toHaveBeenCalled());
    const formData = uploadLocalAttachment.mock.calls[0][0] as FormData;
    expect(formData.get("purpose")).toBe("content");
    expect(formData.get("access_scope")).toBe("private");
    expect(formData.get("owner_user_id")).toBe("1");
  });

  it("edits attachment metadata", async () => {
    render(<StudioStoragePage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("编辑文件"));
    fireEvent.change(screen.getByLabelText("展示名称"), { target: { value: "Updated.md" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(updateAttachment).toHaveBeenCalledWith(99, expect.objectContaining({ original_name: "Updated.md" })));
  });

  it("creates, updates, and deletes attachment categories", async () => {
    render(<StudioStoragePage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "管理分类" }));
    fireEvent.change(screen.getByLabelText("分类名称"), { target: { value: "资料" } });
    fireEvent.change(screen.getByLabelText("分类 slug"), { target: { value: "assets" } });
    fireEvent.click(screen.getByRole("button", { name: "创建分类" }));
    await waitFor(() => expect(createAttachmentCategory).toHaveBeenCalledWith(expect.objectContaining({ name: "资料", slug: "assets" })));

    fireEvent.click(screen.getByLabelText("编辑分类 文章素材"));
    fireEvent.change(screen.getByLabelText("分类名称"), { target: { value: "文章资料" } });
    fireEvent.click(screen.getByRole("button", { name: "保存分类" }));
    await waitFor(() => expect(updateAttachmentCategory).toHaveBeenCalledWith(7, expect.objectContaining({ name: "文章资料" })));

    fireEvent.click(screen.getByLabelText("删除分类 文章素材"));
    fireEvent.click(screen.getByRole("button", { name: "删除" }));
    await waitFor(() => expect(deleteAttachmentCategory).toHaveBeenCalledWith(7));
  });
});
