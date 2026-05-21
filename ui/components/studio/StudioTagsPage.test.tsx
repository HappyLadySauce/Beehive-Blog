import { StrictMode } from "react";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { ApiError } from "@/lib/api/client";
import { resetTagsPageModuleStateForTests, StudioTagsPage } from "./StudioTagsPage";

const listTags = vi.hoisted(() => vi.fn());
const createTag = vi.hoisted(() => vi.fn());
const updateTag = vi.hoisted(() => vi.fn());
const deleteTag = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/tags", () => ({
  createTag,
  deleteTag,
  listTags,
  updateTag
}));

const tags = {
  items: [
    {
      color: "#2f8f79",
      content_count: 2,
      created_at: "2026-05-15T00:00:00Z",
      description: "AI topics",
      id: 3,
      name: "AI",
      slug: "ai",
      updated_at: "2026-05-15T00:00:00Z"
    }
  ],
  page: 1,
  page_size: 20,
  total: 1
};

function renderTagsPage(strict = false) {
  const element = (
    <ToastProvider>
      <StudioTagsPage />
    </ToastProvider>
  );
  return render(strict ? <StrictMode>{element}</StrictMode> : element);
}

describe("StudioTagsPage", () => {
  beforeEach(() => {
    resetTagsPageModuleStateForTests();
    listTags.mockReset();
    createTag.mockReset();
    updateTag.mockReset();
    deleteTag.mockReset();
    listTags.mockResolvedValue(tags);
    createTag.mockResolvedValue({ id: 4 });
    updateTag.mockResolvedValue(tags.items[0]);
    deleteTag.mockResolvedValue({});
  });

  it("loads tags with content count", async () => {
    renderTagsPage();

    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());
    expect(screen.getByText("AI topics")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(listTags).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20 }));
  });

  it("keeps the tags table visible when the list is empty", async () => {
    listTags.mockResolvedValue({ ...tags, items: [], total: 0 });

    renderTagsPage();

    await waitFor(() => expect(screen.getByText("暂无标签")).toBeInTheDocument());
    expect(screen.queryByTestId("content-category-reserved-strip")).not.toBeInTheDocument();
    expect(screen.queryByTestId("attachment-category-strip")).not.toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "标签" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Slug" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "操作" })).toBeInTheDocument();
    expect(screen.getByText("共 1 页")).toBeInTheDocument();
  });

  it("dedupes identical in-flight list requests", async () => {
    let resolveTags: (value: typeof tags) => void = () => undefined;
    listTags.mockReturnValue(new Promise((resolve) => { resolveTags = resolve; }));

    renderTagsPage(true);
    resolveTags(tags);

    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());
    expect(listTags).toHaveBeenCalledTimes(1);
  });

  it("creates and updates tags", async () => {
    renderTagsPage();
    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建标签" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "Product" } });
    fireEvent.change(screen.getByLabelText("颜色"), { target: { value: "#123456" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createTag).toHaveBeenCalledWith(expect.objectContaining({ color: "#123456", name: "Product", slug: "product" })));

    fireEvent.click(screen.getByRole("button", { name: "编辑 AI" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "AI Updated" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(updateTag).toHaveBeenCalledWith(3, expect.objectContaining({ name: "AI Updated" })));
  });

  it("lets users pick tag colors from a palette or native color picker", async () => {
    renderTagsPage();
    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建标签" }));
    fireEvent.click(screen.getByRole("button", { name: "选择颜色 #2563eb" }));
    expect(screen.getByLabelText("颜色")).toHaveValue("#2563eb");

    fireEvent.change(screen.getByLabelText("颜色选择器"), { target: { value: "#db2777" } });
    expect(screen.getByLabelText("颜色")).toHaveValue("#db2777");

    fireEvent.click(screen.getByRole("button", { name: "清空颜色" }));
    expect(screen.getByLabelText("颜色")).toHaveValue("");
  });

  it("trims blank tag color before validation and mutation", async () => {
    renderTagsPage();
    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建标签" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "Product" } });
    fireEvent.change(screen.getByLabelText("颜色"), { target: { value: "   " } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createTag).toHaveBeenCalledWith(expect.objectContaining({ color: null, name: "Product", slug: "product" })));
  });

  it("auto-generates slug from Chinese tag name on create", async () => {
    renderTagsPage();
    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建标签" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "产品动态" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createTag).toHaveBeenCalledWith(expect.objectContaining({
      name: "产品动态",
      slug: expect.stringMatching(/^[a-z0-9]+(-[a-z0-9]+)*$/)
    })));
  });

  it("shows upstream conflict when referenced tag deletion fails", async () => {
    deleteTag.mockRejectedValue(new ApiError("tag is referenced by 2 content item(s); remove references first", 409, 409));
    renderTagsPage();
    await waitFor(() => expect(screen.getByText("AI")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "删除 AI" }));
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "删除" }));

    await waitFor(() => expect(deleteTag).toHaveBeenCalledWith(3));
    await waitFor(() => expect(screen.getByText("tag is referenced by 2 content item(s); remove references first")).toBeInTheDocument());
  });
});
