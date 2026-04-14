# Beehive Blog v2 文档索引

当前 v2 文档包括：

- [架构草案](./v2-architecture.md)
- [搜索、索引与 RAG 设计](./v2-search-rag-plan.md)
- [需求分析](./v2-requirements-analysis.md)
- [产品设计](./v2-product-design.md)
- [开发路线规划](./v2-roadmap.md)
- [领域模型设计](./v2-domain-model.md)
- [API 设计草案](./v2-api-design.md)
- [微服务契约设计](./v2-service-contracts.md)
- [数据库初版设计](./v2-database-schema.md)
- [权限矩阵设计](./v2-permission-matrix.md)
- [事件流与 MCP 设计](./v2-event-and-mcp-design.md)
- [go-zero 项目布局设计](./v2-gozero-project-layout.md)
- [部署拓扑设计](./v2-deployment-topology.md)
- [API DTO 与错误码规范](./v2-api-dto-and-error-codes.md)
- [迁移计划](./v2-migration-plan.md)
- [服务启动顺序与落地步骤](./v2-service-bootstrap-order.md)


建议接下来继续补充：

- `v2-first-contract-drafts.md`
- `v2-local-dev-setup.md`

当前文档推进顺序建议：

1. 先固化领域模型、数据库和服务边界
2. 再细化权限矩阵和 API DTO
3. 最后进入 go-zero 工程骨架搭建
