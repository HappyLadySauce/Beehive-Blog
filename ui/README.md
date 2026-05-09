# Beehive Blog UI

Next.js App Router frontend for Beehive-Blog.

## Commands

```bash
pnpm install
pnpm dev
pnpm test
pnpm lint
pnpm build
```

## Environment

- `BEEHIVE_API_BASE_URL`: server-side rewrite target for Go, defaults to `http://localhost:8080`.
- `NEXT_PUBLIC_API_BASE_URL`: browser API base, defaults to `/api/v1`.
- `NEXT_PUBLIC_SITE_URL`: canonical site URL, defaults to `http://localhost:3000`.
- `PUBLIC_CONTENT_ENDPOINT`: optional public JSON endpoint for SSR posts. If omitted, the UI uses seeded public posts until content APIs are available.
