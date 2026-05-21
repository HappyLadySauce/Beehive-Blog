import { expect, test } from "@playwright/test";

test.describe("Public Web smoke", () => {
  test("home page renders hero and latest posts", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { level: 1 })).toContainText("知识蜂巢");
    await expect(page.getByRole("heading", { name: "最新文章" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "最新笔记" })).toBeVisible();
    await expect(page.getByRole("heading", { exact: true, name: "项目" })).toBeVisible();
    await expect(page.locator("section[aria-labelledby='latest-posts']").getByText(/公开文章|发布第一篇公开文章/)).toBeVisible();
  });

  test("posts index renders public articles or an empty state", async ({ page }) => {
    await page.goto("/posts");

    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.getByText(/暂无公开文章|分钟阅读/)).toBeVisible();
  });

  test("notes index renders public notes or an empty state", async ({ page }) => {
    await page.goto("/notes");

    await expect(page.getByRole("heading", { level: 1, name: "笔记" })).toBeVisible();
    await expect(page.getByText(/暂无公开笔记|分钟阅读/)).toBeVisible();
  });

  test("projects index renders public projects or an empty state", async ({ page }) => {
    await page.goto("/projects");

    await expect(page.getByRole("heading", { level: 1, name: "项目" })).toBeVisible();
    await expect(page.getByText(/暂无公开项目|分钟阅读/)).toBeVisible();
  });
});
