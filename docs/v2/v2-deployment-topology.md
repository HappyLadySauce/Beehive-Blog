# Beehive Blog v2 部署拓扑设计

## 1. 目标

本文件定义 v2 第一阶段的部署拓扑、环境划分、基础组件关系，以及本地开发和 Docker 部署的建议结构。

目标：

- 让服务拓扑和微服务设计一致
- 让本地开发环境可快速启动
- 为后续 staging / production 部署预留一致模型

## 2. 第一阶段部署原则

第一阶段不追求复杂集群，追求：

- 可本地开发
- 可 Docker 化
- 能支撑多服务联调
- 搜索、Redis、PostgreSQL 都能独立运行

## 3. 第一阶段基础组件

建议第一阶段最少组件如下：

- `gateway`
- `identity-service`
- `content-service`
- `search-service`
- `review-service`
- `agent-service`
- `indexer-worker`
- `postgresql`
- `redis`
- `meilisearch` 或 `elasticsearch`

说明：

- 如果第一阶段先选 `Meilisearch`，部署复杂度最低
- `Elasticsearch` 可以作为后续增强方案，不一定第一天就上

## 4. 本地开发拓扑

```text
Web Public / Studio
    |
    v
  gateway
   |   |   |   |   |
   v   v   v   v   v
identity content search review agent
              ^
              |
         indexer-worker

postgresql <-> content / review / identity / agent / search
redis      <-> gateway / identity / search / agent / worker
meilisearch or elasticsearch <-> search / indexer-worker
```

## 5. 服务职责与基础设施连接

## 5.1 gateway

依赖：

- `identity-service`
- `content-service`
- `search-service`
- `review-service`
- `agent-service`
- `redis` 可选

## 5.2 identity-service

依赖：

- `postgresql`
- `redis`

## 5.3 content-service

依赖：

- `postgresql`
- `redis` 可选

## 5.4 review-service

依赖：

- `postgresql`

## 5.5 search-service

依赖：

- `postgresql`
- `redis`
- `meilisearch` 或 `elasticsearch`

## 5.6 agent-service

依赖：

- `postgresql`
- `redis`
- `search-service`
- 后续模型服务或 LLM provider

## 5.7 indexer-worker

依赖：

- `postgresql`
- `redis` 或事件表
- `meilisearch` 或 `elasticsearch`

## 6. 环境划分

第一阶段建议至少定义 3 套环境：

- `local`
- `staging`
- `production`

## 6.1 local

特点：

- 本机开发
- Docker Compose 起依赖
- 服务可本机进程运行，也可部分容器化

## 6.2 staging

特点：

- 供联调和预发布验证
- 与 production 尽量接近

## 6.3 production

特点：

- 正式服务
- 强制开启更严格的配置、日志、备份和告警

## 7. 本地开发建议

## 7.1 推荐模式

建议采用：

- 基础依赖走 Docker
- Go 服务本机运行

原因：

- 启动速度更快
- 调试更方便
- 依赖环境一致

## 7.2 本地依赖建议

建议 Docker 启动：

- `postgresql`
- `redis`
- `meilisearch`

如后续选择 Elasticsearch，再切换：

- `elasticsearch`

## 8. Docker Compose 结构建议

建议在：

- `docker/local/`

维护本地联调编排。

推荐结构：

```text
docker/
  local/
    docker-compose.yaml
    env/
  staging/
    docker-compose.yaml
  production/
    compose.example.yaml
```

## 9. 网络拓扑建议

第一阶段所有服务位于同一私有网络即可。

对外暴露：

- `gateway`
- `web-public`
- `web-studio`

不对外暴露：

- `identity-service`
- `content-service`
- `search-service`
- `review-service`
- `agent-service`
- `indexer-worker`
- `postgresql`
- `redis`
- `meilisearch/elasticsearch`

## 10. 配置管理建议

每个服务建议保留独立配置文件：

- `etc/<service>.yaml`

同时把共享环境变量收敛到：

- `deploy/local/.env`
- `deploy/staging/.env`
- `deploy/production/.env`

建议配置类型：

- 服务端口
- PostgreSQL DSN
- Redis 地址
- 搜索引擎地址
- JWT 配置
- Trace 配置
- 外部模型服务配置

## 11. 数据持久化建议

必须持久化：

- PostgreSQL 数据目录
- Redis 数据目录
- 搜索引擎数据目录
- 附件存储目录

建议挂载目录：

- `data/postgresql/`
- `data/redis/`
- `data/meilisearch/`
- `data/elasticsearch/`
- `data/uploads/`

## 12. 搜索引擎部署建议

## 12.1 第一阶段推荐

优先建议：

- `Meilisearch`

原因：

- 启动简单
- 占用更低
- 适合早期知识库搜索

## 12.2 后续增强

当搜索复杂度提升时，再考虑：

- `Elasticsearch`

适合场景：

- 更复杂的聚合
- 更复杂排序
- 更大规模索引

## 13. 事件基础设施建议

第一阶段建议优先：

- 数据库事件表
  或
- `Redis Stream`

不建议第一阶段就强依赖：

- Kafka
- NATS JetStream

原因：

- 初期服务还不多
- 先验证业务链路更重要

## 14. 可观测性建议

第一阶段建议至少具备：

- 结构化日志
- trace_id
- 请求耗时
- 基础错误统计

第二阶段再补：

- Prometheus
- Grafana
- Loki / ELK

## 15. 生产部署方向

第一阶段可以接受：

- Docker Compose

第二阶段再考虑：

- Kubernetes

条件是：

- 服务数量增加
- 环境数量增加
- 发布与扩缩容复杂度提高

## 16. 当前结论

v2 第一阶段部署应以：

**Docker 化基础依赖 + go-zero 多服务本机/容器联调 + Meilisearch 优先 + PostgreSQL/Redis 持久化**

为核心策略。
