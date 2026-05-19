import { expect, test } from "@playwright/test";

test.describe("Studio route guard", () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test("redirects unauthenticated users to login with next param", async ({ page }) => {
    await page.goto("/studio/settings");

    await expect(page).toHaveURL(/\/login/);
    expect(new URL(page.url()).searchParams.get("next")).toBe("/studio/settings");
    await expect(page.getByRole("heading", { name: "登录 Studio" })).toBeVisible();
  });
});
