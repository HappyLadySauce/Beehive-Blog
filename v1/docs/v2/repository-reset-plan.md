# Beehive Blog v2 仓库清理与保留方案

## 1. 目标

本次仓库清理的目标不是“删除旧代码”，而是：

- 冻结 v1
- 清理无复用价值的实现
- 保留对 v2 规划有价值的资料
- 为新的微服务仓库结构腾出空间

## 2. 建议保留内容

### 2.1 必须保留

- `docs/` 下所有 v2 文档
- v1 设计和需求文档，作为存档
- `docker/` 目录
- 如有数据库初始化样例、演示数据、接口样例，也建议先归档后再清理

### 2.2 建议归档到 v1 archive

建议后续统一迁移到类似目录：

```text
docs/archive-v1/
```

建议归档内容：

- v1 需求文档
- v1 开发指南
- v1 Hexo 同步设计
- v1 附件、刷新 token 等专题设计
- 旧 API 截图、流程图、页面稿

## 3. 建议清理内容

在完成归档后，可清理：

- v1 Go 业务代码
- v1 Hexo 前端
- v1 React 管理后台
- v1 Swagger 生成产物
- 与 v2 无关的脚本和临时构建产物

## 4. 清理前检查清单

真正执行清理前，建议确认以下事项：

1. v1 文档已归档
2. v1 数据库结构与迁移样例已保留
3. v1 如有测试账号、环境变量样例、部署经验，已转入文档
4. 当前工作区里用户自己的未提交内容已确认是否保留
5. v2 新目录结构已经确定

## 5. v2 推荐新仓库结构

```text
beehive-blog/
  docs/
    archive-v1/
    v2/
  docker/
  deploy/
  apps/
    gateway/
    content-api/
    search-api/
    agent-api/
    mcp-server/
    review-api/
    publish-worker/
    indexer-worker/
    notifier-worker/
  shared/
    contracts/
    proto/
    libs/
    configs/
  scripts/
```

## 6. 清理后的第一批落地文件

仓库清空后，建议最先创建这些内容：

- `README.md`
- `docs/v2/README.md`
- `docs/v2/microservice-blueprint.md`
- `docs/v2/framework-selection.md`
- `docs/v2/v2-domain-model.md`
- `docs/v2/v2-api-design.md`
- `docker/`
- `apps/`
- `shared/`

## 7. 当前建议

如果你准备马上清理仓库，建议顺序是：

1. 先补齐 v2 文档
2. 再把 v1 文档移动到 `docs/archive-v1/`
3. 然后再清理旧代码目录

这样会更稳，不容易把以后要参考的内容删掉。
