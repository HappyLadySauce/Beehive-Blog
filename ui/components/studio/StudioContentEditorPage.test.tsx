import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { StudioContentEditorPage } from "./StudioContentEditorPage";

const replace = vi.hoisted(() => vi.fn());
const uploadLocalAttachmentsBatch = vi.hoisted(() => vi.fn());
const createContent = vi.hoisted(() => vi.fn());
const getContent = vi.hoisted(() => vi.fn());
const updateContent = vi.hoisted(() => vi.fn());
const transitionContentStatus = vi.hoisted(() => vi.fn());
const setContentCategories = vi.hoisted(() => vi.fn());
const setContentTags = vi.hoisted(() => vi.fn());
const listContentVersions = vi.hoisted(() => vi.fn());
const upsertAutoContentVersion = vi.hoisted(() => vi.fn());
const listCategories = vi.hoisted(() => vi.fn());
const listTags = vi.hoisted(() => vi.fn());

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace })
}));

vi.mock("@/components/auth/AuthProvider", () => ({
  useAuth: () => ({
    claims: { uid: 1, role: "admin" }
  })
}));

vi.mock("@/lib/api/attachments", () => ({
  attachmentContentUrl: (id: number) => `/api/bff/attachments/${id}/content`,
  uploadLocalAttachmentsBatch
}));

vi.mock("@/lib/api/contents", async () => {
  const { planStatusTransition } = await import("@/lib/content-status");
  return {
    createContent,
    getContent,
    setContentCategories,
    setContentTags,
    transitionContentStatus,
    transitionContentStatusTo: vi.fn(async (id: number, from: string, to: string) => {
      let last = null;
      for (const status of planStatusTransition(from, to)) {
        last = await transitionContentStatus(id, { status });
      }
      return last;
    }),
    updateContent,
    listContentVersions,
    upsertAutoContentVersion
  };
});

vi.mock("@/lib/api/categories", () => ({
  listCategories
}));

vi.mock("@/lib/api/tags", () => ({
  listTags
}));

vi.mock("./StudioMarkdownCodeMirror", () => {
  return {
    StudioMarkdownCodeMirror: ({
      mode,
      scrollTarget,
    value,
    onChange,
    onFiles,
    onSelectionChange
  }: {
    mode: "live" | "source";
    scrollTarget?: { id: number; line: number } | null;
    value: string;
    onChange: (value: string) => void;
    onFiles: (files: File[]) => void;
    onSelectionChange: (selection: { from: number; to: number }) => void;
  }) => {
    return (
      <textarea
        aria-label="Markdown 正文"
        data-mode={mode}
        data-scroll-line={scrollTarget?.line ?? ""}
        value={mode === "live" && !value.includes("\n") ? value.replace(/^#{1,4}\s+/, "") : value}
        onChange={(event) => {
          onChange(event.currentTarget.value);
          onSelectionChange({ from: event.currentTarget.value.length, to: event.currentTarget.value.length });
        }}
        onFocus={() => {
          if (mode === "live" && value.startsWith("## ")) onChange(value);
        }}
        onDrop={(event) => {
          event.preventDefault();
          onFiles(Array.from(event.dataTransfer.files));
        }}
      />
    );
  }
  };
});

const tags = {
  items: [
    { id: 3, name: "AI", slug: "ai", created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }
  ],
  page: 1,
  page_size: 100,
  total: 1
};

const categories = {
  items: [
    { id: 5, name: "工程", slug: "engineering", parent_id: null, sort_order: 0, created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }
  ],
  page: 1,
  page_size: 100,
  total: 1
};

const contentDetail = {
  ai_access: "allowed",
  author_id: 1,
  author_username: "admin",
  body: "# Existing",
  created_at: "2026-05-15T00:00:00Z",
  excerpt: "Excerpt",
  id: 9,
  reading_time_minutes: 1,
  slug: "hello-world",
  status: "draft",
  tags: tags.items,
  title: "Hello World",
  type: "article",
  updated_at: "2026-05-15T00:00:00Z",
  view_count: 0,
  visibility: "public",
  word_count: 10
};

function renderEditor(mode: "create" | "edit" = "create") {
  return render(
    <ToastProvider>
      <StudioContentEditorPage contentId={mode === "edit" ? 9 : undefined} mode={mode} />
    </ToastProvider>
  );
}

describe("StudioContentEditorPage", () => {
  beforeEach(() => {
    replace.mockReset();
    uploadLocalAttachmentsBatch.mockReset();
    createContent.mockReset();
    getContent.mockReset();
    updateContent.mockReset();
    transitionContentStatus.mockReset();
    setContentCategories.mockReset();
    setContentTags.mockReset();
    listCategories.mockReset();
    listTags.mockReset();
    listContentVersions.mockReset();
    upsertAutoContentVersion.mockReset();
    listCategories.mockResolvedValue(categories);
    listTags.mockResolvedValue(tags);
    listContentVersions.mockResolvedValue({ items: [] });
    upsertAutoContentVersion.mockResolvedValue({ id: 1 });
    getContent.mockResolvedValue(contentDetail);
    createContent.mockResolvedValue({ id: 10 });
    updateContent.mockResolvedValue(contentDetail);
    transitionContentStatus.mockResolvedValue(contentDetail);
    setContentCategories.mockResolvedValue({});
    setContentTags.mockResolvedValue({});
    uploadLocalAttachmentsBatch.mockResolvedValue({
      failed: 0,
      items: [
        {
          attachment: {
            access_scope: "public",
            created_at: "2026-05-15T00:00:00Z",
            filename: "diagram.png",
            id: 77,
            mime_type: "image/png",
            object_key: "content/diagram.png",
            original_name: "diagram.png",
            purpose: "content",
            size: 128,
            status: "active",
            storage_mount_id: 10,
            updated_at: "2026-05-15T00:00:00Z",
            upload_status: "ready"
          },
          filename: "diagram.png",
          index: 0
        }
      ],
      uploaded: 1
    });
  });

  it("creates content, binds tags, and redirects to edit page", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText("标题"), { target: { value: "New Post" } });
    fireEvent.change(screen.getByLabelText("Slug"), { target: { value: "new-post" } });
    fireEvent.change(screen.getByLabelText("Markdown 正文"), { target: { value: "# Draft" } });
    fireEvent.click(screen.getByLabelText("AI"));
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createContent).toHaveBeenCalledWith(expect.objectContaining({ body: "# Draft", slug: "new-post", title: "New Post" })));
    expect(setContentTags).toHaveBeenCalledWith(10, { tag_ids: [3] });
    expect(replace).toHaveBeenCalledWith("/studio/content/10/edit");
  });

  it("keeps editor chrome in fixed panes and shows metadata in the left rail", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    const navigation = screen.getByLabelText("文档导航");
    expect(navigation).toBeInTheDocument();
    expect(screen.getByLabelText("内容属性")).toBeInTheDocument();
    expect(screen.getByLabelText("编辑器顶部栏")).toBeInTheDocument();
    expect(within(navigation).getByText("未设置 Slug")).toBeInTheDocument();
    expect(within(navigation).getByText(/字符 \/ 1 词/)).toBeInTheDocument();
  });

  it("generates a heading outline and sends line targets to the editor", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText("Markdown 正文"), { target: { value: "# 角色定位\n\n## 核心任务\n\n### 关键输入" } });

    fireEvent.click(screen.getByRole("button", { name: "核心任务" }));

    expect(screen.getByLabelText("Markdown 正文")).toHaveAttribute("data-scroll-line", "3");
  });

  it("switches between live, source, and preview modes", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText("Markdown 正文"), { target: { value: "## 标题" } });
    expect(screen.getByRole("button", { name: /编辑预览/ }).className).toContain("editorModeTabActive");
    expect(screen.getByLabelText("Markdown 正文")).toHaveValue("标题");

    fireEvent.click(screen.getByRole("button", { name: /源码/ }));
    expect(screen.getByLabelText("Markdown 正文")).toHaveAttribute("data-mode", "source");

    fireEvent.click(screen.getByRole("button", { name: "预览" }));
    expect(screen.getByLabelText("Markdown 预览")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "标题" })).toBeInTheDocument();
    expect(screen.queryByLabelText("Markdown 正文")).not.toBeInTheDocument();
  });

  it("does not render the old markdown formatting toolbar above the editor", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    expect(screen.queryByRole("button", { name: "加粗" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "撤销" })).not.toBeInTheDocument();
  });

  it("loads existing content, updates it, and transitions status", async () => {
    renderEditor("edit");

    await waitFor(() => expect(getContent).toHaveBeenCalledWith(9));
    expect(screen.getByLabelText("标题")).toHaveValue("Hello World");
    fireEvent.change(screen.getByLabelText("Markdown 正文"), { target: { value: "# Updated" } });
    fireEvent.click(screen.getByRole("combobox", { name: "内容状态" }));
    fireEvent.click(screen.getByRole("option", { name: "已发布" }));
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(updateContent).toHaveBeenCalledWith(9, expect.objectContaining({ body: "# Updated" })));
    expect(setContentTags).toHaveBeenCalledWith(9, { tag_ids: [3] });
    expect(transitionContentStatus).toHaveBeenCalledTimes(2);
    expect(transitionContentStatus).toHaveBeenNthCalledWith(1, 9, { status: "review" });
    expect(transitionContentStatus).toHaveBeenNthCalledWith(2, 9, { status: "published" });
    expect(upsertAutoContentVersion).toHaveBeenCalledWith(9, { change_summary: "Auto-saved latest content changes" });
  });

  it("does not render content relations in the editor sidebar", async () => {
    renderEditor("edit");

    await waitFor(() => expect(getContent).toHaveBeenCalledWith(9));

    expect(screen.queryByText("内容关联")).not.toBeInTheDocument();
  });

  it("does not update the auto version when only taxonomy changes", async () => {
    renderEditor("edit");

    await waitFor(() => expect(getContent).toHaveBeenCalledWith(9));
    fireEvent.click(screen.getByLabelText("AI"));
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(updateContent).toHaveBeenCalledWith(9, expect.objectContaining({ title: "Hello World" })));
    expect(setContentTags).toHaveBeenCalledWith(9, { tag_ids: [] });
    expect(upsertAutoContentVersion).not.toHaveBeenCalled();
  });

  it("imports markdown files into the editor body", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("导入 Markdown 文件")).toBeInTheDocument());
    const file = new File(["# Imported\n\nBody"], "imported.md", { type: "text/markdown" });
    fireEvent.change(screen.getByLabelText("导入 Markdown 文件"), { target: { files: [file] } });

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toHaveValue("# Imported\n\nBody"));
  });

  it("uploads dropped editor assets and inserts markdown references", async () => {
    renderEditor("create");

    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText("Markdown 正文"), { target: { value: "# Draft" } });
    const image = new File(["image"], "diagram.png", { type: "image/png" });
    fireEvent.drop(screen.getByLabelText("Markdown 正文"), { dataTransfer: { files: [image] } });

    await waitFor(() => expect(uploadLocalAttachmentsBatch).toHaveBeenCalledTimes(1));
    const formData = uploadLocalAttachmentsBatch.mock.calls[0][0] as FormData;
    expect(formData.get("owner_user_id")).toBe("1");
    expect(formData.get("purpose")).toBe("content");
    expect(formData.get("access_scope")).toBe("public");
    await waitFor(() => expect(screen.getByLabelText("Markdown 正文")).toHaveValue("# Draft\n\n![diagram.png](/api/bff/attachments/77/content)"));
  });
});
