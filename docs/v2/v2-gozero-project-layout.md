# Beehive Blog v2 go-zero 项目布局设计

## 1. 目标

本文件定义 v2 在 `go-zero` 框架下的推荐仓库结构、服务目录布局、共享模块组织方式，以及 `goctl` 的使用边界。

目标：

- 让文档中的服务边界可以直接映射到仓库
- 让 `goctl` 生成代码和手写领域代码各归其位
- 避免后期目录混乱

## 2. 设计原则

### 2.1 用 go-zero 生成骨架，不用它主导领域设计

先有领域模型和服务边界，再用 `goctl` 生成 API / RPC 骨架。

不要先跑脚手架再倒推架构。

### 2.2 单仓库，多服务

v2 建议采用 monorepo。

原因：

- 当前服务边界已明确
- 公共模型、事件、DTO、proto、配置需要共享
- 搜索、AI、MCP、worker 之间协作紧密

### 2.3 共享代码要严格收敛

共享目录只放：

- 通用库
- 契约
- 配置模型
- 事件定义

不要把业务逻辑塞进 shared。

## 3. 推荐仓库结构

```text
beehive-blog/
  docs/
    archive-v1/
    v2/
  docker/
  deploy/
    local/
    staging/
    production/
  apps/
    gateway/
    identity-service/
    content-service/
    review-service/
    search-service/
    agent-service/
    indexer-worker/
    mcp-server/
  shared/
    proto/
    contracts/
    events/
    libs/
    configs/
    constants/
  scripts/
    dev/
    codegen/
    db/
  sql/
    migrations/
    seeds/
```

## 4. apps 目录建议

## 4.1 gateway

用途：

- 对外 HTTP 入口
- 聚合认证上下文
- 路由分发

推荐结构：

```text
apps/gateway/
  api/
    gateway.api
  cmd/
    gateway/
      main.go
  internal/
    config/
    handler/
    logic/
    svc/
    types/
  etc/
    gateway.yaml
```

## 4.2 identity-service

建议拆成：

- `api/`：对外 HTTP 接口
- `rpc/`：供内部服务调用的 RPC 接口

推荐结构：

```text
apps/identity-service/
  api/
    identity.api
  rpc/
    identity.proto
  cmd/
    api/
      main.go
    rpc/
      main.go
  internal/
    api/
      config/
      handler/
      logic/
      svc/
      types/
    rpc/
      server/
      logic/
      svc/
  etc/
    identity-api.yaml
    identity-rpc.yaml
```

## 4.3 content-service

这是 v2 核心服务。

推荐结构：

```text
apps/content-service/
  api/
    content.api
  rpc/
    content.proto
  cmd/
    api/
      main.go
    rpc/
      main.go
  internal/
    api/
      config/
      handler/
      logic/
      svc/
      types/
    rpc/
      server/
      logic/
      svc/
    domain/
    repository/
  etc/
    content-api.yaml
    content-rpc.yaml
```

说明：

- `domain/` 放领域规则
- `repository/` 放数据库访问适配
- 不要把所有规则堆进 `logic/`

## 4.4 review-service

结构与 content-service 类似，但边界更小：

```text
apps/review-service/
  api/
  rpc/
  cmd/
  internal/
    api/
    rpc/
    domain/
    repository/
  etc/
```

## 4.5 search-service

由于搜索既有对外接口，也有内部索引逻辑，建议结构如下：

```text
apps/search-service/
  api/
    search.api
  rpc/
    search.proto
  cmd/
    api/
      main.go
    rpc/
      main.go
  internal/
    api/
    rpc/
    domain/
    repository/
    searchengine/
  etc/
```

说明：

- `searchengine/` 用于封装 Meilisearch / Elasticsearch 客户端
- 不要把搜索引擎 SDK 直接散落到业务逻辑里

## 4.6 agent-service

建议结构：

```text
apps/agent-service/
  api/
    agent.api
  rpc/
    agent.proto
  cmd/
    api/
      main.go
    rpc/
      main.go
  internal/
    api/
    rpc/
    domain/
    repository/
    providers/
    prompts/
  etc/
```

说明：

- `providers/` 放模型调用适配层
- `prompts/` 放服务内 prompt 组装逻辑或引用共享模板

## 4.7 indexer-worker

worker 不一定适合完整按 API/RPC 结构生成，建议手写为主。

推荐结构：

```text
apps/indexer-worker/
  cmd/
    worker/
      main.go
  internal/
    config/
    consumer/
    jobs/
    svc/
    searchengine/
    summarizer/
  etc/
    indexer-worker.yaml
```

## 4.8 mcp-server

第二阶段补，建议结构：

```text
apps/mcp-server/
  cmd/
    server/
      main.go
  internal/
    config/
    tools/
    resources/
    svc/
  etc/
    mcp-server.yaml
```

## 5. shared 目录建议

## 5.1 shared/proto

放跨服务 proto 定义与生成脚本。

建议：

- 按服务分子目录
- 不把业务实现放进 proto 目录

## 5.2 shared/contracts

放接口契约和 DTO 约定，例如：

- 通用分页结构
- 错误码
- 统一响应模型

## 5.3 shared/events

放领域事件结构和 topic 命名常量。

例如：

- `content.go`
- `review.go`
- `agent.go`

## 5.4 shared/libs

放真正通用、无业务歧义的库，例如：

- 日志
- trace
- 时间工具
- ID 生成
- 密码工具

不要放：

- 具体内容业务规则

## 5.5 shared/configs

放配置结构体和配置加载公共逻辑。

## 5.6 shared/constants

放枚举常量：

- role
- status
- visibility
- ai_access
- relation_type

## 6. sql 目录建议

## 6.1 migrations

用于放正式迁移脚本：

```text
sql/migrations/
  001_users_and_agents.sql
  002_content_core.sql
  003_content_profiles.sql
  ...
```

## 6.2 seeds

用于初始化数据，例如：

- owner 账号初始化说明
- 默认标签
- 默认配置

## 7. scripts 目录建议

建议拆成：

```text
scripts/
  codegen/
  db/
  dev/
```

### codegen

放 `goctl` 相关脚本，例如：

- 生成 API 骨架
- 生成 RPC 骨架
- 生成 model

### db

放数据库迁移、本地初始化脚本。

### dev

放本地开发辅助脚本，例如：

- 一键启动依赖
- 本地检查
- 本地生成代码

## 8. goctl 使用边界

## 8.1 适合用 goctl 的地方

- 新建 API 服务骨架
- 新建 RPC 服务骨架
- 生成 handler / logic / types
- 生成 proto 对应代码
- 生成 model 基础代码

## 8.2 不适合完全依赖 goctl 的地方

- 领域模型设计
- 服务边界设计
- 搜索引擎适配层
- AI provider 适配层
- MCP 工具封装
- 复杂 worker 编排

结论：

`goctl` 是生成器，不是架构师。

## 9. 推荐生成顺序

建议按这个顺序起盘：

1. 先手写目录骨架
2. 再创建 `shared/`、`sql/`、`scripts/`
3. 再为每个服务写 `api` / `proto`
4. 再用 `goctl` 生成服务骨架
5. 最后手工补 `domain/`、`repository/`、`searchengine/` 等目录

## 10. 推荐的 goctl 落地策略

### 第一阶段

建议只对以下服务使用 goctl 生成完整骨架：

- `gateway`
- `identity-service`
- `content-service`
- `review-service`
- `search-service`
- `agent-service`

### 对 worker

建议手写，不强行套 goctl 风格。

原因：

- worker 更偏异步任务执行
- 目录诉求和 API/RPC 服务不同

## 11. 推荐的工程规范

### 11.1 每个服务都必须有

- `cmd/`
- `etc/`
- `internal/`

### 11.2 核心服务建议额外有

- `domain/`
- `repository/`

### 11.3 搜索服务建议额外有

- `searchengine/`

### 11.4 AI 服务建议额外有

- `providers/`
- `prompts/`

## 12. 第一阶段最小仓库骨架

如果要尽快进入开发，建议最小骨架如下：

```text
beehive-blog/
  docs/
  docker/
  apps/
    gateway/
    identity-service/
    content-service/
    search-service/
    indexer-worker/
  shared/
    events/
    libs/
    constants/
  sql/
    migrations/
  scripts/
    codegen/
```

## 13. 当前结论

v2 在 go-zero 下最合理的落地方式是：

**单仓库、多服务、共享契约收敛、API/RPC 用 goctl 生成、领域规则和异步任务逻辑手工组织。**
