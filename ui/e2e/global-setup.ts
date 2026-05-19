import { execSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const e2eDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(e2eDir, "../..");

const defaultDsn =
  "postgres://Beehive-Blog:Beehive-Blog@127.0.0.1:5432/Beehive-Blog?sslmode=disable";

export default async function globalSetup() {
  const dsn = process.env.E2E_DATABASE_DSN ?? defaultDsn;

  try {
    execSync(`go run ./sql/migrate/ -dsn "${dsn}" -mode versioned`, {
      cwd: repoRoot,
      stdio: "inherit",
      env: process.env
    });
  } catch (error) {
    const hint =
      "Ensure Postgres is running (docker compose -f docker/Infrastructure/docker-compose.yaml up -d).";
    throw new Error(`E2E database migration failed for DSN ${dsn}. ${hint}`, { cause: error });
  }
}
