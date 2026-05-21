import { StrictMode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { resetContentPageModuleStateForTests, StudioContentPage } from "./StudioContentPage";

const listContents = vi.hoisted(() => vi.fn());
const deleteContent = vi.hoisted(() => vi.fn());
const listTags = vi.hoisted(() => vi.fn());
const createCategory = vi.hoisted(() => vi.fn());
const deleteCategory = vi.hoisted(() => vi.fn());
const listCategories = vi.hoisted(() => vi.fn());
const updateCategory = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/contents", () => ({
  deleteContent,
  listContents
}));

vi.mock("@/lib/api/tags", () => ({
  listTags
}));

vi.mock("@/lib/api/categories", () => ({
  createCategory,
  deleteCategory,
  listCategories,
  updateCategory
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
      tags: [{ id: 3, name: "AI", slug: "ai", created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }],
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
    { id: 3, name: "AI", slug: "ai", created_at: "2026-05-15T00:00:00Z", updated_at: "2026-05-15T00:00:00Z" }
  ],
  page: 1,
  page_size: 100,
  total: 1
};

const categories = {
  items: [
    {
      content_count: 2,
      created_at: "2026-05-15T00:00:00Z",
      id: 7,
      name: "工程",
      parent_id: null,
      slug: "engineering",
      sort_order: 0,
      updated_at: "2026-05-15T00:00:00Z"
    }
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
    deleteContent.mockReset();
    listTags.mockReset();
    createCategory.mockReset();
    deleteCategory.mockReset();
    listCategories.mockReset();
    updateCategory.mockReset();
    listContents.mockResolvedValue(contentList);
    listTags.mockResolvedValue(tags);
    listCategories.mockResolvedValue(categories);
    createCategory.mockResolvedValue({ id: 12 });
    updateCategory.mockResolvedValue(categories.items[0]);
    deleteCategory.mockResolvedValue({});
    deleteContent.mockResolvedValue({});
  });

  it("loads contents and active tag metadata", async () => {
    renderContentPage();

    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());
    expect(screen.getByRole("region", { name: "内容工作台" })).toBeInTheDocument();
    expect(screen.getByText("AI")).toBeInTheDocument();
    expect(screen.getByText("工程")).toBeInTheDocument();
    expect(listTags).toHaveBeenCalledWith({ page: 1, page_size: 100 });
    expect(listCategories).toHaveBeenCalledWith({ page: 1, page_size: 100 });
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
    let resolveCategories: (value: typeof categories) => void = () => undefined;
    listContents.mockReturnValue(new Promise((resolve) => { resolveContents = resolve; }));
    listTags.mockReturnValue(new Promise((resolve) => { resolveTags = resolve; }));
    listCategories.mockReturnValue(new Promise((resolve) => { resolveCategories = resolve; }));

    renderContentPage(true);
    resolveContents(contentList);
    resolveTags(tags);
    resolveCategories(categories);

    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());
    expect(listContents).toHaveBeenCalledTimes(1);
    expect(listTags).toHaveBeenCalledTimes(1);
    expect(listCategories).toHaveBeenCalledTimes(1);
  });

  it("filters contents from the inline category strip", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("工程")).toBeInTheDocument());

    fireEvent.click(screen.getByText("工程").closest("button")!);

    await waitFor(() => expect(listContents).toHaveBeenLastCalledWith(expect.objectContaining({ category_id: 7 })));
  });

  it("creates categories from the content page reserved strip", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("工程")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "新建" }));
    fireEvent.change(screen.getByLabelText("名称"), { target: { value: "架构" } });
    fireEvent.click(screen.getByRole("button", { name: "创建分类" }));

    await waitFor(() => expect(createCategory).toHaveBeenCalledWith(expect.objectContaining({
      name: "架构",
      slug: expect.stringMatching(/^[a-z0-9]+(-[a-z0-9]+)*$/)
    })));
    expect(listCategories).toHaveBeenCalledTimes(2);
  });

  it("links create and edit actions to standalone editor routes", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());

    expect(screen.getByRole("link", { name: /新建内容/ })).toHaveAttribute("href", "/studio/content/new");
    expect(screen.getByRole("link", { name: "编辑 Hello World" })).toHaveAttribute("href", "/studio/content/9/edit");
  });

  it("deletes content and refreshes through direct network request", async () => {
    renderContentPage();
    await waitFor(() => expect(screen.getByText("Hello World")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "删除 Hello World" }));
    fireEvent.click(screen.getByRole("button", { name: "删除" }));
    await waitFor(() => expect(deleteContent).toHaveBeenCalledWith(9));
    expect(listContents).toHaveBeenCalledTimes(2);
  });
});
