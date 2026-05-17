# Beehive-Blog v1：附件管理架构

本文档定义附件管理层。附件管理是文件管理之上的业务索引层，负责回答“这个文件在业务系统里是什么、被谁引用、能否公开、能否删除”。

文件目录和挂载浏览见 [文件管理架构](file-management-architecture.md)；存储实例和驱动见 [文件存储驱动架构](file-storage-driver-architecture.md)。

---

## 1. 目标与边界

附件管理负责：

- 附件元数据索引：上传人、MIME、大小、状态、可见性、上传状态。
- 附件分类：`attachment.categories` 和 `attachment.attachment_categories`。
- 文章、头像、系统配置等业务引用关系。
- 引用计数、孤儿附件、可删除性判断。
- 附件公开/私有访问策略。
- 附件搜索与筛选。

附件管理不负责：

- 根目录挂载路径展示。
- 文件夹导航和 OpenList 式文件列表。
- 存储实例配置和健康检查。
- 具体上传路径模板生成。

---

## 2. 当前模型与目标模型

当前已有模型：

| 模型 | 当前职责 | 目标职责 |
| --- | --- | --- |
| `attachment.attachments` | 附件元数据、存储定位、用途、可见性、上传状态 | 继续作为业务附件索引主表 |
| `attachment.categories` | 附件分类树 | 保留，归属附件管理，不出现在文件管理页 |
| `attachment.attachment_categories` | 附件与分类多对多 | 保留 |
| `attachment.file_nodes` | 文件系统节点 | 由文件管理维护，附件可通过 `file_node_id` 引用 |

`purpose` 暂不要求立即从数据库删除。它是旧附件语义和当前校验逻辑的一部分，但前端文件管理不再展示它。后续可由“附件引用类型 + 上传策略”逐步替代。

---

## 3. 附件分类

附件分类是业务索引，不是文件目录。

正确用法：

- “文章素材”
- “项目截图”
- “头像资源”
- “公开配图”
- “待清理”

错误用法：

- 把分类当 `/local/blog` 这样的路径目录。
- 在文件管理页里用分类筛选替代目录导航。
- 通过分类决定真实对象存储路径。

分类 CRUD 应放在附件管理页或附件管理面板中。

---

## 4. 引用关系

附件管理需要聚合业务引用关系。目标引用类型包括：

| 引用方 | 示例 |
| --- | --- |
| 文章正文 | Markdown、富文本或内容块中的图片/文件引用 |
| 文章封面 | post cover attachment |
| 用户头像 | `identity.users.avatar_attachment_id` |
| 项目/经历 | 后续内容实体的证明材料 |
| 系统设置 | 站点图标、默认头像、邮件附件等后续能力 |

建议新增引用查询：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/attachments/:id/references` | 查询单个附件的业务引用。 |
| `GET` | `/api/v1/attachments/references?attachment_id=` | 批量或筛选引用。 |

引用结果至少表达：

- `source_type`
- `source_id`
- `source_title`
- `relation`
- `status`
- `updated_at`

---

## 5. API 草案

现有 `/api/v1/attachments` 可继续演进为附件管理 API：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/attachments` | 附件索引列表，可按状态、上传人、分类、引用状态筛选。 |
| `GET` | `/api/v1/attachments/:id` | 附件详情。 |
| `PATCH` | `/api/v1/attachments/:id` | 更新展示名、状态、可见性等业务元数据。 |
| `DELETE` | `/api/v1/attachments/:id` | 删除附件索引，删除前必须检查引用关系。 |
| `PUT` | `/api/v1/attachments/:id/categories` | 替换附件分类绑定。 |
| `GET` | `/api/v1/attachments/:id/references` | 查询引用关系。 |
| `GET/POST` | `/api/v1/attachment/categories` | 分类列表和创建。 |
| `GET/PATCH/DELETE` | `/api/v1/attachment/categories/:id` | 分类详情、更新和删除。 |

附件上传入口后续不建议继续直接挂在附件管理 API 上，而应迁移到文件管理上传或上传策略接口；上传完成后再创建/更新附件索引。

---

## 6. Studio 交互

附件管理页应提供：

- 附件表格：名称、MIME、大小、上传人、可见性、上传状态、引用计数、更新时间。
- 分类筛选与分类 CRUD。
- 引用关系抽屉：列出引用它的文章、头像、项目等。
- 孤儿附件视图：无引用、可清理、疑似废弃。
- 删除前检查：有引用时禁止删除或要求先解除引用。

附件管理页可以跳转到文件管理页中的实际文件位置；文件管理页不反向承载附件分类和引用聚合。

---

## 7. 验收标准

- 附件分类只出现在附件管理域。
- 附件详情能看到关联文件节点和存储实例。
- 删除附件前能判断是否被文章、头像或其他业务对象引用。
- 文件管理页移除附件用途和分类管理后，附件管理仍能完成分类和业务索引维护。

