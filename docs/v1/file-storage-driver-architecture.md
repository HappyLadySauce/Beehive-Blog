# Beehive-Blog v1：文件存储驱动架构

本文档描述 **HTTP API v1** 与数据库层的文件存储驱动设计，供后续迁移、后端接口、Studio 管理端与测试开发对齐。实现变更时须同步更新本文档。

**延伸阅读**：[产品设计原则](../product-principles.md)、[v1 登录与注册规则](./login-and-registration-rules.md)、[OpenList 驱动扩展](https://doc.openlist.team/guide/drivers/develop)、[OpenList 通用存储配置](https://doc.oplist.org/guide/drivers/common.html)、[OpenList driver API](https://openlistteam.github.io/docs/zh/guide/api/admin/driver.html)、[OpenList storage API](https://openlistteam.github.io/docs/zh/guide/api/admin/storage.html)。

---

## 1. 背景与目标

旧附件系统以 `attachment.attachments` 为核心，直接在附件行中保存 `storage_type`、`bucket`、`object_key`、`local_path` 等字段。这种设计能覆盖 `local`、`s3`、`oss` 的早期场景，但有两个明显限制：

1. **管理员无法管理存储实例**：存储配置来自应用配置与设置项，不适合在 Studio 中添加、禁用、检查或切换。
2. **文件无法稳定绑定不同存储实例**：同一 `storage_type` 下不能表达多个 S3 bucket、多个本地根目录或后续 WebDAV / OpenList / OneDrive 等驱动实例。

新设计的目标是引入“驱动模板 + 存储实例 + 文件节点 + 附件对象”的分层：

- 管理员可在 Studio 中选择驱动并创建多个存储实例。
- 不同文件可绑定不同 `storage_mount_id`。
- 文件元数据、业务引用、目录结构以本项目数据库为主，驱动只负责字节流存取与 provider 交互。
- 旧附件数据在迁移中尽量保留，但最终代码和数据库不再保留旧定位字段。

---

## 2. OpenList 参考点

OpenList 的存储设计值得参考，但本项目不直接复制其代码或完整文件系统能力。

可借鉴点：

- **驱动模板与存储实例分离**：OpenList 有 driver info / driver names API，存储实例则保存 `mount_path`、`driver`、`addition` 等配置。
- **挂载路径唯一**：OpenList 将 `mount_path` 作为挂载项标识；重复路径会导致存储创建失败。
- **通用字段 + 驱动额外配置**：通用配置覆盖排序、备注、缓存、代理等；驱动特有字段放在 `addition` JSON 中。
- **驱动可扩展**：新增驱动主要通过注册新的 driver package 和配置模板完成，而不是修改所有业务代码。

本项目的取舍：

- **数据库为主**：OpenList 更像可浏览远端网盘的文件列表程序；Beehive-Blog 的附件、内容引用、权限与发布流程必须以本项目数据库为真相源。
- **不做一期全量网盘浏览器**：一期只把当前 `local`、`s3`、`oss` 变成可管理的存储实例，给后续 OpenList / WebDAV / OneDrive 等驱动留接口。
- **不沿用 `addition` 字符串**：数据库使用 `JSONB config` 和 `JSONB storage_metadata`，避免应用层反复解析字符串。

---

## 3. 本项目设计原则

1. **驱动是代码能力，存储实例是数据库配置**
   `Driver` 由 Go 代码注册，声明支持的能力与配置 schema；`StorageMount` 由管理员创建，保存某个驱动的实际配置。

2. **文件绑定存储实例，不再绑定存储类型**
   附件必须写入 `storage_mount_id`，并通过 `storage_mount_id + object_key + storage_metadata` 定位对象。`storage_type`、`bucket`、`local_path` 不属于最终模型。

3. **业务元数据与对象存储位置分离**
   附件的用途、归属、可见性、上传状态、分类绑定仍由数据库管理；驱动只关心 `object_key` 和存储配置。

4. **禁用只阻止新写入，不默认切断旧文件读取**
   禁用 storage mount 后，不允许新上传；已有文件是否可读由后端策略控制，默认保留读取能力，避免误伤已发布内容。

5. **目标模型优先，旧字段只存在于迁移输入中**
   迁移允许读取旧 `storage_type / bucket / local_path` 来回填新字段；迁移完成后新代码、新响应和最终 schema 均不得依赖旧字段。

---

## 4. 领域模型

| 模型 | 职责 |
|------|------|
| `StorageDriver` | 驱动模板。描述驱动名、展示名、配置 schema、能力集合和启用状态。 |
| `StorageMount` | 管理员创建的存储实例。包含挂载路径、驱动类型、实例配置、状态、默认标记。 |
| `FileNode` | 虚拟文件系统节点。表达目录 / 文件层级、完整路径、排序和挂载归属。 |
| `Attachment` | 业务附件对象。表达归属、用途、MIME、大小、访问范围、上传状态、生命周期和业务引用。 |
| `DriverRegistry` | 后端运行时注册表。根据 `driver_name` 找到具体驱动实现。 |
| `DriverBackend` | 驱动接口。提供保存、预签名上传、预签名下载、删除、健康检查等能力。 |

推荐读取链路：

```mermaid
flowchart LR
  Attachment["attachment.attachments"]
  Mount["attachment.storage_mounts"]
  Registry["DriverRegistry"]
  Driver["DriverBackend"]
  Provider["Local / S3 / OSS / Future Provider"]

  Attachment --> Mount
  Mount --> Registry
  Registry --> Driver
  Driver --> Provider
```

---

## 5. 数据库表设计

所有新增表继续放在 `attachment` schema 下，并沿用 GORM 标准时间字段：`created_at`、`updated_at`、`deleted_at`。唯一约束优先使用 `WHERE deleted_at IS NULL` 部分索引，允许软删后复用。

### 5.1 `attachment.storage_drivers`

驱动模板表，用于 Studio 渲染驱动选择与配置表单。驱动是否真的可用仍以 Go 代码注册为准，表中数据服务于展示、校验和管理端体验。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `BIGSERIAL PRIMARY KEY` | 主键 |
| `name` | `VARCHAR(64) NOT NULL` | 内部驱动名，如 `local`、`s3`、`oss` |
| `display_name` | `VARCHAR(128) NOT NULL` | 管理端展示名 |
| `description` | `TEXT NULL` | 驱动说明 |
| `config_schema` | `JSONB NOT NULL DEFAULT '{}'` | 驱动配置表单 schema |
| `capabilities` | `JSONB NOT NULL DEFAULT '{}'` | 能力集合，如上传、下载、删除、预签名、健康检查 |
| `status` | `VARCHAR(16) NOT NULL DEFAULT 'active'` | `active | disabled` |
| `created_at` / `updated_at` / `deleted_at` | `TIMESTAMPTZ` | 标准时间字段 |

约束与索引：

- 活跃行内 `name` 唯一。
- `status IN ('active', 'disabled')`。
- 后端启动时可校验数据库 driver 与代码 registry 是否一致，缺失时记录日志，不阻断启动。

### 5.2 `attachment.storage_mounts`

存储实例表，相当于 OpenList mounted storage，但面向本项目附件与内容系统裁剪。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `BIGSERIAL PRIMARY KEY` | 主键 |
| `driver_name` | `VARCHAR(64) NOT NULL` | 绑定的驱动名 |
| `mount_path` | `VARCHAR(512) NOT NULL` | 唯一挂载路径，如 `/local`、`/images` |
| `name` | `VARCHAR(128) NOT NULL` | 管理端名称 |
| `remark` | `TEXT NULL` | 管理备注 |
| `config` | `JSONB NOT NULL DEFAULT '{}'` | 驱动实例配置 |
| `order_index` | `INT NOT NULL DEFAULT 0` | 管理端排序 |
| `is_default` | `BOOLEAN NOT NULL DEFAULT false` | 是否默认存储 |
| `disabled` | `BOOLEAN NOT NULL DEFAULT false` | 是否禁用 |
| `status` | `VARCHAR(16) NOT NULL DEFAULT 'unknown'` | `unknown | work | error` |
| `last_checked_at` | `TIMESTAMPTZ NULL` | 最近健康检查时间 |
| `last_error` | `TEXT NULL` | 最近错误 |
| `created_by` | `BIGINT NULL` | 创建管理员 |
| `created_at` / `updated_at` / `deleted_at` | `TIMESTAMPTZ` | 标准时间字段 |

约束与索引：

- 活跃行内 `mount_path` 唯一。
- `mount_path` 必须以 `/` 开头，不允许空路径、`//`、`.`、`..` 路径片段。
- 同一时间最多一个 `is_default = true AND disabled = false AND deleted_at IS NULL`。
- `status IN ('unknown', 'work', 'error')`。
- 索引：`driver_name`、`disabled/order_index/id`、`status`、`created_by`。

### 5.3 `attachment.file_nodes`

虚拟文件系统节点表。它表达 Studio 文件管理视图中的目录与文件层级；业务附件可以绑定文件节点，也可以只作为无目录对象存在。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `BIGSERIAL PRIMARY KEY` | 主键 |
| `parent_id` | `BIGINT NULL` | 自引用父节点 |
| `storage_mount_id` | `BIGINT NOT NULL` | 所属存储实例 |
| `node_type` | `VARCHAR(16) NOT NULL` | `directory | file` |
| `name` | `VARCHAR(255) NOT NULL` | 节点名 |
| `path` | `TEXT NOT NULL` | 物化路径，如 `/1/9/20/` |
| `full_path` | `TEXT NOT NULL` | 用户可见路径，如 `/images/avatar/a.png` |
| `depth` | `INT NOT NULL DEFAULT 0` | 树深度 |
| `sort_order` | `INT NOT NULL DEFAULT 0` | 同级排序 |
| `status` | `VARCHAR(16) NOT NULL DEFAULT 'active'` | `active | hidden | archived` |
| `created_at` / `updated_at` / `deleted_at` | `TIMESTAMPTZ` | 标准时间字段 |

约束与索引：

- `parent_id` 引用 `attachment.file_nodes(id)`。
- `storage_mount_id` 引用 `attachment.storage_mounts(id)`。
- 活跃行内同一 `parent_id + lower(name)` 唯一；根节点需单独处理 `parent_id IS NULL`。
- 活跃行内 `storage_mount_id + full_path` 唯一。
- `node_type IN ('directory', 'file')`。
- `status IN ('active', 'hidden', 'archived')`。
- 物化路径与深度由触发器维护，可复用 `attachment.categories` 的触发器设计思路。

### 5.4 `attachment.attachments` 调整

现有附件表保留业务元数据职责，但存储定位改为引用存储实例。

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `storage_mount_id` | `BIGINT NOT NULL` | 指向 `attachment.storage_mounts(id)` |
| `file_node_id` | `BIGINT NULL` | 指向 `attachment.file_nodes(id)` |
| `object_key` | `VARCHAR(1024) NOT NULL` | 所选 mount 内的对象键 |
| `storage_metadata` | `JSONB NOT NULL DEFAULT '{}'` | provider 扩展元数据，如 version id、headers、etag 详情 |

字段语义调整：

- `object_key` 保留并作为所有驱动的内部对象键。
- `bucket`、`local_path`、`storage_type` 已从最终模型删除，只能在迁移脚本中作为旧数据来源出现。
- 远端下载、直传、删除统一通过 `storage_mount_id -> driver_name -> DriverBackend` 解析。

最终约束：

- `storage_mount_id` 与 `object_key` 必填。
- 活跃行内 `storage_mount_id + object_key` 唯一。
- API 响应不输出 `storage_type`、`bucket`、`local_path`。

---

## 6. 上传、下载与删除流程

### 6.1 上传

管理员上传时可显式传 `storage_mount_id`；未传则使用默认启用 storage mount。

流程：

1. 校验管理员身份、文件用途、MIME、大小、访问范围。
2. 解析 storage mount：显式 `storage_mount_id` 优先，否则查找默认启用 mount。
3. 校验 mount 未软删、未禁用、driver 已注册。
4. 生成 `object_key`。
5. local 类驱动走服务端 `Save`；远端类驱动走 `PresignUpload` 并创建 `pending` 附件行。
6. 附件行写入 `storage_mount_id`、`object_key`、`storage_metadata`、`upload_status`。

禁用策略：

- `disabled = true` 的 mount 不允许新上传。
- `status = error` 是否允许新上传由驱动能力决定；默认拒绝，避免把文件写入不可用存储。

### 6.2 下载

下载以附件行为入口，不直接从前端暴露 provider 配置。

流程：

1. 读取附件并校验 `upload_status = ready`、业务 `status = active`。
2. 非管理员读取还需校验 `access_scope = public` 及后续内容发布策略。
3. 读取 `storage_mount_id` 对应的 mount。
4. 通过 `driver_name` 获取驱动实现。
5. local 类驱动返回本地文件；远端类驱动返回预签名下载 URL 或走服务端代理。

读取策略：

- mount 被禁用时，默认允许读取已有 ready 文件。
- mount 被软删或驱动不存在时，返回不可用错误，并记录英文日志。

### 6.3 删除

删除分为数据库软删与对象清理，不在同一事务内强绑定。

推荐流程：

1. 管理员删除附件时先软删 `attachment.attachments`。
2. 头像等业务外键按现有触发器解除引用。
3. 后续由清理任务根据 `storage_mount_id + object_key` 调用驱动删除真实对象。
4. 删除真实对象失败时记录重试状态，不回滚附件软删。

---

## 7. 管理端能力

### 7.1 驱动查询

建议 API：

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/file-drivers` | 列出可用驱动模板 |
| `GET` | `/api/v1/file-drivers/:name` | 查询单个驱动配置 schema 与能力 |

返回应包含：

- `name`
- `display_name`
- `description`
- `config_schema`
- `capabilities`
- `status`

### 7.2 存储实例管理

建议 API：

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/storage-mounts` | 列出存储实例 |
| `POST` | `/api/v1/storage-mounts` | 创建存储实例 |
| `GET` | `/api/v1/storage-mounts/:id` | 查询存储实例 |
| `PATCH` | `/api/v1/storage-mounts/:id` | 更新名称、备注、配置、排序、默认标记 |
| `POST` | `/api/v1/storage-mounts/:id/enable` | 启用存储实例 |
| `POST` | `/api/v1/storage-mounts/:id/disable` | 禁用存储实例 |
| `POST` | `/api/v1/storage-mounts/:id/check` | 执行健康检查 |
| `DELETE` | `/api/v1/storage-mounts/:id` | 软删除存储实例 |

删除约束：

- 已有附件引用的 mount 默认不允许删除，只允许禁用。
- 若产品需要“强制删除”，必须先定义文件读取失败、清理任务和审计日志策略。

### 7.3 Studio 展示

Studio 存储页至少需要：

- 驱动选择下拉。
- 存储实例卡片或表格。
- 状态：正常、错误、禁用、未知。
- 操作：添加、编辑、启用、禁用、健康检查。
- 默认存储标记。

---

## 8. 迁移与兼容策略

这是破坏式迁移。迁移脚本可以读取旧字段保留现有附件数据，但迁移完成后旧字段、旧配置和旧 API 均不再兼容。

### 阶段一：新增结构

- 新增 `storage_drivers`、`storage_mounts`、`file_nodes`。
- 为 `attachments` 新增 `storage_mount_id`、`file_node_id`、`storage_metadata`。
- 初始化 `local`、`s3`、`oss` 驱动模板。
- 创建默认 local storage mount；远端 mount 后续由管理员通过 `/api/v1/storage-mounts` 创建。

### 阶段二：回填旧数据

- `storage_type = local`：回填到 local mount，`object_key` 优先使用已有 `object_key`，否则由 `local_path` 回填。
- `storage_type = s3`：回填到可用 s3 mount，`object_key` 保持不变。
- `storage_type = oss`：回填到可用 oss mount，`object_key` 保持不变。
- 无法映射 mount 或缺少 `object_key` 的附件必须导致迁移失败，阻止继续 drop 旧字段。

### 阶段三：切换写路径

- 上传和预签名接口接受 `storage_mount_id`。
- 未传 `storage_mount_id` 时使用默认 mount。
- 新附件必须写入 `storage_mount_id` 和 `object_key`。
- 新代码不再根据 `storage_type` 选择驱动。

### 阶段四：收敛旧字段

- 确认所有附件均有 `storage_mount_id` 与 `object_key`。
- 将 `attachments.storage_mount_id` 与 `attachments.object_key` 设为 `NOT NULL`。
- 删除 `storage_type`、`bucket`、`local_path` 及其索引、CHECK、注释、Go model 字段和 API 响应字段。
- `setting.application_settings.payload` 不再承载任何文件服务配置；迁移删除旧 `attachment` key，运行时解析旧 payload 时直接忽略该 key。

### 上传策略

上传策略是代码内固定规则，不进入 `options`、`config.yaml`、CLI flags 或 `settings.application_settings`：

- 单文件最大 `2 GiB`。
- 普通附件允许所有 MIME。
- avatar 必须是 `image/*`。
- 预签名 URL TTL 固定为 `15m`。

如果后续需要按存储实例调整策略，应扩展 `storage_mounts.config` 或新增独立策略表，不回到全局 settings/options。

---

## 9. 测试与验收标准

数据库迁移测试：

- 活跃 `storage_drivers.name` 唯一。
- 活跃 `storage_mounts.mount_path` 唯一。
- 同一时间最多一个启用默认 mount。
- `file_nodes` 同级重名被拒绝，软删后可复用。
- `attachments.storage_mount_id` 可正确引用 mount。

后端单元测试：

- 未传 `storage_mount_id` 时选择默认 mount。
- 传入不存在、软删或禁用的 mount 时上传失败。
- 不同文件可绑定不同 mount。
- local 上传仍能落盘并下载。
- s3 / oss 预签名从 mount config 获取 bucket 与 base URL。
- 删除附件通过 `storage_mount_id + object_key` 调用 driver，driver 删除失败只记录日志，不阻断软删。

接口测试：

- 管理员可列出驱动、创建 mount、禁用 mount、执行健康检查。
- 非管理员不能访问驱动与存储实例管理接口。
- 禁用 mount 后不允许新上传，但已有 ready 文件默认仍可下载。

前端测试：

- Studio 存储页可展示驱动列表与存储实例状态。
- 添加存储实例时按 `config_schema` 渲染字段。
- 上传入口可选择 storage mount；不选择时使用默认 mount。

---

## 10. 后续扩展边界

后续可扩展：

- WebDAV、OpenList、OneDrive、Google Drive 等新驱动。
- 存储健康检查任务与状态告警。
- 对象清理任务与失败重试。
- 下载代理、302 跳转、服务端中转等策略。
- 敏感配置加密存储。
- 文件节点索引、搜索、AI 可读边界。

后续暂不纳入一期：

- 完整远端网盘实时浏览。
- 多用户私有网盘挂载。
- WebDAV / FTP / SFTP 对外服务。
- OpenList 级别的缓存策略、任务队列、离线下载和批量打包下载。

---

*若本文与代码、迁移或 OpenAPI/Swagger 不一致，以实现和最新迁移为准，并在同一变更中同步修订本文。*
