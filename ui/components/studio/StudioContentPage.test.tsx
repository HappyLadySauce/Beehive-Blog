import { StrictMode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ToastProvider } from "@/components/toast/ToastProvider";
import { resetContentPageModuleStateForTests, StudioContentPage } from "./StudioContentPage";

const listContents = vi.hoisted(() => vi.fn());
const deleteContent = vi.hoisted(() => vi.fn());
const listTags = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api/contents", () => ({
  deleteContent,
  listContents
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
    deleteContent.mockReset();
    listTags.mockReset();
    listContents.mockResolvedValue(contentList);
    listTags.mockResolvedValue(tags);
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
