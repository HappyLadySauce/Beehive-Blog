import { expect, test } from "@playwright/test";

test.describe("Public Web smoke", () => {
  test("home page renders hero and latest posts", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { level: 1 })).toContainText("知识蜂巢");
    await expect(page.getByRole("heading", { name: "最新文章" })).toBeVisible();
    await expect(page.getByRole("link", { name: /AI 协作写作回路/ })).toBeVisible();
  });

  test("posts index lists seeded public articles", async ({ page }) => {
    await page.goto("/posts");

    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.getByRole("link", { name: /Public Web 与 Studio 的产品边界/ })).toBeVisible();
  });
});
