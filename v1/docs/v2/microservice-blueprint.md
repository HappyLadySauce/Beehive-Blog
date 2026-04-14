# Beehive Blog v2 微服务架构蓝图

## 1. 总体原则

v2 采用微服务架构，但不建议做“过细微服务”。

原则上应该按领域边界拆分，而不是按表拆分、按页面拆分。

建议控制在：

- 入口层
- 核心内容层
- 检索层
- AI 协作层
- 异步任务层

## 2. 推荐服务划分

## 2.1 gateway

职责：

- 对外统一 API 入口
- 鉴权、限流、审计
- 路由转发
- 统一错误码和 tracing 注入

对接对象：

- Public Web
- Studio
- 外部 Agent
- MCP Server

## 2.2 identity-service

职责：

- 用户、角色、权限
- 登录、注册、token、refresh token
- 第三方身份接入预留

说明：

这个服务边界明确，适合独立。

## 2.3 content-service

职责：

- 内容实体管理
- 版本管理
- 分类、标签、关系、附件引用
- 发布状态流转

这是 v2 的核心服务。

## 2.4 review-service

职责：

- 草稿审核
- AI 输出审核
- 发布审批
- 审阅记录与意见

说明：

AI 协作体系里，这个服务很重要，建议单独边界。

## 2.5 search-service

职责：

- 索引同步
- 搜索 API
- 聚合筛选
- 相关内容推荐
- 混合检索

依赖：

- PostgreSQL
- Meilisearch 或 Elasticsearch
- 向量索引能力

## 2.6 agent-service

职责：

- 面向智能体的业务入口
- RAG 组装
- AI 草稿生成
- Prompt 模板装配
- 引用结果封装

说明：

不要让模型调用逻辑散落在 content-service 里。

## 2.7 mcp-server

职责：

- 暴露 MCP tools / resources
- 面向 Hermes、OpenClaw、其他 Agent 客户端

建议只做协议适配，不承载复杂业务规则。

复杂业务仍调用：

- search-service
- content-service
- review-service
- agent-service

## 2.8 publish-worker

职责：

- 发布已审核内容
- 生成站点页面所需产物
- 导出 Markdown / Feed / Sitemap
- 未来可兼容 Hexo 静态导出

## 2.9 indexer-worker

职责：

- 内容切片
- 摘要生成
- embedding 生成
- 搜索索引同步

这是检索和 RAG 的关键异步服务。

## 2.10 notifier-worker

职责：

- 邮件通知
- 站内通知
- webhook 推送

## 3. 服务关系

```text
Public Web / Studio / Agent
  -> gateway
    -> identity-service
    -> content-service
    -> review-service
    -> search-service
    -> agent-service
    -> mcp-server

content-service
  -> event bus
    -> indexer-worker
    -> publish-worker
    -> notifier-worker
```

## 4. 事件流建议

建议 v2 从一开始就采用事件驱动的异步链路。

关键事件：

- `content.created`
- `content.updated`
- `content.published`
- `content.archived`
- `draft.submitted`
- `review.approved`
- `review.rejected`
- `agent.output.created`

事件用途：

- 更新搜索索引
- 触发摘要生成
- 触发 embedding
- 触发发布
- 触发通知

## 5. 数据边界建议

### 5.1 强一致主数据

建议保留在 PostgreSQL：

- 用户
- 内容实体
- 版本
- 审阅记录
- 关系
- 发布记录

### 5.2 搜索副本数据

建议同步到：

- Meilisearch
- Elasticsearch

### 5.3 缓存与会话

建议使用：

- Redis

用途：

- token / session
- 热门搜索缓存
- 检索结果缓存
- 异步任务状态

## 6. 不建议一开始就拆出去的内容

为了避免微服务过早失控，以下内容不建议单独拆服务：

- taxonomy-service
- attachment-service
- summary-service
- relation-service

这些可以先归入 content-service 或 indexer-worker 的边界中。

## 7. v2 第一阶段建议最小服务集

如果要控制复杂度，第一阶段可以先落这 5 个：

- gateway
- identity-service
- content-service
- search-service
- indexer-worker

第二阶段再补：

- review-service
- agent-service
- mcp-server
- publish-worker
- notifier-worker

## 8. 当前推荐

推荐把 v2 的微服务目标定义为：

**领域边界清晰、部署可拆分、代码仓库统一、早期不过度细分。**

这样既能满足微服务架构目标，也不会一上来把开发复杂度拉得过高。
