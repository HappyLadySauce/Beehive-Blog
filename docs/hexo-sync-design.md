# Hexo 数据同步方案设计文档（方案 B：Build-time SSG）

> 本文档定义了 Beehive-Blog 项目中"数据库内容 → Hexo 静态文件"的完整同步方案，用于指导后续开发实现。

---

## 1. 方案概述

### 1.1 核心思路

将 PostgreSQL 数据库中的文章/分类/标签数据，通过 Go 同步服务转换为 Hexo 兼容的 Markdown 文件，写入 `ui/hexo/source/_posts/` 目录，再由 `hexo generate` 生成纯静态 HTML。

```
PostgreSQL ──查询──→ Go Sync Service ──转换──→ _posts/*.md ──构建──→ public/(静态HTML)
```

### 1.2 选择理由

| 维度 | 说明 |
|------|------|
| **SEO** | 纯 HTML 输出，搜索引擎友好 |
| **性能** | CDN 可缓存，访问速度快 |
| **兼容性** | 复用现有 Hexo 主题模板体系，改动最小 |
| **稳定性** | 静态站点不依赖运行时数据库连接 |

### 1.3 适用场景

- 以内容展示为主的个人博客
- SEO 是重要指标
- 文章更新频率中等（非实时性要求）
- 评论/点赞等互动功能通过 API 单独处理

---

## 2. 数据映射关系

### 2.1 文章字段映射

数据库 `articles` 表 → Hexo Front-matter 的字段对应关系：

| 数据库字段 | 类型 | Hexo Front-matter | 必填 | 说明 |
|-----------|------|-------------------|:----:|------|
| `id` | bigint | `beehive_id` | ✅ | 回溯键：静态页面通过此 ID 回调 API 获取评论/点赞 |
| `title` | varchar(200) | `title` | ✅ | 文章标题 |
| `slug` | varchar(100) | 文件名 + URL 路径 | ✅ | URL 别名，如 `my-first-post` |
| `content` | text | 正文（`---` 之后） | ✅ | Markdown 原文直接写入 |
| `summary` | varchar(500) | `description` | ❌ | 文章摘要/描述 |
| `cover_image` | varchar(500) | `cover` | ❌ | 封面图 URL |
| `status` | varchar(20) | 控制是否生成文件 | ✅ | 仅 `published` 状态生成文件 |
| `password` | varchar(100) | `password` | ❌ | 密码保护（空字符串表示无保护） |
| `is_pinned` | boolean | `pin` | ❌ | 是否置顶 |
| `pin_order` | int | `pin_order` | ❌ | 置顶排序权重 |
| `view_count` | bigint | `views` | ❌ | 浏览量（构建时快照值） |
| `published_at` | timestamp | `date` | ✅ | 发布时间 |
| `updated_at` | timestamp | `updated` | ❌ | 更新时间 |
| `category_id` → `categories.name` | 关联 | `categories: [名称]` | ❌ | 分类名称数组 |
| `tags` → `tags[].name` | 多对多关联 | `tags: [标签1, 标签2]` | ❌ | 标签名称数组 |

### 2.2 生成的 Markdown 文件示例

数据库中一篇已发布文章转换后的结果：

```markdown
---
title: "我的第一篇文章"
description: "这是文章摘要内容"
date: 2026-04-02T10:30:00+08:00
updated: 2026-04-02T15:00:00+08:00
categories:
  - 建站记录
tags:
  - Hexo
  - Beehive
cover: /uploads/images/cover.jpg
password: ""
pin: false
pin_order: 0
views: 128
beehive_id: 42
---

这里是文章的 **Markdown 正文** 内容...

## 二级标题

正文继续...
```

### 2.3 分类与标签的处理

分类和标签在数据库中是独立表，通过关联表连接到文章。同步时需要联表查询：

**分类（Category）**：
- 一篇文章属于一个分类（`category_id`）
- 同步时将 `category.name` 写入 Front-matter 的 `categories` 数组
- 分类页面（`/categories/:slug/`）由 Hexo 自动根据 front-matter 生成

**标签（Tag）**：
- 一篇文章可有多个标签（通过 `article_tags` 多对多关联）
- 同步时将所有 `tag.name` 写入 Front-matter 的 `tags` 数组
- 标签页面（`/tags/:name/`）由 Hexo 自动根据 front-matter 生成

---

## 3. 文件命名规范

### 3.1 命名规则

```
beehive-{database_id}-{slug}.md
```

### 3.2 命名示例

| 数据库 ID | Slug | 生成文件名 |
|----------|------|-----------|
| 42 | `my-first-post` | `beehive-42-my-first-post.md` |
| 10 | `hello-world` | `beehive-10-hello-world.md` |
| 7 | `hexo-theme-guide` | `beehive-7-hexo-theme-guide.md` |
| 15 | `` （空 slug） | `beehive-15-post-15.md` |

### 3.3 设计理由

1. **唯一性保证**：即使两篇文章 slug 相同，因 DB ID 不同也不会冲突
2. **可追溯性**：从文件名直接看到数据库主键 ID
3. **隔离性**：`beehive-` 前缀明确标识该文件由同步服务管理，与手动创建的原生 Hexo 文章区分
4. **排序友好**：按文件名字符串排序 ≈ 按创建时间排序
5. **冲突安全**：手动在 `_posts/` 中创建的非 `beehive-` 前缀文件不会被同步逻辑触碰

---

## 4. URL 路由配置

### 4.1 当前配置

当前 [ui/hexo/_config.yml](../ui/hexo/_config.yml) 使用日期路径：

```yaml
permalink: :year/:month/:day/:title/
```

生成的 URL 示例：`/2026/04/02/my-first-post/`

### 4.2 目标配置

需修改为 slug 驱动路径，与需求文档中 `/archives/{slug}` 路由一致：

```yaml
permalink: archives/:slug/
```

生成的 URL 示例：`/archives/my-first-post/`

### 4.3 影响范围

修改 permalink 后的影响：

| 组件 | 是否受影响 | 说明 |
|------|:---------:|------|
| `url_for(post.path)` | ❌ 不影响 | Hexo 内部自动适配新 permalink 规则 |
| `url_for(tag.path)` | ❌ 不影响 | 标签路径不受 permalink 影响 |
| `url_for(cat.path)` | ❌ 不影响 | 分类路径不受 permalink 影响 |
| 已有外部链接 | ⚠️ 受影响 | 旧日期 URL 将失效，需 301 重定向或接受断裂 |
| 搜索引擎索引 | ⚠️ 受影响 | 建议 Google Search Console 提交新 sitemap |

---

## 5. 系统架构设计

### 5.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                      触发层（3种方式）                           │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐   │
│  │ 管理后台手动 │  │ CRUD 自动触发 │  │ 定时任务补偿       │   │
│  │ /admin/sync  │  │ 写DB后同步    │  │ Cron 每60分钟      │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬───────────┘   │
│         └────────────────┼───────────────────┘                │
│                          ▼                                     │
│              ┌───────────────────────┐                         │
│              │    Sync Service       │                         │
│              │    (Go 同步服务)       │                         │
│              └───────────┬───────────┘                         │
└──────────────────────────┼────────────────────────────────────┘
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                      转换层                                      │
│                                                                  │
│  PostgreSQL                                                      │
│    ├── articles (JOIN category, tags)                            │
│    ▼                                                             │
│  FrontMatter Converter                                           │
│    ├── 构建 YAML front-matter                                    │
│    ├── 拼接 Markdown 正文                                        │
│    ▼                                                             │
│  ui/hexo/source/_posts/                                         │
│    ├── beehive-42-my-first-post.md                               │
│    ├── beehive-10-hello-world.md                                 │
│    └── beehive-7-hexo-theme-guide.md                             │
│                                                                  │
└──────────────────────────┬───────────────────────────────────────┘
                           ▼
              ┌───────────────────────┐
              │   hexo generate       │
              │   (构建静态站点)       │
              └───────────┬───────────┘
                          ▼
              ┌───────────────────────┐
              │  ui/hexo/public/      │
              │  (纯静态 HTML)         │
              └───────────────────────┘
```

### 5.2 新增代码模块结构

```
cmd/app/
├── sync/                              # ← 新增目录：同步核心模块
│   ├── service.go                     #    同步服务（全量/单篇/清理）
│   ├── converter.go                   #    DB模型 → Hexo Markdown 转换器
│   ├── types.go                       #    同步相关 DTO 与常量定义
│   └── converter_test.go              #    转换单元测试
│
├── routes/admin/
│   └── sync.go                        # ← 新增：管理员同步接口 Handler
│
├── middlewares/
│   └── (无新增)
│
└── types/api/v1/
    └── sync.go                        # ← 新增：同步接口请求/响应 DTO

pkg/options/
└── hexo.go                            # ← 新增：Hexo 同步相关配置结构体

configs/
└── Beehive-Blog.yaml                  # ← 扩展：增加 hexo 配置段
```

---

## 6. 核心模块详细设计

### 6.1 配置项 (`pkg/options/hexo.go`)

```go
package options

type HexoConfig struct {
	PostsDir    string `mapstructure:"posts_dir"`    // _posts 目录路径（相对或绝对）
	AutoSync    bool   `mapstructure:"auto_sync"`    // CRUD 后是否自动同步
	HexoCommand string `mapstructure:"hexo_command"` // hexo build 命令（可选自动执行）
	WatchMode   bool   `mapstructure:"watch_mode"`   // 开发环境 watch 模式
}
```

**configs/Beehive-Blog.yaml 新增段**：

```yaml
hexo:
  posts_dir: "ui/hexo/source/_posts"
  auto_sync: true
  hexo_command: ""          # 留空表示不自动执行 hexo generate
  watch_mode: false
```

### 6.2 数据传输对象 (`cmd/app/types/api/v1/sync.go`)

```go
package v1

type SyncResponse struct {
	Total   int      `json:"total"`   // 同步文章总数
	Created int      `json:"created"` // 新建文件数
	Updated int      `json:"updated"` // 更新文件数
	Deleted int      `json:"deleted"` // 清理孤立文件数
	Files   []string `json:"files"`   // 涉及的文件名列表
}

type SyncStatusResponse struct {
	LastSyncTime  string `json:"last_sync_time"`  // 上次同步时间
	TotalPosts    int    `json:"total_posts"`     // 数据库已发布文章总数
	LocalFiles    int    `json:"local_files"`     // _posts 中 beehive-* 文件数
	PendingSync   bool   `json:"pending_sync"`    // 是否有待同步变更
}
```

### 6.3 同步类型常量 (`cmd/app/sync/types.go`)

```go
package sync

type SyncAction string

const (
	SyncActionCreate SyncAction = "created"
	SyncActionUpdate SyncAction = "updated"
	SyncActionDelete SyncAction = "deleted"
)

type SyncResult struct {
	Total   int
	Created int
	Updated int
	Deleted int
	Files   []string
}
```

### 6.4 转换器 (`cmd/app/sync/converter.go`)

**职责**：将单个 `Article` Go 对象（含关联的 Category、Tags）转换为符合 Hexo 格式的 `.md` 文件字节内容。

**核心函数签名**：

```go
func ArticleToHexoMarkdown(
    article *models.Article,
    categoryName string,
    tagNames []string,
) ([]byte, error)
```

**转换流程**：

```
Article 对象
    │
    ├─→ 提取字段 → 构建 HexoFrontMatter struct
    │               │
    │               └─→ yaml.Marshal() → YAML 字节
    │
    ├─→ 拼接格式：
    │       "---\n"
    │       + YAML 字节
    │       "---\n"
    │       + article.Content (Markdown 正文)
    │
    └─→ 返回完整 []byte
```

**HexoFrontMatter 结构体定义**：

```go
type HexoFrontMatter struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Date        time.Time `yaml:"date"`
	Updated     time.Time `yaml:"updated,omitempty"`
	Categories  []string  `yaml:"categories,omitempty"`
	Tags        []string  `yaml:"tags,omitempty"`
	Cover       string    `yaml:"cover,omitempty"`
	Password    string    `yaml:"password,omitempty"`
	IsPinned    bool      `yaml:"pin,omitempty"`
	PinOrder    int       `yaml:"pin_order,omitempty"`
	Views       int64     `yaml:"views,omitempty"`
	BeehiveID   int64     `yaml:"beehive_id"`
}
```

**辅助函数**：

```go
func GenerateHexoFileName(article *models.Article) string
// 生成文件名：beehive-{id}-{slug}.md
// 若 slug 为空则使用 post-{id} 作为 fallback
```

### 6.5 同步服务 (`cmd/app/sync/service.go`)

**职责**：管理同步生命周期——查询数据库、调用转换器、写文件、清理孤立文件。

**核心方法**：

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewSyncService` | `(postsDir string, db *gorm.DB) *SyncService` | 构造函数 |
| `SyncAll` | `(ctx context.Context) (*SyncResult, error)` | 全量同步：发布文章 → `.md` 文件 |
| `SyncSingle` | `(ctx context.Context, articleID int64) error` | 单篇同步：增量更新一篇文章 |
| `DeletePostFile` | `(article *models.Article) error` | 删除指定文章对应的 .md 文件 |
| `CleanupOrphaned` | `(activeArticles []models.Article) (int, error)` | 清理孤立文件：DB 中不存在但 _posts 中还存在的 beehive-* 文件 |

**SyncAll 流程**：

```
1. 查询 DB：SELECT * FROM articles WHERE status='published' AND deleted_at IS NULL
   Preload Category, Preload Tags, ORDER BY published_at DESC
        │
2. 遍历每篇文章：
   ├─→ converter.ArticleToHexoMarkdown() → 得到 []byte
   ├─→ 判断文件是否存在（os.Stat）
   │   ├─ 存在 → 覆盖写入 → action = updated
   │   └─ 不存在 → 新建写入 → action = created
   └─→ 记录到 SyncResult
        │
3. CleanupOrphaned()：扫描 _posts 中所有 beehive-*.md
   ├─ 解析文件名中的 ID
   ├─ 检查该 ID 在当前 activeArticles 中是否存在
   └─ 不存在 → 删除文件（DB 已删但文件残留）
        │
4. 返回 SyncResult
```

**SyncSingle 流程**：

```
1. 根据 articleID 查询单篇文章（Preload Category + Tags）
        │
2. 判断 status：
   ├─ published → 调用 converter → 写入/覆盖文件
   └─ draft/archived/private/deleted → DeletePostFile()
```

### 6.6 同步 API 接口 (`cmd/app/routes/admin/sync.go`)

**接口定义**：

#### POST /api/v1/admin/sync/posts — 手动触发全量同步

- **权限**：管理员（`RequireRoles(admin)`）
- **请求体**：空（或可选 `{ "rebuild": true }` 表示同步后同时执行 hexo build）
- **响应**：`SyncResponse`

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 15,
    "created": 2,
    "updated": 13,
    "deleted": 1,
    "files": [
      "beehive-42-my-first-post.md",
      "beehive-10-hello-world.md"
    ]
  }
}
```

#### GET /api/v1/admin/sync/status — 查询同步状态

- **权限**：管理员
- **响应**：`SyncStatusResponse`

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "last_sync_time": "2026-04-02T15:30:00+08:00",
    "total_posts": 15,
    "local_files": 14,
    "pending_sync": true
  }
}
```

### 6.7 路由注册

在 [cmd/app/router/router.go](../cmd/app/router/router.go) 的 admin 分组下追加：

```go
adminGroup := r.Group("/admin")
adminGroup.Use(middlewares.Auth(svcCtx))
adminGroup.Use(middlewares.RequireRoles(models.UserRoleAdmin))
{
    adminGroup.GET("/ping", admin.Ping(svcCtx))

    // ====== 同步接口（新增）======
    adminGroup.POST("/sync/posts", admin.SyncPosts(svcCtx))
    adminGroup.GET("/sync/status", admin.SyncStatus(svcCtx))
}
```

---

## 7. 触发机制详解

### 7.1 三级触发策略

```
Level 1: 实时触发（CRUD 后立即同步）
    ↓ 失败降级
Level 2: 手动触发（管理员调用 POST /admin/sync/posts）
    ↓ 兜底保障
Level 3: 定时补偿（Cron 全量同步，每小时一次）
```

### 7.2 Level 1：CRUD 自动同步

在每个文章管理的 Handler 中，数据库操作成功后调用 `SyncSingle`：

**创建文章** (`POST /articles`)：

```go
func CreateArticle(svcCtx *svc.ServiceContext) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req types.CreateArticleRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            response.Fail(c, http.StatusBadRequest, "参数错误")
            return
        }

        article := &models.Article{...}
        if err := svcCtx.DB.Create(article).Error; err != nil {
            response.Fail(c, 500, "创建失败")
            return
        }

        if svcCtx.Config.Hexo.AutoSync {
            syncSvc := sync.NewSyncService(svcCtx.Config.Hexo.PostsDir, svcCtx.DB)
            go func() {
                if err := syncSvc.SyncSingle(context.Background(), article.ID); err != nil {
                    log.Printf("[SYNC] 文章 %d 同步失败: %v", article.ID, err)
                }
            }()
        }

        response.Success(c, articleToDTO(article))
    }
}
```

**更新文章** (`PUT /articles/:id`)：同理，更新后调用 `SyncSingle(articleID)`

**删除文章** (`DELETE /articles/:id`)：软删除后调用 `DeletePostFile(article)`

**修改状态** (`PUT /articles/:id/status`)：若改为 `published` 则 `SyncSingle`，若从 `published` 改为其他状态则 `DeletePostFile`

> **注意**：Level 1 使用 goroutine 异步执行，不阻塞 API 响应。

### 7.3 Level 2：手动全量同步

管理员通过接口手动触发，适用于：
- 首次初始化同步
- 数据修复后重新同步
- Level 1 同步失败后的补救

### 7.4 Level 3：定时任务兜底

在应用启动时注册一个定时任务：

```go
func StartSyncScheduler(syncSvc *sync.SyncService) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for range ticker.C {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
            result, err := syncSvc.SyncAll(ctx)
            cancel()

            if err != nil {
                log.Printf("[SYNC-SCHEDULER] 全量同步失败: %v", err)
            } else {
                log.Printf("[SYNC-SCHEDULER] 完成: +%d ~%d -%d (共%d篇)",
                    result.Created, result.Updated, result.Deleted, result.Total)
            }
        }
    }()
}
```

在 `cmd/app/app.go` 的启动流程中调用：

```go
if config.Hexo.AutoSync {
    syncSvc := sync.NewSyncService(config.Hexo.PostsDir, db)
    StartSyncScheduler(syncSvc)
}
```

---

## 8. 冲突处理与边界情况

### 8.1 冲突场景矩阵

| 场景 | 处理策略 | 说明 |
|------|----------|------|
| 数据库删除文章（软删除） | 删除对应 `.md` 文件 | `CleanupOrphaned` 或 `DeletePostFile` |
| 文章状态从 published 改为 draft | 删除 `.md` 文件 | 草稿不应出现在静态站点 |
| 文章 slug 变更 | 删除旧文件，生成新文件 | 文件名含 slug |
| `_posts` 中有非 `beehive-` 前缀的 `.md` | 完全保留不触碰 | 手动创建的原生 Hexo 文章 |
| 手动编辑了 `beehive-*.md` 文件 | 下次同步时被覆盖 | 以数据库为唯一数据源 |
| 同步过程中断（部分写入） | 幂等重试即可 | 每次同步都是全量覆盖写 |
| 两篇文章 slug 相同 | 不冲突（文件名含 DB ID） | `beehive-1-slug.md` vs `beehive-2-slug.md` |
| 文章无 slug | Fallback 为 `post-{id}` | 保证文件名有效 |
| `_posts` 目录不存在 | 自动创建 | `os.MkdirAll(postsDir, 0755)` |
| 文章内容含特殊字符（YAML 注入） | 内容在 `---` 分隔符之后 | 不影响 front-matter 解析 |

### 8.2 并发安全

- 文件写入使用 `os.WriteFile`（原子操作：写临时文件 + rename）
- 多个 `SyncSingle` 并发执行时，每个操作独立写不同文件，天然无竞争
- `SyncAll` 与 `SyncSingle` 可能并发：对同一文件的并发写入由 OS 文件锁保证最终一致性（最后一次写入胜出，且内容来自同一 DB 状态）

---

## 9. Hexo 主题适配

### 9.1 需要适配的自定义 Front-matter 字段

同步引入了以下标准 Hexo 不包含的自定义字段，主题模板需要正确读取：

| 自定义字段 | 使用位置 | 用途 |
|-----------|---------|------|
| `beehive_id` | post.ejs 底部评论区 | 作为评论 API 的 `articleId` 参数 |
| `pin` | index.ejs 排序、post-card.ejs | 显示置顶标记、置顶排序 |
| `pin_order` | index.ejs 排序 | 多篇置顶时的内部排序 |
| `views` | post-card.ejs | 显示浏览量数字 |
| `cover` | post-card.ejs | 封面图展示（已有支持） |
| `password` | post.ejs | 密码保护校验（前端拦截） |

### 9.2 index.ejs 改造要点

**文件**: [ui/hexo/themes/happyladysacue/layout/index.ejs](../ui/hexo/themes/happyladysacue/layout/index.ejs)

当前首页按 `date` 降序排列。需要改为：**置顶优先 → 然后按日期降序**：

```ejs
<%
const posts = (page.posts && page.posts.toArray ? page.posts.toArray() : [])
  .sort(function(a, b) {
    if (a.pin && !b.pin) return -1;
    if (!a.pin && b.pin) return 1;
    if (a.pin && b.pin) return (b.pin_order || 0) - (a.pin_order || 0);
    return b.date - a.date;
  });
%>
```

### 9.3 post-card.ejs 改造要点

**文件**: [ui/hexo/themes/happyladysacue/layout/partials/post-card.ejs](../ui/hexo/themes/happyladysacue/layout/partials/post-card.ejs)

增加置顶标记：

```ejs
<%# 在 <article class="post-card"> 内部顶部追加 %>
<% if (currentPost.pin) { %>
  <span class="post-card__pinned-badge" title="已置顶">
    <svg>...</svg> 置顶
  </span>
<% } %>
```

现有的 `cover` 和 `views` 字段读取逻辑无需改动，已兼容。

### 9.4 post.ejs 改造要点

**文件**: [ui/hexo/themes/happyladysacue/layout/post.ejs](../ui/hexo/themes/happyladysacue/layout/post.ejs)

在文章底部追加评论区动态加载容器：

```ejs
<%# 文章正文结束后、footer 前插入 %>
<div id="beehive-comments-root"
     data-beehive-id="<%= page.beehive_id %>"
     data-api-base="<%= theme.behive.api_base %>">
</div>
```

密码保护文章的前端拦截逻辑（伪代码）：

```javascript
if (page.password && page.password !== '') {
  showPasswordModal(page.password);
}
```

### 9.5 sidebar.ejs 改造要点

**文件**: [ui/hexo/themes/happyladysacue/layout/partials/sidebar.ejs](../ui/hexo/themes/happyladysacue/layout/partials/sidebar.ejs)

侧边栏的分类和标签数据来源无需改变——仍然从 `site.categories` 和 `site.tags` 读取，这些数据由 Hexo 从 `_posts/*.md` 的 front-matter 中自动提取。同步服务正确写入 categories/tags 后，侧边栏会自动更新。

---

## 10. 评论/互动功能的 Hybrid 处理

### 10.1 架构定位

```
┌─────────────────────────────────────────┐
│           静态 HTML (hexo generate)      │
│                                          │
│  ✅ 文章标题、内容、分类、标签             │  ← 构建时确定（来自 .md）
│  ✅ 发布时间、摘要、封面图                 │  ← 构建时确定
│  ✅ 浏览量（构建时快照）                   │  ← 构建时确定
│                                          │
│  ⚡ 评论列表                              │  ← 运行时 fetch API
│  ⚡ 点赞/收藏状态                          │  ← 运行时 fetch API
│  ⚡ 浏览计数（+1）                         │  ← 运行时 post API
│  ⚡ 用户登录态                            │  ← 运行时 auth API
│                                          │
└─────────────────────────────────────────┘
```

### 10.2 评论区动态加载

在 `post.ejs` 中预留的 `#beehive-comments-root` 容器，由前端 JS 负责渲染：

```javascript
// themes/happyladysacue/source/js/beehive-comments.js（新建）
const BeehiveComments = {
  apiBase: window.__BEEHIVE_API_BASE__,

  async loadComments(articleId) {
    const res = await fetch(`${this.apiBase}/api/v1/articles/${articleId}/comments`);
    const json = await res.json();
    if (json.code === 200) {
      this.renderComments(json.data);
    }
  },

  renderComments(comments) {
    const root = document.getElementById('beehive-comments-root');
    // 渲染评论列表 DOM...
  },

  init() {
    const root = document.getElementById('beehive-comments-root');
    if (!root) return;
    const articleId = root.dataset.beehiveId;
    if (articleId) this.loadComments(articleId);
  }
};

document.addEventListener('DOMContentLoaded', () => BeehiveComments.init());
```

### 10.3 浏览计数

文章详情页加载时发送浏览记录：

```javascript
async function recordView(articleId) {
  try {
    await fetch(`${apiBase}/api/v1/articles/${articleId}/view`, { method: 'POST' });
  } catch (e) { /* 静默失败 */ }
}
```

> 注意：静态 HTML 中的 `views` 值是构建时的快照，实际浏览量以 API 为准。如需在前端显示实时浏览量，可额外调用一个轻量接口获取。

---

## 11. 开发实施路线

### Phase 1：基础同步能力（MVP）

**目标**：能通过 API 手动触发同步，数据库文章正确输出为 `.md` 文件

- [ ] 创建 `pkg/options/hexo.go` — Hexo 配置结构体
- [ ] 创建 `cmd/app/sync/types.go` — DTO 与常量
- [ ] 创建 `cmd/app/sync/converter.go` — Article → Markdown 转换器
- [ ] 编写 `cmd/app/sync/converter_test.go` — 单元测试
- [ ] 创建 `cmd/app/sync/service.go` — 同步服务（SyncAll + SyncSingle + CleanupOrphaned）
- [ ] 创建 `cmd/app/types/api/v1/sync.go` — 接口 DTO
- [ ] 创建 `cmd/app/routes/admin/sync.go` — Handler
- [ ] 在 `router.go` 注册同步路由
- [ ] 在 `configs/Beehive-Blog.yaml` 添加 `hexo:` 配置段
- [ ] 修改 `ui/hexo/_config.yml` 的 `permalink` 为 `archives/:slug/`
- [ ] **验证**：启动后端 → 创建测试文章 → 调用 `POST /admin/sync/posts` → 检查 `_posts/` 目录输出

### Phase 2：自动化集成

**目标**：文章 CRUD 操作后自动触发同步；可选全量/定时补偿后执行 `hexo clean` + `hexo generate`（见 `hexo.clean_args` / `hexo.generate_args`）。

- [x] 文章 Create Handler 中加入异步 `SyncSingle` 调用（`hexo.auto_sync`）
- [x] 文章 Update Handler 中加入异步 `SyncSingle` 调用
- [x] 文章 Delete Handler 中加入 `DeletePostFile` 调用
- [x] 文章 Status Change Handler 中判断同步/删除
- [x] 手动全量 `POST .../sync/posts` 且 `rebuild: true` 时 `RunHexoRebuild`（clean 后 generate）
- [x] 可选 `rebuild_after_auto_sync`：单篇自动同步后再重建静态站
- [ ] **验证**：通过管理 API 创建/编辑/删除文章 → 自动检查 `_posts/` 变化

### Phase 3：主题适配

**目标**：Hexo 主题正确渲染同步生成的自定义字段

- [ ] 修改 `index.ejs` — 置顶排序逻辑
- [ ] 修改 `post-card.ejs` — 置顶标记
- [ ] 修改 `post.ejs` — 评论区容器 + beehive_id 传递
- [ ] 新建 `beehive-comments.js` — 评论动态加载
- [ ] 可选：密码保护前端拦截
- [ ] **验证**：`hexo server` 本地预览 → 检查文章列表/详情/分类/标签页

### Phase 4：生产化加固

**目标**：可靠性、监控、性能优化

- [ ] 同步失败重试机制（指数退避）
- [ ] 同步日志结构化输出（含耗时、文件数量）
- [ ] `GET /admin/sync/status` 接口实现
- [ ] 可选：同步完成后自动执行 `hexo generate`（exec.Command）
- [ ] 可选：Webhook 通知部署（同步 → build → CD）
- [ ] **验证**：压力测试 + 边界情况测试

---

## 12. 注意事项与风险

### 12.1 不要做的事

- ❌ 不要在同步服务中修改数据库内容 —— 单向流动：DB → 文件
- ❌ 不要让同步阻塞 API 请求 —— 必须异步（goroutine 或队列）
- ❌ 不要删除非 `beehive-` 前缀的 `.md` 文件 —— 保护手动创建的内容
- ❌ 不要在 Front-matter 中存储敏感信息（password 字段仅存空字符串或哈希）

### 12.2 性能考量

| 操作 | 预期耗时 | 说明 |
|------|---------|------|
| SyncSingle（单篇） | < 10ms | 一次 DB 查询 + 一个文件写入 |
| SyncAll（100篇） | < 2s | 批量查询 + 顺序写文件 |
| hexo generate（100篇） | 5-30s | 取决于内容和插件复杂度 |
| CleanupOrphaned | < 100ms | 目录扫描 + ID 匹配 |

### 12.3 安全注意

- `POST /api/v1/admin/sync/posts` 接口必须限制为管理员权限
- 同步服务的 `postsDir` 路径应限定在项目目录内，防止路径遍历攻击
- 文件写入权限遵循最小权限原则

---

## 13. 附录：快速参考

### 13.1 关键文件清单

| 文件路径 | 类型 | 说明 |
|---------|------|------|
| `pkg/options/hexo.go` | 新增 | Hexo 同步配置结构体 |
| `cmd/app/sync/types.go` | 新增 | 同步 DTO 与常量 |
| `cmd/app/sync/converter.go` | 新增 | DB → Markdown 转换器 |
| `cmd/app/sync/service.go` | 新增 | 同步服务核心逻辑 |
| `cmd/app/sync/converter_test.go` | 新增 | 转换器单元测试 |
| `cmd/app/types/api/v1/sync.go` | 新增 | 接口请求/响应 DTO |
| `cmd/app/routes/admin/sync.go` | 新增 | 同步接口 Handler |
| `cmd/app/router/router.go` | 修改 | 注册同步路由 |
| `configs/Beehive-Blog.yaml` | 修改 | 添加 hexo 配置段 |
| `ui/hexo/_config.yml` | 修改 | permalink 改为 `archives/:slug/` |
| `ui/hexo/themes/happyladysacue/layout/index.ejs` | 修改 | 置顶排序 |
| `ui/hexo/themes/happyladysacue/layout/partials/post-card.ejs` | 修改 | 置顶标记 |
| `ui/hexo/themes/happyladysacue/layout/post.ejs` | 修改 | 评论区容器 |
| `ui/hexo/themes/happyladysacue/source/js/beehive-comments.js` | 新增 | 评论动态加载 |

### 13.2 API 接口速查

| 方法 | 路径 | 权限 | 说明 |
|------|------|:----:|------|
| POST | `/api/v1/admin/sync/posts` | admin | 手动触发全量同步 |
| GET | `/api/v1/admin/sync/status` | admin | 查询同步状态 |

### 13.3 数据流向速查

```
管理员操作
    │
    ▼
Go API Handler (routes/admin/article.go)
    │
    ▼
写 PostgreSQL (articles / article_tags)
    │
    ▼
SyncSingle() [异步 goroutine]
    │
    ▼
Converter: Article → YAML Front-matter + Markdown
    │
    ▼
WriteFile: ui/hexo/source/_posts/beehive-{id}-{slug}.md
    │
    ▼
hexo generate → ui/hexo/public/archives/{slug}/index.html
    │
    ▼
用户浏览器访问静态 HTML
    │
    ▼
[可选] beehive-comments.js 加载评论 (fetch API)
```
