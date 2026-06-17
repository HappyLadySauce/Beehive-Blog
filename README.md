# Beehive Blog

个人博客、AI 协作创作空间，以及面向智能体的个人知识中台。后端提供 Go REST API，前端分为两个产品面：

- **Public Web** — SEO 优先的读者站（文章、笔记、项目等）
- **Studio** — 管理员工作台（创作、审阅、发布、附件、用户与设置）

## 技术栈

| 层 | 技术 |
| --- | --- |
| 后端 API | Go 1.26、Gin、GORM、PostgreSQL、Redis、JWT、Swagger |
| 前端 | Next.js 16（App Router）、React 19、TypeScript、Vitest、Playwright |
| 基础设施 | Docker Compose（Postgres 16 + Redis 8） |

## 仓库结构

```
Beehive-Blog/
├── cmd/          # Go 应用入口与 HTTP 路由
├── pkg/          # 共享业务与配置包
├── sql/          # 迁移 CLI 与 migrations/
├── ui/           # Next.js 前端（Public + Studio + BFF）
├── configs/      # 本地配置（gitignore，从 config-example 复制）
├── docker/       # 本地依赖编排
└── docs/         # 产品与架构文档
```

## 前置要求

- Go >= 1.26
- Node.js >= 22、[pnpm](https://pnpm.io/)
- Docker（推荐，用于 PostgreSQL / Redis）

## 快速启动

### 1. 启动依赖

在仓库根目录执行：

```bash
docker compose -f docker/Infrastructure/docker-compose.yaml up -d
```

默认暴露 `127.0.0.1:5432`（PostgreSQL）与 `127.0.0.1:6379`（Redis），数据库用户/库名均为 `Beehive-Blog`。

### 2. 准备配置

`configs/` 目录已被 git 忽略，需从模板复制本地配置：

```bash
cp configs/config-example.yaml configs/config.yaml
```

Windows PowerShell：

```powershell
Copy-Item configs/config-example.yaml configs/config.yaml
```

模板默认值已与 docker-compose 对齐。生产环境请通过环境变量注入 JWT secret 等敏感项（`BEEHIVE_BLOG_` 前缀，详见 [CLAUDE.md](CLAUDE.md)）。

### 3. 数据库迁移

Windows PowerShell：

```powershell
.\sql\migrate.ps1
```

Unix / macOS：

```bash
./sql/migrate.sh
```

默认 DSN：`postgres://Beehive-Blog:Beehive-Blog@127.0.0.1:5432/Beehive-Blog?sslmode=disable`。迁移运维说明见 [sql/migrate/README.md](sql/migrate/README.md)。

### 4. 启动后端

默认监听 `http://127.0.0.1:8080`：

```bash
go run ./cmd/ --config configs/config.yaml
```

- 存活探针：`GET /livez`
- 就绪探针：`GET /readyz`
- API 文档：`GET /swagger/index.html`

### 5. 启动前端

默认监听 `http://localhost:3000`：

```bash
cd ui
pnpm install
pnpm dev
```

Next.js 通过 rewrite 将 `/api/v1/*` 代理到 Go API（`BEEHIVE_API_BASE_URL`，默认 `http://localhost:8080`）。

### 6. 登录 Studio

全新数据库迁移后会种子一名默认管理员（仅首次安装）：

| 字段 | 值 |
| --- | --- |
| 用户名 | `admin` |
| 密码 | `Admin@123` |

首次登录后请在 Studio 中修改密码。

访问 `http://localhost:3000/studio` 进入工作台。

## 开发命令

### 后端

```bash
go build -o beehive-blog ./cmd/
go test ./...
```

更多命令、配置加载与 SQL 模式说明见 [CLAUDE.md](CLAUDE.md)。

### 前端

```bash
cd ui
pnpm dev          # 开发服务器
pnpm test         # 单元测试（Vitest）
pnpm test:e2e     # E2E（Playwright，需依赖服务已启动）
pnpm lint
pnpm build
```

E2E 前置条件与环境变量详见 [ui/README.md](ui/README.md)。

## 配置说明

配置来源优先级（高 → 低）：**命令行标志 > 环境变量 > 配置文件**。

后端关键配置块（`configs/config.yaml`）：

| 块 | 说明 |
| --- | --- |
| `insecure` | 监听地址、端口、可信反向代理 |
| `database` | PostgreSQL 连接与连接池 |
| `cache` | Redis |
| `jwt` | 签发与校验参数 |
| `email` | SMTP 出站（可被数据库设置覆盖） |
| `github-oauth2` | GitHub 登录 |

前端关键环境变量：

| 变量 | 说明 |
| --- | --- |
| `BEEHIVE_API_BASE_URL` | 服务端访问 Go API 的地址（BFF / SSR），默认 `http://localhost:8080` |
| `NEXT_PUBLIC_API_BASE_URL` | 浏览器 API 基址，默认 `/api/v1` |
| `NEXT_PUBLIC_SITE_URL` | 站点 canonical URL，默认 `http://localhost:3000` |

完整列表见 [ui/README.md](ui/README.md)。

## API 与架构

- REST 前缀：`/api/v1`
- 响应信封：`{ "code": number, "message": string, "data": ... }`
- 主要资源域：`auth`、`users`、`contents`、`tags`、`attachments`、`storage-mounts`、`settings`、`comments`

```mermaid
flowchart TB
  browser[Browser]
  next[Next.js_ui]
  bff[BFF_Routes]
  gin[Gin_API]
  pg[(PostgreSQL)]
  redis[(Redis)]
  browser --> next
  next --> bff
  bff --> gin
  gin --> pg
  gin --> redis
```

Public 页面走 SSR/SEO 策略，Studio 为 Client-heavy 工作台；BFF 负责 Cookie 会话与同源代理，详见 [docs/frontend/react-ssr-seo-architecture.md](docs/frontend/react-ssr-seo-architecture.md)。

## 文档索引

| 文档 | 用途 |
| --- | --- |
| [docs/product-principles.md](docs/product-principles.md) | 产品边界与权限口径 |
| [docs/frontend/react-ssr-seo-architecture.md](docs/frontend/react-ssr-seo-architecture.md) | Public / Studio / BFF / SEO |
| [docs/v1/](docs/v1/) | 登录、附件、存储驱动等 v1 规则 |
| [CLAUDE.md](CLAUDE.md) | 工程命令、迁移、SQL 模式 |
| [DESIGN.md](DESIGN.md) | UI 设计契约 |
| [ui/README.md](ui/README.md) | 前端命令与 E2E |
| [sql/migrate/README.md](sql/migrate/README.md) | 迁移运维说明 |

## 许可证

许可证待定。
