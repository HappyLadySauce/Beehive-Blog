---
version: alpha
name: Beehive-Blog-design-system
description: "Beehive Blog is a knowledge-work product with two surfaces: an SEO-first public reading site and a dense administrator Studio for writing, organizing, reviewing, publishing, storage, users, permissions, and settings. The visual system uses warm paper canvas, crisp ink text, restrained teal actions, and honey accents. It should feel like a calm personal knowledge middle office, not a marketing landing page or a decorative CMS theme."
---

# Beehive Blog DESIGN.md

This file is the root design contract for AI agents and engineers changing the Beehive-Blog UI.

本文档是 Beehive-Blog UI 的根级设计契约，供 AI agent 与工程师在修改界面时读取。

It must stay aligned with:

- `docs/product-principles.md`
- `docs/frontend/react-ssr-seo-architecture.md`
- `docs/frontend/studio-layout-implementation-plan.md`
- `ui/app/globals.css`
- `ui/components/studio/Studio.module.css`

## 1. Visual Theme & Atmosphere

Beehive Blog is a personal blog, AI-assisted writing workspace, and agent-readable personal knowledge middle office. Its UI must communicate careful editing, durable knowledge, review gates, and controlled publishing.

Beehive Blog 是个人博客、AI 协作创作空间，以及面向智能体的个人知识中台。界面必须传达严谨整理、长期知识沉淀、人工审阅与受控发布。

### Product Surfaces

| Surface | Design role | Required feeling |
| --- | --- | --- |
| Public Web | Reading, discovery, SEO, published narrative | Editorial, readable, quiet, content-first |
| Studio | Admin-only creation, storage, settings, users, review, publication | Dense, scannable, reliable, operational |
| Auth | Login, register, OAuth callback, account state | Secure, direct, low-friction |

### Atmosphere

- Warm paper canvas with crisp black text.
- Teal is the primary action and active navigation color.
- Honey is a supporting brand signal, not the dominant palette.
- Studio should feel like a workbench: compact panels, clear status, obvious actions.
- Public pages should feel like a reading surface: semantic HTML, comfortable line length, strong metadata, no unnecessary widgets.

### Avoid

- Marketing-style split hero layouts for Studio.
- Decorative gradient orbs, bokeh blobs, heavy glassmorphism, or ornamental backgrounds.
- Nested cards inside decorative cards.
- A palette dominated by only beige, only teal, or only honey.
- Huge display typography inside dashboards, sidebars, cards, tables, and forms.

## 2. Color Palette & Roles

Use semantic tokens first. Existing CSS variables in `ui/app/globals.css` are the implementation anchor.

优先使用语义色。当前实现以 `ui/app/globals.css` 中的 CSS 变量为锚点。

| Token | Hex | Role |
| --- | --- | --- |
| `canvas` / `--bg` | `#f7f5ee` | App background and public page base |
| `surface` / `--surface` | `#fffdf8` | Panels, cards, dropdowns, modals |
| `surface-soft` | `#fbfaf6` | Code blocks, quiet inset areas, read-only JSON |
| `ink` / `--ink` | `#191713` | Primary text, icons, headings |
| `muted` / `--muted` | `#6f695f` | Secondary text, metadata, helper copy |
| `line` / `--line` | `#ded7ca` | Borders, dividers, table rules |
| `primary` / `--accent` | `#0f766e` | Primary CTA, active nav, focus affinity |
| `primary-strong` / `--accent-dark` | `#115e59` | Hover, selected text, links |
| `primary-soft` | `#edf7f5` | Active nav fill, success-like quiet state |
| `honey` / `--honey` | `#d89422` | Brand accent, warning-adjacent highlights |
| `warning-soft` | `#fff7e6` | Pending states and caution backgrounds |
| `warning` | `#925d05` | Pending text and non-destructive warning |
| `success` | `#16794c` | Confirmed healthy or completed state |
| `danger` / `--danger` | `#b42318` | Destructive actions and errors |
| `danger-soft` | `#fff1f0` | Error message background |
| `focus` / `--focus-ring` | `rgba(15, 118, 110, 0.2)` | Keyboard focus ring |

### Color Rules

- Primary actions use teal with white text.
- Destructive actions use red soft background and red text unless the action is the final destructive confirmation.
- Status cannot rely on color alone; combine color with text and, where useful, an icon.
- Code and JSON blocks use warm off-white surfaces, monospace type, and clear overflow behavior.
- Do not use purple/blue gradients or dark slate themes unless a specific future dark mode is designed.

## 3. Typography Rules

Current stack: `Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`.

当前字体栈保持系统 sans 与 Inter 优先，不引入额外字体依赖。

| Role | Size | Weight | Line height | Use |
| --- | --- | --- | --- | --- |
| Public hero title | `clamp(2.4rem, 4.5vw, 5rem)` | 800 | 1.02 | Public homepage only |
| Page title | `clamp(2.2rem, 4vw, 4rem)` | 800 | 1.05 | Public article/page title |
| Studio page title | `clamp(1.65rem, 2vw, 2.2rem)` | 800 | 1.15 | Studio topbar |
| Panel title | `1.15rem` to `1.35rem` | 800 | 1.25 | Studio panel/card headings |
| Body | `16px` | 400 | 1.55 to 1.75 | Public and general UI copy |
| Studio dense body | `14px` to `15px` | 400 | 1.45 to 1.6 | Tables, cards, metadata |
| Button | `14px` to `16px` | 800 | 1.2 | Primary and secondary commands |
| Eyebrow | `12px` to `13px` | 800 to 900 | 1.3 | Section label, uppercase allowed |
| Code | `13px` | 400 to 600 | 1.45 | JSON, API fields, config |

### Typography Rules

- Letter spacing must be `0`, including uppercase eyebrows.
- Do not scale font sizes directly with viewport width except the existing public hero/page title clamps.
- Chinese UI labels should be concise and action-oriented.
- English is allowed for API fields, status enums, technical nouns, and code-like values.
- Logs are always English; UI text is primarily Chinese unless the value is a technical protocol, endpoint, provider, or enum.

## 4. Component Styling

### Buttons

Primary buttons:

- Background `primary`, text white, 8px radius.
- Use lucide icons for clear actions such as add, save, edit, delete, health check, logout.
- Minimum height: 44px for text buttons, 40px for compact icon buttons.
- Disable state uses opacity and `cursor: not-allowed`.

Secondary buttons:

- Surface background, line border, ink text.
- Use for edit, cancel, filter, neutral navigation, and non-primary operations.

Danger buttons:

- Soft danger background, danger border/text.
- Use only for delete, revoke, disable when destructive or security-sensitive.

Icon buttons:

- Use a familiar lucide icon when the icon alone is sufficient.
- Provide `aria-label` when text is not visible.

### Navigation

Site header:

- Sticky, warm background, compact height around 64px.
- Brand is simple: icon plus `Beehive Blog`.
- Public nav should stay small and avoid marketing navigation sprawl.

Studio sidebar:

- Fixed-width workbench rail around 220px on desktop.
- Active nav uses `primary-soft` background and `primary-strong` text.
- Use icon plus Chinese label.
- On narrow screens, collapse behind a clear navigation toggle.

### Panels and Cards

Use panels for independent tool areas, not as decorative page sections.

- Background `surface`.
- 1px `line` border.
- 8px radius.
- Soft shadow only where it improves separation.
- Avoid cards inside cards. If nested content is required, use bordered rows, lists, or tables.

Public post cards:

- Use readable text rhythm and tag rows.
- Keep summaries visible without dense operational metadata.
- Do not make cards taller through decorative graphics.

Studio panels:

- Use compact headers with optional right-aligned action icon.
- Prefer tables, rows, and grouped controls over large marketing cards.
- Keep repeated items consistent in height, spacing, and action placement.

### Storage Page

The storage page is operational and should be easy to scan.

- Left/main area: storage mount list or table.
- Right/support area: driver templates and capability summaries.
- Mount item hierarchy: name, path, status/default, driver, updated/check time, config preview, actions.
- Remote config errors must be visible near the affected mount, not hidden in a toast only.
- Health check, enable/disable, set default, edit, and delete should have consistent positions.
- JSON config blocks must be scrollable and visually secondary.
- Driver capability JSON should be compact; do not let it dominate the page.

### Status Badges

Use short status labels:

| State | Label examples | Style |
| --- | --- | --- |
| Healthy | `正常`, `active`, `Ready` | `primary-soft` or success styling |
| Pending | `待处理`, `Pending`, `未知` | warning soft background |
| Disabled | `已禁用` | neutral/warning style |
| Error | `异常`, `失败` | danger soft background |
| Default | `默认` | neutral pill or honey outline |

Status text must remain readable in screenshots and grayscale.

### Forms

- Labels above inputs.
- Inputs minimum height 44px.
- Helper text uses muted color.
- Required validation errors appear near the form or affected field.
- Sensitive settings such as SMTP password or OAuth secret must clearly communicate keep/set/clear behavior.
- Test actions must use persisted settings when the workflow says so.

### Selects and Menus

- Use custom select only when current Studio pattern requires it.
- Keep popovers above sticky headers and modals via correct z-index.
- Keyboard focus must be visible.

### Tables and Dense Lists

- Use tables for users, storage lists, and any repeated operational records when comparison matters.
- Header text can be uppercase but must keep letter spacing at 0.
- Row actions should be grouped at the right.
- Avoid hidden horizontal overflow on mobile; convert to stacked rows or scrollable containers.

### Modals

- Overlay must sit above the sticky site header.
- Dialog width should fit content: 520px for simple confirmation, 760px for forms.
- Keep modal actions at the bottom, cancel first, primary/destructive last.
- Close icon requires an accessible label.

### Messages and Errors

- Success messages use soft teal/success.
- Error messages use soft red and clear Chinese text.
- Do not expose backend stack traces or unmodeled internal fields.
- Technical error fields may appear in code blocks only when intended for admins.

## 5. Layout Principles

### Public Web

- Use `main.page` with max content width from `--page-max-width`.
- Public content must be SSR/SEO friendly; important article/list body content cannot depend on post-hydration client fetching only.
- Reading line length should stay comfortable, especially article pages around 720-820px.
- Avoid landing-page filler. First screen should expose the product/content directly.

### Studio

- `StudioLayout` is a two-column grid: sidebar plus main content.
- Main content uses `container-type: inline-size` and container queries for internal grids.
- Whole-page structural shifts use media queries.
- Do not center Studio into a narrow island on desktop; it should use available width.
- Keep page topbar, actions, filters, and messages in a predictable vertical order.

### Spacing Scale

| Token | Value | Use |
| --- | --- | --- |
| `xxs` | 4px | Internal icon/text gaps |
| `xs` | 8px | Small control padding, tight lists |
| `sm` | 12px | Sidebar padding, row gaps |
| `md` | 16px | Default grid gap |
| `lg` | 24px | Panel padding, modal padding |
| `xl` | 32px | Public section breathing room |
| `section` | 48px to 72px | Public page section spacing |

### Radius and Elevation

- Default radius is 8px.
- Pills are allowed only for tags, badges, and avatar-like controls.
- Do not increase cards to large rounded rectangles unless the entire design system changes.
- Use shadows sparingly; line borders should do most structure work.

## 6. Responsive Behavior

Design for these checkpoints:

| Viewport | Expected behavior |
| --- | --- |
| 1280px+ | Studio uses sidebar + main area; storage page can use main/list and support column |
| 1024px | Studio still feels dense; grids reduce columns by container width |
| 900px | Studio support columns collapse to single column; sidebar may become non-sticky |
| 600px | Sidebar nav collapses behind toggle; topbar actions wrap below title |
| 560px | Forms become one column; segmented controls stretch; buttons remain readable |

Rules:

- Text must not overflow buttons, cards, nav items, or table cells.
- Buttons may wrap to a new line, but must not compress icon and label into overlap.
- Minimum interactive target: 40px, preferably 44px for form and command controls.
- JSON/code blocks scroll internally when long.
- Modal width must be `calc(100vw - 32px)` or equivalent on mobile.

## 7. Security and Trust UX

- Studio is admin-only and must never appear in public sitemap or SEO metadata.
- Browser JavaScript must not persist refresh tokens.
- Auth, OAuth, storage credentials, SMTP settings, and GitHub secrets should visually communicate server-side handling.
- Destructive operations require clear confirmation when they remove users, storage mounts, credentials, or sessions.
- Disabled storage mounts block new writes; UI copy must not imply existing files are deleted.
- Health checks should distinguish "not checked", "healthy", "error", and "disabled".

## 8. Do / Don't

### Do

- Use `lucide-react` icons for concrete commands.
- Keep Studio compact, predictable, and easy to scan.
- Use Chinese labels for business actions.
- Keep API names, mount paths, provider names, and JSON keys in monospace/code styling when useful.
- Prefer tables or row lists for operational data.
- Keep Public and Studio route/SEO behavior separate.
- Reuse `ui/app/globals.css` and `ui/components/studio/Studio.module.css` variables.

### Don't

- Do not add a UI framework just for cards, grids, or controls.
- Do not mix Tailwind into Studio while the app uses CSS Modules.
- Do not use decorative blobs, one-off SVG hero art, or abstract honeycomb visuals in admin surfaces.
- Do not make nested card stacks for storage, settings, users, or permissions.
- Do not use color-only status.
- Do not place page sections inside floating cards.
- Do not use viewport-width font scaling for compact UI.
- Do not leak internal stack traces, secrets, refresh tokens, or unmasked credentials.

## 9. Agent Prompt Guide

When an AI agent changes Beehive-Blog UI, use this prompt:

> Build the UI according to root `DESIGN.md`. Keep Beehive Blog as a warm paper, teal-accented knowledge workbench. Public pages are semantic, SSR/SEO-first, readable, and content-led. Studio is admin-only, dense, scannable, and operational, using CSS Modules, 8px radius, lucide-react icons, clear status labels, and predictable row/table layouts. Do not add a UI framework, decorative gradient blobs, nested card stacks, or browser-readable refresh-token storage. Reuse `ui/app/globals.css` and `ui/components/studio/Studio.module.css` tokens.

后续修改 UI 时，先核对：

- 是否符合 Public Web / Studio 双产品面边界。
- 是否继续使用 CSS Modules 与现有变量。
- 是否保持 8px 圆角、清晰状态、可访问按钮、响应式不溢出。
- 是否避免把 Studio 做成营销页面。
- 是否没有削弱 BFF Cookie、HttpOnly、admin-only、SEO 分离等安全边界。

