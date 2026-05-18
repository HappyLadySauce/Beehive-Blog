import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { StudioAttachmentsPage } from "./StudioAttachmentsPage";

const listStorageMounts = vi.hoisted(() => vi.fn());
const listAttachments = vi.hoisted(() => vi.fn());
const listAttachmentCategories = vi.hoisted(() => vi.fn());
const listAttachmentReferences = vi.hoisted(() => vi.fn());
const getAttachmentReferences = vi.hoisted(() => vi.fn());
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
  listStorageMounts
}));

vi.mock("@/lib/api/attachments", () => ({
  attachmentContentUrl: (id: number) => `/api/bff/attachments/${id}/content`,
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
  uploadLocalAttachment
}));

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
      owner_user_id: 1,
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

const references = {
  items: [
    {
      attachment_id: 99,
      source_type: "user",
      source_id: 1,
      source_title: "Admin",
      relation: "avatar",
      status: "active",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ]
};

describe("StudioAttachmentsPage", () => {
  beforeEach(() => {
    listStorageMounts.mockReset();
    listAttachments.mockReset();
    listAttachmentCategories.mockReset();
    listAttachmentReferences.mockReset();
    getAttachmentReferences.mockReset();
    uploadLocalAttachment.mockReset();
    updateAttachment.mockReset();
    deleteAttachment.mockReset();
    completeAttachment.mockReset();
    createAttachmentCategory.mockReset();
    updateAttachmentCategory.mockReset();
    deleteAttachmentCategory.mockReset();

    listStorageMounts.mockResolvedValue(mounts);
    listAttachments.mockResolvedValue(attachments);
    listAttachmentCategories.mockResolvedValue(categories);
    listAttachmentReferences.mockResolvedValue(references);
    getAttachmentReferences.mockResolvedValue(references);
    uploadLocalAttachment.mockResolvedValue(attachments.items[0]);
    updateAttachment.mockResolvedValue(attachments.items[0]);
    deleteAttachment.mockResolvedValue({});
    completeAttachment.mockResolvedValue(attachments.items[0]);
    createAttachmentCategory.mockResolvedValue(categories.items[0]);
    updateAttachmentCategory.mockResolvedValue(categories.items[0]);
    deleteAttachmentCategory.mockResolvedValue({});
  });

  it("loads the attachment library with categories and references", async () => {
    render(<StudioAttachmentsPage />);

    expect(screen.getByText("正在加载附件...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());
    expect(screen.getByText("附件库")).toBeInTheDocument();
    expect(screen.getByText("文章素材")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "1 引用" })).toBeInTheDocument();
  });

  it("passes search and reference filters to the attachment API", async () => {
    render(<StudioAttachmentsPage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText("搜索附件"), { target: { value: "note" } });
    fireEvent.click(screen.getByRole("combobox", { name: "按引用筛选" }));
    fireEvent.click(screen.getByRole("option", { name: "孤儿附件" }));

    await waitFor(() =>
      expect(listAttachments).toHaveBeenLastCalledWith(expect.objectContaining({ reference_status: "orphan", search: "note" }))
    );
  });

  it("uploads a local attachment", async () => {
    render(<StudioAttachmentsPage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "上传" }));
    const file = new File(["hello"], "hello.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("文件"), { target: { files: [file] } });
    fireEvent.click(screen.getByRole("button", { name: "上传附件" }));

    await waitFor(() => expect(uploadLocalAttachment).toHaveBeenCalled());
    const formData = uploadLocalAttachment.mock.calls[0][0] as FormData;
    expect(formData.get("purpose")).toBe("content");
    expect(formData.get("access_scope")).toBe("private");
    expect(formData.get("owner_user_id")).toBe("1");
  });

  it("opens the reference dialog", async () => {
    render(<StudioAttachmentsPage />);
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "1 引用" }));
    await waitFor(() => expect(getAttachmentReferences).toHaveBeenCalledWith(99));
    expect(screen.getByText("Admin")).toBeInTheDocument();
  });
});
