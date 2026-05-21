# Beehive Blog UI

Next.js App Router frontend for Beehive-Blog.

## Commands

```bash
pnpm install
pnpm dev
pnpm test
pnpm test:e2e
pnpm lint
pnpm build
```

## E2E (Playwright)

Prerequisites:

1. Start Postgres and Redis: `docker compose -f docker/Infrastructure/docker-compose.yaml up -d` (from repo root).
2. Install Chromium once: `pnpm exec playwright install chromium`.

Run the full stack E2E suite (DB migrate, Go API, Next dev, then tests):

```bash
pnpm test:e2e
```

Debug interactively:

```bash
pnpm test:e2e:ui
```

Optional env overrides:

- `E2E_DATABASE_DSN` — Postgres DSN for migrations (default matches docker-compose).
- `E2E_ADMIN_USER` / `E2E_ADMIN_PASSWORD` — Studio login (default `admin` / `Admin@123` from SQL seed).
- `PLAYWRIGHT_BASE_URL` — Next base URL (default `http://127.0.0.1:3000`).
- `E2E_GO_API_URL` — Go API URL for rewrites and webServer health (default `http://127.0.0.1:8080`).

## Environment

- `BEEHIVE_API_BASE_URL`: server-side Go API target for BFF and Public SSR content, defaults to `http://localhost:8080`.
- `NEXT_PUBLIC_API_BASE_URL`: browser API base, defaults to `/api/v1`.
- `NEXT_PUBLIC_SITE_URL`: canonical site URL, defaults to `http://localhost:3000`.
- `PUBLIC_CONTENT_ENDPOINT`: deprecated development fallback for seeded-style public JSON posts. Public pages now read Go `/api/v1/contents` for articles, notes, and projects; production does not show fallback posts when the Go API is unavailable.
