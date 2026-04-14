# Beehive Blog v2 服务启动顺序与落地步骤

## 1. 目标

本文件定义 v2 从“只有文档”到“第一批服务可运行”的实际落地顺序。

目标：

- 明确先做什么，后做什么
- 避免同时起太多服务导致混乱
- 让 goctl、数据库、网关、搜索按依赖顺序落地

## 2. 总体原则

先做最小闭环，不做平铺式开发。

优先顺序：

1. 仓库骨架
2. 数据库
3. 身份与内容服务
4. 网关
5. 搜索
6. 审阅
7. AI

## 3. Phase 0：仓库初始化

目标：

- 建立 monorepo 结构
- 固定目录布局

建议动作：

1. 创建 `apps/`
2. 创建 `shared/`
3. 创建 `sql/migrations/`
4. 创建 `scripts/codegen/`
5. 创建 `deploy/`

完成标准：

- 仓库目录结构与 `v2-gozero-project-layout.md` 一致

## 4. Phase 1：数据库先行

目标：

- 建立主数据结构

建议先落 migration：

1. `001_users_and_agents.sql`
2. `002_content_core.sql`
3. `003_content_profiles.sql`
4. `004_tags_and_relations.sql`
5. `005_attachments.sql`
6. `006_comments.sql`
7. `007_reviews.sql`

此阶段先不急着上：

- `008_agent_tasks.sql`
- `009_search_derivatives.sql`

完成标准：

- PostgreSQL 能初始化成功
- 基础表结构可用

## 5. Phase 2：identity-service

目标：

- 打通注册、登录、认证上下文

建议步骤：

1. 写 `identity.api`
2. 写 `identity.proto`
3. 用 `goctl` 生成 API/RPC 骨架
4. 实现用户注册、登录、刷新 token、当前用户

完成标准：

- `register/login/me` 跑通

## 6. Phase 3：content-service

目标：

- 打通核心内容闭环

建议步骤：

1. 写 `content.api`
2. 写 `content.proto`
3. 用 `goctl` 生成骨架
4. 手工补 `domain/` 和 `repository/`
5. 实现：
   - 内容创建
   - 内容更新
   - 内容详情
   - 内容列表
   - 状态变更
   - 标签绑定
   - 关系绑定

完成标准：

- owner 能创建和编辑 article/note/project/experience

## 7. Phase 4：gateway

目标：

- 统一对外入口

建议步骤：

1. 写 `gateway.api`
2. 用 `goctl` 生成 API 骨架
3. 对接 identity/content
4. 实现认证中间件和基础错误码封装

完成标准：

- 前端只需要访问 gateway

## 8. Phase 5：search-service

目标：

- 打通搜索能力

建议步骤：

1. 写 `search.api`
2. 写 `search.proto`
3. 用 `goctl` 生成骨架
4. 手工补 `searchengine/`
5. 接 Meilisearch 或 PostgreSQL FTS
6. 先实现：
   - query
   - suggest
   - related content

完成标准：

- 公开内容可搜索
- Studio 可搜索内容

## 9. Phase 6：indexer-worker

目标：

- 把内容变更自动同步到搜索

建议步骤：

1. 手工创建 worker 结构
2. 接入事件表或 Redis Stream
3. 实现：
   - content -> search document
   - reindex
   - summary generate 占位

完成标准：

- 内容更新后可触发索引更新

## 10. Phase 7：review-service

目标：

- 打通待审流程

建议步骤：

1. 写 `review.api`
2. 写 `review.proto`
3. 生成骨架
4. 实现 review task、approve/reject

完成标准：

- AI 输出或待审 revision 能进入审核流

## 11. Phase 8：agent-service

目标：

- 打通 AI 摘要和草稿链路

建议步骤：

1. 写 `agent.api`
2. 写 `agent.proto`
3. 生成骨架
4. 手工补 `providers/`、`prompts/`
5. 实现：
   - summarize
   - draft generate
   - submit review

完成标准：

- 可以根据已有内容生成 AI 草稿并提交审阅

## 12. 第一批 `.api` / `.proto` 编写顺序

建议顺序：

1. `identity.api`
2. `identity.proto`
3. `content.api`
4. `content.proto`
5. `gateway.api`
6. `search.api`
7. `search.proto`
8. `review.api`
9. `review.proto`
10. `agent.api`
11. `agent.proto`

## 13. 第一批必须实现的接口

## 13.1 identity

- register
- login
- refresh
- me

## 13.2 content

- create content
- update content
- get content detail
- list contents
- update status
- manage tags
- manage relations

## 13.3 gateway

- route auth
- route public content
- route studio content

## 13.4 search

- query
- suggest

## 14. 第一阶段最低运行里程碑

满足以下条件即可视为第一批服务起盘成功：

- 用户可注册登录
- owner 可创建内容
- 内容可被列出和查看
- gateway 可统一对外
- 搜索能查到公开内容

## 15. 建议的周节奏

如果按自然开发节奏推进，建议：

### 第 1 周

- 仓库骨架
- migration
- identity-service

### 第 2 周

- content-service
- gateway

### 第 3 周

- search-service
- indexer-worker

### 第 4 周

- review-service
- agent-service

这只是推荐顺序，不是强制工期。

## 16. 当前结论

v2 的第一批服务启动顺序应固定为：

**数据库 -> identity -> content -> gateway -> search -> indexer -> review -> agent**

这样依赖链最清晰，返工最少。
