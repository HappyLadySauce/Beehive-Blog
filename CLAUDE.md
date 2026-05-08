# CLAUDE.md 讲中文

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 开发环境与命令

- **构建**: `go build -o beehive-blog ./cmd/`
- **运行**: `go run ./cmd/ --config configs/config.yaml`
- **测试全部**: `go test ./...`
- **测试单个包**: `go test ./cmd/app/options/ -run TestInsecureServingValidateJoinsAllMissingFields`
- **数据库迁移**: `go run ./sql/migrate/ -dsn "postgres://user:pass@host:5432/db?sslmode=disable" -mode versioned`
  - 迁移入口脚本: `./sql/migrate.sh` (Unix) 或 `.\sql\migrate.ps1` (Windows PowerShell)
  - `MODE=adaptive` 可按语句执行并跳过"对象已存在"类错误；`MIGRATION_FORCE=1` 覆盖校验和不一致的记录；`MIGRATION_REAPPLY=1` 重跑已应用迁移

## 关键架构

### 启动流程

`cmd/main.go` → `cmd/app/api.go:NewAPICommand()` → cobra 命令解析标志 → Viper 加载 YAML 配置 → `opts.Validate()` → `run()` → `config.Init()` → `svc.NewServiceContext()` 打开 PG + Redis → `serve()` 启动 Gin HTTP 路由

### 配置加载（Kubernetes 风格 CLI）

配置有三个来源，优先级从高到低：**命令行标志 > 环境变量 > 配置文件**。

- `configs/config.yaml` 是默认配置文件
- `pkg/options/config.go` 实现 `--config` 标志注册和 Viper 初始化，支持 `${ENV_VAR}` 展开
- 配置文件搜索路径：当前目录 → `~/.beehive/` → `/etc/beehive/`，文件名与 basename 一致
- 环境变量前缀为 basename 的大写加下划线形式（如 `BEEHIVE_BLOG_`）

### 选项与验证

`pkg/options/` 定义三个选项组，每个都实现 `Validate()` + `AddFlags()`：

| 选项组 | 结构体 | 对应配置项 |
|--------|--------|-----------|
| 不安全监听 | `InsecureServingOptions` | `bind-address`, `bind-port` |
| PostgreSQL | `PostgreOptions` | `host`, `port`, `user`, `password`, `db`, `ssl-mode`, 连接池参数 |
| Redis | `RedisOptions` | `host`, `port`, `password`, `db` |

`cmd/app/options/options.go` 聚合所有选项组，`cmd/app/options/validation.go` 汇总校验。

### 服务上下文

`cmd/app/svc/servicecontext.go` 的 `ServiceContext` 结构体持有 `*gorm.DB`（PostgreSQL via pgx 驱动）、`*redis.Client` 和 `*config.Config`。`NewServiceContext()` 负责建立连接、应用连接池参数、Ping 连通性检查。`Close()` 按先 SQL 后 Redis 顺序释放资源。

### Gin 路由

`cmd/app/router/router.go` 使用包级变量（非 init()）保存 `*gin.Engine` 和 `/api/v1` 分组。暴露：
- `GET /livez` — 存活探针
- `GET /readyz` — 就绪探针
- `GET /api/v1/*` — REST API 分组
- `GET /swagger/*any` — Swagger UI

### 数据库迁移系统

`sql/migrate/main.go` 是独立的迁移 CLI（与主应用解耦）。两种模式：

- **versioned**（默认）：整文件事务执行，记录 SHA-256 校验和到 `schema_migrations` 表
- **adaptive**：按 `;` 拆分语句，遇到 `42P07`/`42701`/`42710` 类 SQLSTATE 时跳过（仅跳过"对象已存在"）

迁移文件按文件名前缀排序（如 `000_` < `001_`），分布在 `sql/migrations/<domain>/` 子目录中。

### SQL 模式设计

- 所有表使用 GORM 标准时间字段：`created_at`, `updated_at`, `deleted_at`（软删）
- 唯一约束通过 `WHERE deleted_at IS NULL` 部分索引实现（允许软删后复用）
- 附件系统：`attachment.attachments` 支持 s3/oss/local 三种后端，`attachment.categories` 使用物化路径树
- 用户系统：`identity.users` 头像外键指向 `attachment.attachments`，附件软删时由数据库触发器自动清空头像引用

## 代码注释规范

代码注释采用**中英双语**：英文在上方，中文在下方。例如：
```go
// Validate checks bounds and sets defaults.
// Validate 校验参数边界并设置默认值。
```

## 日志规范

全部日志输出使用**英文**，通过 `klog` 包记录（k8s.io/klog/v2）。
