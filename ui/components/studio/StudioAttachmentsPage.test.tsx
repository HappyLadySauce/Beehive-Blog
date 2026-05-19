import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { StudioAttachmentsPage } from "./StudioAttachmentsPage";

const listStorageMounts = vi.hoisted(() => vi.fn());
const listAttachments = vi.hoisted(() => vi.fn());
const listAttachmentCategories = vi.hoisted(() => vi.fn());
const listAttachmentReferences = vi.hoisted(() => vi.fn());
const getAttachmentReferences = vi.hoisted(() => vi.fn());
const uploadLocalAttachment = vi.hoisted(() => vi.fn());
const uploadLocalAttachmentsBatch = vi.hoisted(() => vi.fn());
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

vi.mock("next/image", () => ({
  default: (props: { alt: string; src: string }) => <svg aria-label={props.alt} data-src={props.src} role="img" />
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
  uploadLocalAttachment,
  uploadLocalAttachmentsBatch
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
      owner_username: "admin",
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
  ],
  total: 1,
  page: 1,
  page_size: 100
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

function renderAttachmentsPage() {
  return render(
    <ToastProvider>
      <StudioAttachmentsPage />
    </ToastProvider>
  );
}

describe("StudioAttachmentsPage", () => {
  beforeEach(() => {
    listStorageMounts.mockReset();
    listAttachments.mockReset();
    listAttachmentCategories.mockReset();
    listAttachmentReferences.mockReset();
    getAttachmentReferences.mockReset();
    uploadLocalAttachment.mockReset();
    uploadLocalAttachmentsBatch.mockReset();
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
    uploadLocalAttachmentsBatch.mockImplementation(async (formData: FormData) => {
      const files = formData.getAll("files") as File[];
      for (const file of files) {
        const itemFormData = new FormData();
        itemFormData.set("file", file);
        for (const key of ["owner_user_id", "purpose", "access_scope", "storage_mount_id", "category_ids"]) {
          const value = formData.get(key);
          if (value) itemFormData.set(key, value);
        }
        uploadLocalAttachment(itemFormData);
      }
      return {
        failed: 0,
        items: files.map((file, index) => ({ attachment: attachments.items[0], filename: file.name, index })),
        uploaded: files.length
      };
    });
    updateAttachment.mockResolvedValue(attachments.items[0]);
    deleteAttachment.mockResolvedValue({});
    completeAttachment.mockResolvedValue(attachments.items[0]);
    createAttachmentCategory.mockResolvedValue(categories.items[0]);
    updateAttachmentCategory.mockResolvedValue(categories.items[0]);
    deleteAttachmentCategory.mockResolvedValue({});
  });

  it("loads the attachment library with categories and references", async () => {
    renderAttachmentsPage();

    expect(screen.getByText("正在加载附件...")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());
    expect(screen.getByText("附件库")).toBeInTheDocument();
    expect(screen.getByText("文章素材")).toBeInTheDocument();
    expect(screen.getByText("名称 / 类型 / 路径")).toBeInTheDocument();
    expect(screen.getByText("存储实例")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "1 引用" })).toBeInTheDocument();
  });

  it("requests attachments with batched offset pagination", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    expect(listAttachments).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 100 }));
  });

  it("shows pagination when all results fit on one page", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    expect(screen.getByRole("navigation", { name: "分页" })).toBeInTheDocument();
    expect(screen.getByText("共 1 页")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "第 1 页" })).toHaveAttribute("aria-current", "page");
  });

  it("opens the edit dialog when clicking an attachment row", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Note.md"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "编辑附件" })).toBeInTheDocument();
  });

  it("closes the edit dialog when clicking the backdrop", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Note.md"));
    const dialog = screen.getByRole("dialog");
    fireEvent.click(dialog.parentElement!);

    expect(screen.queryByRole("heading", { name: "编辑附件" })).not.toBeInTheDocument();
  });

  it("shows an image preview in the edit dialog for image attachments", async () => {
    listAttachments.mockResolvedValue({
      ...attachments,
      items: [{ ...attachments.items[0], mime_type: "image/png" }]
    });

    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Note.md"));
    const preview = screen.getByRole("img", { name: "Note.md" });
    expect(preview).toHaveAttribute("data-src", "/api/bff/attachments/99/content");
  });

  it("opens a zoomed preview when clicking the edit dialog image", async () => {
    listAttachments.mockResolvedValue({
      ...attachments,
      items: [{ ...attachments.items[0], mime_type: "image/png" }]
    });

    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Note.md"));
    fireEvent.click(screen.getByRole("button", { name: "放大查看 Note.md" }));

    expect(screen.getByRole("dialog", { name: "图片放大预览" })).toBeInTheDocument();
    expect(screen.getAllByRole("img", { name: "Note.md" })).toHaveLength(2);
  });

  it("does not open the edit dialog when toggling row selection", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("选择附件 Note.md"));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("passes search and reference filters to the attachment API", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText("搜索附件"), { target: { value: "note" } });
    fireEvent.click(screen.getByRole("combobox", { name: "按引用筛选" }));
    fireEvent.click(screen.getByRole("option", { name: "孤儿附件" }));

    await waitFor(() =>
      expect(listAttachments).toHaveBeenLastCalledWith(
        expect.objectContaining({ reference_status: "orphan", search: "note", page: 1, page_size: 100 })
      )
    );
  });

  it("switches pages within the current batch without refetching", async () => {
    const batchedAttachments = {
      items: Array.from({ length: 25 }, (_, index) => ({
        ...attachments.items[0],
        id: index + 1,
        original_name: `File-${index + 1}.md`,
        filename: `file-${index + 1}.md`
      })),
      total: 25,
      page: 1,
      page_size: 100
    };
    listAttachments.mockResolvedValue(batchedAttachments);

    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("File-1.md")).toBeInTheDocument());
    expect(listAttachments).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole("button", { name: "第 2 页" }));
    expect(screen.getByText("File-11.md")).toBeInTheDocument();
    expect(listAttachments).toHaveBeenCalledTimes(1);
  });

  it("fetches the next batch when jumping beyond the first ten pages", async () => {
    const firstBatch = {
      items: Array.from({ length: 100 }, (_, index) => ({
        ...attachments.items[0],
        id: index + 1,
        original_name: `File-${index + 1}.md`,
        filename: `file-${index + 1}.md`
      })),
      total: 250,
      page: 1,
      page_size: 100
    };
    const secondBatch = {
      items: Array.from({ length: 100 }, (_, index) => ({
        ...attachments.items[0],
        id: index + 101,
        original_name: `File-${index + 101}.md`,
        filename: `file-${index + 101}.md`
      })),
      total: 250,
      page: 2,
      page_size: 100
    };
    listAttachments.mockImplementation((params) => {
      if (params.page === 2) return Promise.resolve(secondBatch);
      return Promise.resolve(firstBatch);
    });

    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("File-1.md")).toBeInTheDocument());

    for (let step = 0; step < 10; step += 1) {
      fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    }
    await waitFor(() => expect(screen.getByText("File-101.md")).toBeInTheDocument());
    expect(listAttachments).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2, page_size: 100 }));
  });

  it("passes the unassigned category filter to the attachment API", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "未分组 0" }));

    await waitFor(() => expect(listAttachments).toHaveBeenLastCalledWith(expect.objectContaining({ category_mode: "unassigned" })));
  });

  it("uploads a local attachment", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "上传" }));
    const file = new File(["hello"], "hello.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("文件"), { target: { files: [file] } });
    expect(screen.getByText("hello.txt")).toBeInTheDocument();
    expect(screen.getByText("5 B")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "上传 1 个附件" }));

    await waitFor(() => expect(uploadLocalAttachment).toHaveBeenCalledTimes(1));
    const formData = uploadLocalAttachment.mock.calls[0][0] as FormData;
    expect(formData.get("purpose")).toBe("content");
    expect(formData.get("access_scope")).toBe("public");
    expect(formData.get("owner_user_id")).toBe("1");
  });

  it("uploads multiple local attachments", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "上传" }));
    const file1 = new File(["hello"], "hello.txt", { type: "text/plain" });
    const file2 = new File(["world"], "world.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("文件"), { target: { files: [file1, file2] } });
    expect(screen.getByText("hello.txt")).toBeInTheDocument();
    expect(screen.getByText("world.txt")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "上传 2 个附件" }));

    await waitFor(() => expect(uploadLocalAttachment).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(screen.getByText("已上传 2 个附件。")).toBeInTheDocument());
  });

  it("uploads remaining files after removing one from the queue", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "上传" }));
    const file1 = new File(["hello"], "hello.txt", { type: "text/plain" });
    const file2 = new File(["world"], "world.txt", { type: "text/plain" });
    fireEvent.change(screen.getByLabelText("文件"), { target: { files: [file1, file2] } });
    fireEvent.click(screen.getByRole("button", { name: "移除 hello.txt" }));
    expect(screen.queryByText("hello.txt")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "上传 1 个附件" }));

    await waitFor(() => expect(uploadLocalAttachment).toHaveBeenCalledTimes(1));
    const formData = uploadLocalAttachment.mock.calls[0][0] as FormData;
    expect((formData.get("file") as File).name).toBe("world.txt");
  });

  it("bulk edits selected attachments", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("选择附件 Note.md"));
    fireEvent.click(screen.getByRole("button", { name: "编辑已选" }));
    fireEvent.click(screen.getByRole("button", { name: "保存批量修改" }));

    await waitFor(() =>
      expect(updateAttachment).toHaveBeenCalledWith(
        99,
        expect.objectContaining({ access_scope: "private", category_ids: [7], status: "active" })
      )
    );
  });

  it("does not show clear selection in the selection bar", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("选择附件 Note.md"));
    expect(screen.getByRole("button", { name: "批量删除" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "清除选择" })).not.toBeInTheDocument();
  });

  it("bulk deletes selected attachments", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("选择附件 Note.md"));
    fireEvent.click(screen.getByRole("button", { name: "批量删除" }));
    expect(screen.getByText("确认删除已选择的 1 个附件？其中 1 个附件仍被业务对象引用；确认后会删除附件，并将用户头像等引用切换为默认头像。")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "确认" }));

    await waitFor(() => expect(deleteAttachment).toHaveBeenCalledWith(99, { force: true }));
    await waitFor(() => expect(screen.getByText("附件已删除。")).toBeInTheDocument());
    expect(screen.queryByRole("button", { name: "批量删除" })).not.toBeInTheDocument();
  });

  it("keeps force deletion for referenced selections after loading another attachment batch", async () => {
    const firstBatch = {
      items: [attachments.items[0]],
      total: 250,
      page: 1,
      page_size: 100
    };
    const secondBatch = {
      items: [
        {
          ...attachments.items[0],
          id: 101,
          original_name: "File-101.md",
          filename: "file-101.md"
        }
      ],
      total: 250,
      page: 2,
      page_size: 100
    };
    listAttachments.mockImplementation((params) => {
      if (params.page === 2) return Promise.resolve(secondBatch);
      return Promise.resolve(firstBatch);
    });

    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText("选择附件 Note.md"));
    for (let step = 0; step < 10; step += 1) {
      fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    }
    await waitFor(() => expect(screen.getByText("File-101.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "批量删除" }));
    expect(screen.getByText("确认删除已选择的 1 个附件？其中 1 个附件仍被业务对象引用；确认后会删除附件，并将用户头像等引用切换为默认头像。")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "确认" }));

    await waitFor(() => expect(deleteAttachment).toHaveBeenCalledWith(99, { force: true }));
  });

  it("opens the reference dialog", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "1 引用" }));
    await waitFor(() => expect(getAttachmentReferences).toHaveBeenCalledWith(99));
    expect(screen.getByText("Admin")).toBeInTheDocument();
  });

  it("closes the category dialog after creating a category", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建" }));
    fireEvent.change(screen.getByLabelText("分类名称"), { target: { value: "图库" } });
    fireEvent.change(screen.getByLabelText("分类 slug"), { target: { value: "gallery" } });
    fireEvent.click(screen.getByRole("button", { name: "创建分组" }));

    await waitFor(() => expect(createAttachmentCategory).toHaveBeenCalled());
    expect(screen.queryByRole("button", { name: "创建分组" })).not.toBeInTheDocument();
    expect(await screen.findByText("分组已创建。")).toBeInTheDocument();
  });

  it("deletes an existing attachment category from the edit dialog", async () => {
    renderAttachmentsPage();
    await waitFor(() => expect(screen.getByText("Note.md")).toBeInTheDocument());

    const categoryCard = screen.getByText("文章素材").closest("button");
    const editIcon = categoryCard?.querySelector("svg");
    expect(editIcon).toBeTruthy();

    fireEvent.click(editIcon as SVGElement);
    expect(screen.getByRole("button", { name: "删除分组" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "删除分组" }));
    expect(screen.queryByRole("button", { name: "保存分组" })).not.toBeInTheDocument();
    expect(screen.getByText("确认删除 文章素材？后端会软删分组，已绑定附件不会被删除。")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "确认" }));
    await waitFor(() => expect(deleteAttachmentCategory).toHaveBeenCalledWith(7));
  });
});
