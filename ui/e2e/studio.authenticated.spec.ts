import { expect, test } from "@playwright/test";

test.describe("Studio authenticated", () => {
  test("admin can open settings", async ({ page }) => {
    await page.goto("/studio/settings");

    await expect(page).not.toHaveURL(/\/login/);
    await expect(page.getByRole("heading", { name: "设置" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Email" })).toBeVisible();
  });
});
