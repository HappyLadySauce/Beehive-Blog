import { expect, test as setup } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const e2eDir = path.dirname(fileURLToPath(import.meta.url));
const authFile = path.join(e2eDir, ".auth", "admin.json");

const adminUser = process.env.E2E_ADMIN_USER ?? "admin";
const adminPassword = process.env.E2E_ADMIN_PASSWORD ?? "Admin@123";

setup("authenticate as admin", async ({ request }) => {
  fs.mkdirSync(path.dirname(authFile), { recursive: true });

  const response = await request.post("/api/bff/auth/login", {
    data: {
      grant_type: "local",
      account: adminUser,
      password: adminPassword
    }
  });

  expect(response.ok(), `login failed: ${response.status()} ${await response.text()}`).toBeTruthy();
  await request.storageState({ path: authFile });
});
