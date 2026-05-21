# Beehive-Blog v1：文件管理架构

本文档定义 Studio 文件管理层的目标与边界。交互上仍参考 OpenList 式路径浏览，但**当前已落地的数据库真相源**是 `attachment.storage_mounts` + `attachment.attachments`（`storage_mount_id` + `object_key`），不再使用 `attachment.file_nodes` 表（已在迁移 squash 中移除）。

文件管理层只负责文件系统视图，不负责附件分类、文章引用、用途筛选或业务聚合。相关业务索引见 [附件管理架构](attachment-management-architecture.md)；上传路径规则见 [上传策略架构](upload-policy-architecture.md)。

---

## 1. 目标与边界

文件管理层的目标是给管理员一个稳定、可扫描、接近文件系统的工作台：

- `/` 目录显示所有可见存储实例的 `mount_path`（来自 `attachment.storage_mounts`）。
- `/local`、`/media` 等挂载路径下列出该 `storage_mount_id` 下的**附件行**，按 `object_key` 的目录前缀分组展示为文件夹/文件。
- 列表展示：名称、大小、修改时间、状态、操作。
- 支持上传、下载/预览、重命名（更新 `object_key` / 元数据）、删除（软删附件行）。
- 文件管理 UI 不显示 `purpose`，不提供附件分类 CRUD，不展示文章引用关系。

本层的数据真相源（当前 schema）：

| 表 | 职责 |
| --- | --- |
| `attachment.storage_mounts` | 根目录下的挂载入口与驱动配置 |
| `attachment.attachments` | 挂载内的对象登记；`object_key` 为挂载内唯一定位符 |

---

## 2. 路径模型（当前实现）

### 根目录 `/`

根目录是聚合视图，展示启用或管理员可见的存储实例：

| 名称 | 类型 | 来源 |
| --- | --- | --- |
| `/local` | mount | `storage_mounts.mount_path` |
| `/images` | mount | `storage_mounts.mount_path` |
| `/archive` | mount | `storage_mounts.mount_path` |

根目录下不允许直接上传文件。管理员必须进入某个挂载路径或选择上传策略。

### 挂载目录

进入 `/local` 后，应用层查询 `storage_mount_id = local.id` 的 `attachments` 行，并按 `object_key` 解析层级：

- 将 `object_key` 视为 POSIX 风格路径（如 `blog/2026/hero.png`）。
- “文件夹”由共享前缀推导（如 `blog/`、`blog/2026/`），不必单独建目录表。
- 列表 API 对某一浏览路径 `path` 返回：该路径下的直接子目录（前缀）与直接文件（`object_key` 匹配且不再有更深一层 `/`）。

| 浏览路径 | mount_path | 示例 object_key |
| --- | --- | --- |
| `/local` | `/local` | `readme.md`（挂载根下文件） |
| `/local/blog` | `/local` | `blog/post.md` |
| `/local/blog` | `/local` | `blog/assets/a.png` |

**约束（数据库）**：同一 mount 下活跃行的 `(storage_mount_id, object_key)` 唯一（`ux_attachment_attachments_mount_object_key`）。

---

## 3. 与附件管理的关系

- 每条可浏览的“文件”对应一行 `attachment.attachments`（含 `filename`、`mime_type`、`size`、`upload_status` 等）。
- 业务引用（文章封面、头像等）指向 `attachments.id`，不经过独立的文件节点表。
- 文件管理页不维护 `attachment.categories`；分类与引用聚合在附件管理页。

---

## 4. API 草案

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/files?path=/` | 根目录列出存储实例挂载路径。 |
| `GET` | `/api/v1/files?path=/local/blog` | 按 `object_key` 前缀列出子目录与文件。 |
| `POST` | `/api/v1/files/upload` | 上传到指定路径或上传策略；创建 `attachments` 行。 |
| `PATCH` | `/api/v1/files/:id` | 重命名或更新元数据（必要时同步对象存储 key）。 |
| `DELETE` | `/api/v1/files/:id` | 软删附件行；删除前检查业务引用。 |
| `GET` | `/api/v1/files/:id/content` | 下载或预览（经 mount 解析驱动）。 |

根目录项可使用 `type = "mount"`，并包含 `storage_mount_id`、`mount_path`、`driver_name`、`status`。

---

## 5. Studio 交互

- 顶部：面包屑、刷新、上传。
- 主体：名称、大小、修改时间、操作。
- 根目录：仅挂载路径。
- 挂载内：由 `object_key` 前缀展示的目录与文件。

文件页不展示：附件用途、分类 CRUD、文章引用、引用计数（见附件管理页）。

---

## 6. 删除与移动

- 删除：软删 `attachments.deleted_at`；若仍被 `identity.users.avatar_attachment_id`、内容封面等引用，应由附件管理层阻止或要求先解除引用。
- 重命名/移动：更新 `object_key`（及对象存储中的实际对象）；一期可仅支持同 mount 内重命名；跨前缀“移动”需明确对象复制/删除策略后再实现。

---

## 7. 验收标准（当前阶段）

- `/` 能列出所有存储实例挂载路径。
- 进入挂载后能按 `object_key` 前缀列出文件（无需 `file_nodes` 表）。
- 文件页无用途字段、无分类 CRUD、无业务引用聚合。
- 上传后创建 `attachments` 行并写入正确 `storage_mount_id` + `object_key`。
- 下载经后端权限与驱动解析，不暴露 provider 密钥。

---

## 8. 规划中：OpenList 式目录节点表（未实现）

若未来需要库内物化目录树（高效子树、空文件夹、与对象存储弱一致浏览），可重新引入 `attachment.file_nodes` 并由迁移增量添加。**不要**再添加已删除的 `015_attachment_drop_file_nodes.sql`；当前仓库以 squash 后的 `002_attachment_attachments.sql` 为准。

引入节点表时需同步：API、Studio 文件页、附件 `file_node_id`（若恢复）及迁移文档。
