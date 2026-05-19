import { defineConfig, devices } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";

const uiDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(uiDir, "..");
const e2eDir = path.join(uiDir, "e2e");
const e2eConfigPath = path.join(repoRoot, "e2e", "config.beehive-blog.yaml");
const authFile = path.join(uiDir, "e2e", ".auth", "admin.json");

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://127.0.0.1:3000";
const goApiURL = process.env.E2E_GO_API_URL ?? "http://127.0.0.1:8080";
const reuseExistingServer = !process.env.CI;

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [["list"], ["html", { open: "never" }]],
  globalSetup: path.join(e2eDir, "global-setup.ts"),
  use: {
    baseURL,
    trace: "on-first-retry"
  },
  projects: [
    {
      name: "setup",
      testMatch: /auth\.setup\.ts/
    },
    {
      name: "chromium",
      testMatch: /\.spec\.ts/,
      testIgnore: /studio\.authenticated\.spec\.ts/,
      use: { ...devices["Desktop Chrome"] }
    },
    {
      name: "studio-authenticated",
      testMatch: /studio\.authenticated\.spec\.ts/,
      dependencies: ["setup"],
      use: {
        ...devices["Desktop Chrome"],
        storageState: authFile
      }
    }
  ],
  webServer: [
    {
      command: `go run ./cmd/ --config ${e2eConfigPath}`,
      cwd: repoRoot,
      url: `${goApiURL}/livez`,
      reuseExistingServer,
      timeout: 120_000,
      stdout: "pipe",
      stderr: "pipe"
    },
    {
      command: "pnpm run dev",
      cwd: uiDir,
      url: baseURL,
      reuseExistingServer,
      timeout: 120_000,
      env: {
        ...process.env,
        BEEHIVE_API_BASE_URL: goApiURL
      },
      stdout: "pipe",
      stderr: "pipe"
    }
  ]
});
