# Beehive-Blog v1：文件管理架构

本文档定义 Studio 文件管理层。它采用类似 OpenList 的文件列表交互：根目录 `/` 展示所有存储实例挂载路径，进入某个挂载后展示该挂载下的目录和文件。

文件管理层只负责文件系统视图，不负责附件分类、文章引用、用途筛选或业务聚合。相关业务索引见 [附件管理架构](attachment-management-architecture.md)；上传路径规则见 [上传策略架构](upload-policy-architecture.md)。

---

## 1. 目标与边界

文件管理层的目标是给管理员一个稳定、可扫描、接近文件系统的工作台：

- `/` 目录显示所有可见存储实例的 `mount_path`。
- `/local`、`/media` 等挂载路径显示对应 `storage_mount_id` 下的 `file_nodes`。
- 文件夹和文件使用统一列表展示：名称、大小、修改时间、状态、操作。
- 支持新建文件夹、上传、下载/预览、重命名、移动、删除。
- 文件管理 UI 不显示 `purpose`，不提供附件分类 CRUD，不展示文章引用关系。

本层的数据真相源是数据库：

- `attachment.storage_mounts` 表达根目录下的挂载入口。
- `attachment.file_nodes` 表达挂载内目录/文件层级。
- `attachment.attachments.file_node_id` 可选关联文件节点，用于业务附件索引。

---

## 2. 路径模型

### 根目录 `/`

根目录不是实际存储目录，而是聚合视图。它展示所有启用或管理员可见的存储实例：

| 名称 | 类型 | 来源 |
| --- | --- | --- |
| `/local` | mount | `storage_mounts.mount_path` |
| `/images` | mount | `storage_mounts.mount_path` |
| `/archive` | mount | `storage_mounts.mount_path` |

根目录下不允许直接上传文件。管理员必须进入某个挂载路径或选择上传策略。

### 挂载目录

进入 `/local` 后，列表查询 `storage_mount_id = local.id` 且 `parent_id IS NULL` 的 `file_nodes`。进入 `/local/blog` 后，先解析 `/blog` 对应目录节点，再查询其子节点。

`file_nodes.full_path` 是挂载内部路径，不包含 mount path：

| 浏览路径 | mount_path | file_nodes.full_path |
| --- | --- | --- |
| `/local` | `/local` | `/` |
| `/local/blog` | `/local` | `/blog` |
| `/local/blog/a.png` | `/local` | `/blog/a.png` |

---

## 3. 数据模型

### `attachment.file_nodes`

文件节点表表达目录和文件：

| 字段 | 说明 |
| --- | --- |
| `storage_mount_id` | 所属存储实例。 |
| `parent_id` | 父目录节点；挂载根下节点为 `NULL`。 |
| `node_type` | `directory | file`。 |
| `name` | 当前层级名称。 |
| `path` | 物化路径，用于高效子树查询。 |
| `full_path` | 挂载内部可见路径。 |
| `depth` | 树深度。 |
| `sort_order` | 同级排序。 |
| `status` | `active | hidden | archived`。 |

约束应保持与当前迁移一致：

- 同一 mount 下活跃 `full_path` 唯一。
- 同一父目录下活跃 `name` 唯一。
- 根节点同名需单独约束，因为 `parent_id IS NULL`。

### 与附件的关系

`attachment.attachments.file_node_id` 是业务附件对文件节点的可选引用：

- 一个文件节点可以有一个主附件记录，用于下载、权限、业务引用。
- 也允许存在暂未绑定业务附件的文件节点，作为文件管理能力的扩展空间。
- 文件管理不直接编辑附件分类；它只维护文件节点和物理对象。

---

## 4. API 草案

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/files?path=/` | 根目录列出存储实例挂载路径。 |
| `GET` | `/api/v1/files?path=/local/blog` | 列出某个目录下的文件和文件夹。 |
| `POST` | `/api/v1/files/directories` | 在指定目录下新建文件夹。 |
| `POST` | `/api/v1/files/upload` | 上传到指定路径或上传策略。 |
| `PATCH` | `/api/v1/files/:id` | 重命名、移动或更新状态。 |
| `DELETE` | `/api/v1/files/:id` | 删除文件或目录。 |
| `GET` | `/api/v1/files/:id/content` | 下载或预览文件内容。 |

建议 `GET /api/v1/files` 返回：

```json
{
  "path": "/local/blog",
  "items": [
    {
      "id": 12,
      "type": "directory",
      "name": "assets",
      "size": null,
      "updated_at": "2026-05-17T00:00:00Z"
    }
  ]
}
```

根目录的 mount 项可使用 `type = "mount"`，并包含 `storage_mount_id`、`mount_path`、`driver_name`、`status`、`disabled`。

---

## 5. Studio 交互

文件页采用 OpenList 式文件列表：

- 顶部：面包屑路径、刷新、视图切换、上传、新建文件夹。
- 主体：表格或紧凑列表，列为名称、大小、修改时间、状态、操作。
- 根目录：只展示挂载路径，不展示附件列表。
- 挂载目录：展示文件夹和文件，文件夹可点击进入。
- 移动端：保留名称和操作优先，大小/时间可压缩为次级信息。

文件页不展示：

- 附件用途。
- 附件分类筛选。
- 分类 CRUD。
- 文章引用关系。
- 引用计数和孤儿附件。

这些能力属于附件管理页。

---

## 6. 删除与移动

删除文件节点时必须区分：

- 无业务附件引用的文件：可按文件管理策略删除。
- 有附件引用的文件：必须查询附件管理层的引用关系，若仍被文章、头像或系统配置引用，则阻止删除或要求显式解除引用。
- 目录删除：必须定义是否允许递归删除；默认建议先不支持非空目录删除。

移动文件时必须同步：

- `file_nodes.parent_id`
- `file_nodes.full_path`
- `file_nodes.path`
- 子树路径
- 对应附件的 `object_key` 是否需要迁移

一期建议只实现重命名和同目录操作；跨目录移动应在对象迁移策略明确后实现。

---

## 7. 验收标准

- `/` 能列出所有存储实例挂载路径。
- 进入挂载后能列出该挂载下文件夹和文件。
- 文件页无用途字段、无分类 CRUD、无业务引用聚合。
- 文件上传后创建或绑定 `file_nodes`。
- 文件下载仍通过后端权限和驱动解析，不暴露 provider 密钥。
- 移动端无页面级横向溢出，文件列表可滚动或压缩显示。

