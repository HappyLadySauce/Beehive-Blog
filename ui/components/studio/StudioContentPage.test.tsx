import { StrictMode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { resetContentPageModuleStateForTests, StudioContentPage } from "./StudioContentPage";

const listContents = vi.hoisted(() => vi.fn());
const createContent = vi.hoisted(() => vi.fn());
const updateContent = vi.hoisted(() => vi.fn());
const deleteContent = vi.hoisted(() => vi.fn());
const transitionContentStatus = vi.hoisted(() => vi.fn());
const setContentTags = vi.hoisted(() => vi.fn());
const listTags = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/contents", () => ({
  createContent,
  deleteContent,
  listContents,
  setContentTags,
  transitionContentStatus,
  updateContent
}));

vi.mock("@/lib/api/tags", () => ({
  listTags
}));

const contentList = {
  items: [
    {
      ai_access: "allowed",
      author_id: 1,
      author_username: "admin",
      body: "Body",
      created_at: "2026-05-15T00:00:00Z",
      excerpt: "Excerpt",
      id: 9,
      reading_time_minutes: 1,
      slug: "hello-world",
      status: "draft",
      tags: [{ id: 3, name: "AI", slug: "ai", status: "active", created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }],
      title: "Hello World",
      type: "article",
      updated_at: "2026-05-15T00:00:00Z",
      view_count: 0,
      visibility: "public",
      word_count: 10
    }
  ],
  page: 1,
  page_size: 20,
  total: 1
};

const tags = {
  items: [
    { id: 3, name: "AI", slug: "ai", status: "active", created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }
  ],
  page: 1,
  page_size: 100,
  total: 1
};

function renderContentPage(strict = false) {
  const element = (
    <ToastProvider>
      <StudioContentPage />
    </ToastProvider>
  );
  return render(strict ? <StrictMode>{element}</StrictMode> : element);
}

describe("StudioContentPage", () => {
  beforeEach(() => {
    resetContentPageModuleStateForTests();
    listContents.mockReset();
    createContent.mockReset();
    updateContent.mockReset();
    deleteContent.mockReset();
    transitionContentStatus.mockReset();
    setContentTags.mockReset();
    listTags.mockReset();
    listContents.mockResolvedValue(contentList);
    listTags.mockResolvedValue(tags);
    createContent.mockResolvedValue({ id: 10 });
    updateContent.mockResolvedValue(contentList.items[0]);
    transitionContentStatus.mockResolvedValue(contentList.items[0]);
    setContentTags.mockResolvedValue({});
    deleteContent.mockResolvedValue({});
  });

  it("loads contents and active tag metadata", async () => {
    renderContentPage();

    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());
    expect(screen.getByRole("region", { name: "内容工作台" })).toBeInTheDocument();
    expect(screen.getByText("AI")).toBeInTheDocument();
    expect(listTags).toHaveBeenCalledWith({ page: 1, page_size: 100, status: "active" });
    expect(listContents).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20 }));
  });

  it("keeps the content table visible when the list is empty", async () => {
    listContents.mockResolvedValue({ ...contentList, items: [], total: 0 });

    renderContentPage();

    await waitFor(() => expect(screen.getByText("暂无内容")).toBeInTheDocument());
    expect(screen.getByTestId("content-category-reserved-strip")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "标题" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "操作" })).toBeInTheDocument();
    expect(screen.getByText("共 1 页")).toBeInTheDocument();
  });

  it("dedupes identical in-flight list and metadata requests", async () => {
    let resolveContents: (value: typeof contentList) => void = () => undefined;
    let resolveTags: (value: typeof tags) => void = () => undefined;
    listContents.mockReturnValue(new Promise((resolve) => { resolveContents = resolve; }));
    listTags.mockReturnValue(new Promise((resolve) => { resolveTags = resolve; }));

    renderContentPage(true);
    resolveContents(contentList);
    resolveTags(tags);

    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());
    expect(listContents).toHaveBeenCalledTimes(1);
    expect(listTags).toHaveBeenCalledTimes(1);
  });

  it("creates content, binds tags, and refreshes through direct network request", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建内容" }));
    fireEvent.change(screen.getByLabelText("标题"), { target: { value: "New Post" } });
    fireEvent.change(screen.getByLabelText("Slug"), { target: { value: "new-post" } });
    fireEvent.click(screen.getByLabelText("AI"));
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(createContent).toHaveBeenCalledWith(expect.objectContaining({ slug: "new-post", title: "New Post" })));
    expect(setContentTags).toHaveBeenCalledWith(10, { tag_ids: [3] });
    expect(listContents).toHaveBeenCalledTimes(2);
  });

  it("updates status and deletes content", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "编辑 Hello World" }));
    fireEvent.click(screen.getByRole("combobox", { name: "内容状态" }));
    fireEvent.click(screen.getByRole("option", { name: "已发布" }));
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(transitionContentStatus).toHaveBeenCalledWith(9, { status: "published" }));

    fireEvent.click(screen.getByRole("button", { name: "删除 Hello World" }));
    fireEvent.click(screen.getByRole("button", { name: "删除" }));
    await waitFor(() => expect(deleteContent).toHaveBeenCalledWith(9));
  });
});
