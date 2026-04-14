# Beehive Blog v2 微服务契约设计

## 1. 目标

本文件定义 v2 第一阶段各微服务的职责边界、核心数据归属、接口归属和事件协作方式。

目的：

- 防止服务边界模糊
- 防止同一业务规则散落在多个服务里
- 为 go-zero 服务骨架与目录划分提供依据

## 2. 服务拆分原则

第一阶段采用“领域边界优先、数量受控”的微服务策略。

不追求过细拆分，先确保：

- 业务边界清晰
- 读写链路清晰
- 搜索与 AI 可独立演进

## 3. 第一阶段服务总览

建议第一阶段服务如下：

- `gateway`
- `identity-service`
- `content-service`
- `review-service`
- `search-service`
- `agent-service`
- `indexer-worker`

可第二阶段再补：

- `mcp-server`
- `publish-worker`
- `notifier-worker`

## 4. gateway

### 职责

- 统一对外 API 入口
- 认证上下文注入
- 基础限流
- 路由分发
- 请求追踪 ID 注入
- 统一错误码包装

### 不负责

- 不承载核心业务规则
- 不直接操作数据库
- 不直接执行搜索索引更新

## 5. identity-service

### 职责

- 用户注册
- 用户登录
- token / refresh token
- 当前用户上下文
- Agent client 身份管理

### 数据归属

- `user`
- `agent_client`

### 负责接口

- `POST /api/v2/auth/register`
- `POST /api/v2/auth/login`
- `POST /api/v2/auth/refresh`
- `POST /api/v2/auth/logout`
- `GET /api/v2/auth/me`

### 发布事件

- `user.registered`
- `user.logged_in`
- `agent_client.created`
- `agent_client.updated`

## 6. content-service

### 职责

- 统一内容主实体管理
- 内容版本管理
- 标签管理
- 内容关系管理
- 附件绑定
- 评论管理
- 内容状态与可见性控制

这是第一阶段最核心的服务。

### 数据归属

- `content_item`
- `content_revision`
- `project_profile`
- `experience_profile`
- `timeline_event_profile`
- `portfolio_profile`
- `tag`
- `content_tag`
- `content_relation`
- `attachment`
- `content_attachment`
- `comment`

### 负责接口

- `GET /api/v2/public/articles`
- `GET /api/v2/public/articles/:slug`
- `GET /api/v2/public/projects`
- `GET /api/v2/public/projects/:slug`
- `GET /api/v2/public/experiences`
- `GET /api/v2/public/experiences/:slug`
- `GET /api/v2/public/timeline`
- `GET /api/v2/public/portfolio`
- `GET /api/v2/public/pages/:slug`
- `GET /api/v2/public/tags`
- `GET /api/v2/public/content/:id/comments`
- `POST /api/v2/public/content/:id/comments`
- `GET /api/v2/studio/contents`
- `POST /api/v2/studio/contents`
- `GET /api/v2/studio/contents/:id`
- `PUT /api/v2/studio/contents/:id`
- `PUT /api/v2/studio/contents/:id/status`
- `PUT /api/v2/studio/contents/:id/visibility`
- `PUT /api/v2/studio/contents/:id/ai-access`
- `DELETE /api/v2/studio/contents/:id`
- `GET /api/v2/studio/contents/:id/revisions`
- `GET /api/v2/studio/contents/:id/revisions/:revisionId`
- `POST /api/v2/studio/contents/:id/revisions/:revisionId/restore`
- `GET /api/v2/studio/contents/:id/relations`
- `POST /api/v2/studio/contents/:id/relations`
- `DELETE /api/v2/studio/contents/:id/relations/:relationId`
- `GET /api/v2/studio/tags`
- `POST /api/v2/studio/tags`
- `PUT /api/v2/studio/tags/:id`
- `DELETE /api/v2/studio/tags/:id`
- `GET /api/v2/studio/attachments`
- `POST /api/v2/studio/attachments/upload`
- `GET /api/v2/studio/attachments/:id`
- `DELETE /api/v2/studio/attachments/:id`
- `GET /api/v2/studio/comments`
- `PUT /api/v2/studio/comments/:id/hide`
- `PUT /api/v2/studio/comments/:id/show`
- `DELETE /api/v2/studio/comments/:id`

### 发布事件

- `content.created`
- `content.updated`
- `content.deleted`
- `content.status_changed`
- `content.visibility_changed`
- `content.ai_access_changed`
- `content.revision_created`
- `content.revision_restored`
- `content.relation_changed`
- `attachment.created`
- `comment.created`
- `comment.updated`

### 消费事件

- `review.approved`
- `review.rejected`
- `agent.output.accepted`

### 领域规则归属

以下规则只能放在 content-service：

- 内容状态流转校验
- slug 唯一约束规则
- 内容类型合法性校验
- 关系绑定合法性
- 评论与内容的归属校验

## 7. review-service

### 职责

- 审阅任务创建
- 审阅通过 / 驳回
- AI 输出与版本的审核流控制

### 数据归属

- `review_task`
- `review_decision`

### 负责接口

- `GET /api/v2/studio/reviews`
- `POST /api/v2/studio/reviews/:id/approve`
- `POST /api/v2/studio/reviews/:id/reject`

### 发布事件

- `review.created`
- `review.approved`
- `review.rejected`

### 消费事件

- `agent.output.submitted`
- `content.status_changed`

## 8. search-service

### 职责

- 面向用户和 Studio 提供搜索能力
- 维护检索副本读取能力
- 提供 related content、suggest 等接口

### 数据归属

- `search_document`
- `content_chunk`
- `content_summary`

### 负责接口

- `GET /api/v2/public/search`
- `GET /api/v2/search/query`
- `GET /api/v2/search/suggest`
- `GET /api/v2/search/contents/:id/related`
- `POST /api/v2/search/rebuild/:contentId`
- `POST /api/v2/search/rebuild-all`

### 发布事件

- `search.indexed`
- `search.rebuilt`
- `summary.generated`
- `chunk.generated`

### 消费事件

- `content.created`
- `content.updated`
- `content.deleted`
- `content.revision_created`
- `content.revision_restored`
- `content.status_changed`
- `content.visibility_changed`
- `content.ai_access_changed`
- `content.relation_changed`

## 9. agent-service

### 职责

- 接收 AI 相关任务请求
- 组装上下文
- 生成摘要 / 草稿 / 周报
- 记录 AI 输出及其来源

### 数据归属

- `agent_task`
- `agent_output`
- `agent_output_source`

### 负责接口

- `POST /api/v2/agent/summarize`
- `POST /api/v2/agent/drafts/generate`
- `POST /api/v2/agent/weekly-digest`
- `POST /api/v2/agent/relations/suggest`
- `POST /api/v2/agent/tags/suggest`
- `GET /api/v2/agent/tasks/:id`
- `GET /api/v2/agent/outputs/:id`
- `POST /api/v2/agent/outputs/:id/submit-review`

### 发布事件

- `agent.task.created`
- `agent.output.generated`
- `agent.output.submitted`
- `agent.output.accepted`
- `agent.output.rejected`

### 消费事件

- `content.created`
- `content.updated`
- `search.indexed`
- `summary.generated`

## 10. indexer-worker

### 职责

- 监听内容变更事件
- 构建 search document
- 内容切片
- 摘要生成
- 推送搜索索引

### 不对外暴露 HTTP API

它是内部异步任务服务。

### 消费事件

- `content.created`
- `content.updated`
- `content.deleted`
- `content.revision_created`
- `content.revision_restored`
- `content.status_changed`
- `content.visibility_changed`
- `content.ai_access_changed`

### 发布事件

- `chunk.generated`
- `summary.generated`
- `search.indexed`

## 11. 第二阶段预留服务

### mcp-server

职责：

- MCP tools / resources 暴露
- 协议适配

### publish-worker

职责：

- 站点发布产物生成
- feed / sitemap / export 生成

### notifier-worker

职责：

- 评论提醒
- 系统通知
- webhook

## 12. 服务间调用建议

第一阶段建议：

- 查询优先同步 RPC / HTTP
- 索引、摘要、发布优先走事件异步

同步调用适合：

- 登录鉴权
- 内容详情读取
- Studio 管理操作

异步调用适合：

- 索引更新
- 摘要生成
- AI 任务后处理
- 发布派生任务

## 13. 第一阶段最小上线组合

如果需要控制开发复杂度，建议先按以下顺序落服务：

1. `identity-service`
2. `content-service`
3. `gateway`
4. `search-service`
5. `indexer-worker`
6. `review-service`
7. `agent-service`

## 14. 当前结论

v2 第一阶段的服务边界已经可以明确为：

**identity 负责身份，content 负责内容主数据，review 负责审核，search 负责检索副本与搜索能力，agent 负责 AI 任务，indexer 负责索引与摘要异步链路。**
