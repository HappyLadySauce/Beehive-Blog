import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/lib/api/content", () => ({
  getPublicContent: vi.fn(async () => ({
    type: "note",
    typeLabel: "笔记",
    href: "/notes/public-note",
    slug: "public-note",
    title: "公开笔记",
    description: "笔记摘要",
    body: "## 笔记标题\n\n- 记录项",
    publishedAt: "2026-05-21T00:00:00.000Z",
    tags: ["Note"],
    readingMinutes: 1
  })),
  listPublicContents: vi.fn(async () => [])
}));

import NoteDetailPage from "./page";

describe("NoteDetailPage", () => {
  it("renders markdown body without exposing raw heading syntax", async () => {
    render(await NoteDetailPage({ params: Promise.resolve({ slug: "public-note" }) }));

    expect(screen.getByRole("heading", { level: 2, name: "笔记标题" })).toBeInTheDocument();
    expect(screen.getByText("记录项")).toBeInTheDocument();
    expect(screen.queryByText("## 笔记标题")).not.toBeInTheDocument();
  });
});
