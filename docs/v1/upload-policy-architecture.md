# Beehive-Blog v1：上传策略架构

上传策略层负责决定文件上传到哪里、如何命名、允许什么文件、默认如何公开。它位于文件管理和附件管理之间：文件管理选择当前目录或策略，附件管理消费上传后的业务索引。

---

## 1. 目标与边界

上传策略解决的问题：

- 上传时应该选择哪个 `storage_mount_id`。
- 文件应该落在哪个目录或对象键。
- 文件名冲突时如何处理。
- 允许的最大文件大小、MIME、扩展名。
- 默认访问范围。
- 是否创建附件索引或只创建文件节点。

上传策略不负责：

- 管理存储实例配置。
- 展示文件目录。
- 管理附件分类和业务引用。
- 替代后端权限校验。

---

## 2. 默认策略

推荐默认策略：

```text
/{yyyy}/{mm}/{dd}/{uid}/{filename}
```

含义：

| 占位符 | 说明 |
| --- | --- |
| `{yyyy}` | 四位年份。 |
| `{mm}` | 两位月份。 |
| `{dd}` | 两位日期。 |
| `{uid}` | 当前上传用户 ID。 |
| `{filename}` | 清理后的原始文件名。 |

文章内上传可扩展：

```text
/posts/{post_id}/{filename}
```

头像上传可扩展：

```text
/avatars/{uid}/{filename}
```

---

## 3. 策略模型草案

建议后续新增 `attachment.upload_policies` 或等价配置表。

| 字段 | 说明 |
| --- | --- |
| `id` | 主键。 |
| `name` | 策略名称。 |
| `slug` | 稳定标识。 |
| `target_mount_id` | 默认目标存储实例，可为空表示使用当前目录 mount。 |
| `path_template` | 路径模板。 |
| `max_size` | 单文件最大字节数。 |
| `allowed_mime` | 允许 MIME 列表或模式。 |
| `allowed_extensions` | 允许扩展名。 |
| `name_conflict_policy` | `rename | overwrite | reject`。 |
| `default_access_scope` | `private | public`。 |
| `create_attachment` | 是否创建附件索引。 |
| `status` | `active | disabled`。 |

一期可以先不建表，用代码内固定策略承载；但文档和 API 应为策略表演进留出位置。

---

## 4. API 草案

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/upload-policies` | 列出上传策略。 |
| `POST` | `/api/v1/upload-policies` | 创建上传策略。 |
| `PATCH` | `/api/v1/upload-policies/:id` | 更新上传策略。 |
| `DELETE` | `/api/v1/upload-policies/:id` | 禁用或删除上传策略。 |

文件上传接口建议接受：

| 字段 | 说明 |
| --- | --- |
| `policy_id` | 上传策略 ID，可选。 |
| `path` | 当前目录路径，可选。 |
| `storage_mount_id` | 明确目标存储实例，可选。 |
| `file` | multipart 文件字段。 |

解析优先级建议：

1. 显式 `policy_id`。
2. 当前目录 `path` 对应 mount 的默认策略。
3. 系统默认策略。

---

## 5. 与 `purpose` 的关系

当前 `attachments.purpose` 仍存在于数据库和后端校验中，暂不要求立即删除。

后续目标：

- 文件管理上传不向用户展示 `purpose`。
- 上传策略决定路径、格式和默认访问范围。
- 附件管理通过引用类型表达业务用途，例如文章正文、文章封面、头像、系统资源。
- `purpose` 可在后续迁移中退化为内部兼容字段，或被引用类型替代。

---

## 6. 安全与校验

上传策略只是管理入口，后端仍必须执行硬校验：

- 管理员身份和权限。
- 文件大小上限。
- MIME 与扩展名。
- 文件名清理和路径穿越防护。
- 禁用 mount 不允许新写入。
- 私有/公开访问策略。
- 日志统一英文，不记录敏感凭据。

---

## 7. 验收标准

- 上传入口不再要求管理员手动理解 `purpose`。
- 上传文件能稳定落到策略决定的目录。
- 文件名冲突处理可预测。
- 后续新增文章内上传、头像上传时，只需增加策略或引用类型，不需要改文件管理页核心交互。

