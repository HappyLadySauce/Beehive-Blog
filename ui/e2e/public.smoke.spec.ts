import { expect, test } from "@playwright/test";

test.describe("Public Web smoke", () => {
  test("home page renders hero and latest posts", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByText("Theme-AnZhiYu")).toBeVisible();
    await expect(page.getByRole("link", { name: "首页", exact: true })).toBeVisible();
    await expect(page.getByRole("complementary", { name: "侧栏" }).getByRole("heading", { name: "标签" })).toBeVisible();
    await expect(page.locator("section[aria-labelledby='latest-posts']").getByText(/分钟|暂无公开文章/).first()).toBeVisible();
  });

  test("home page supports search and shortcut panel", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    await page.locator(".anz-header-actions").getByRole("button", { name: "站内搜索" }).click();
    await expect(page.getByRole("dialog", { name: "站内搜索" })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog", { name: "站内搜索" })).toBeHidden();

    await page.keyboard.down("Shift");
    await expect(page.getByText("博客快捷键")).toBeVisible();
    await page.keyboard.up("Shift");
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
