---
name: beehive-api-routes
description: >-
  Standardizes Beehive-Blog Gin HTTP API layering (handler vs business methods),
  ServiceContext injection, v1 request/response DTOs under cmd/app/types, and
  common JSON responses with safe status codes. Use when adding or changing
  routes under cmd/app/routes, handlers, ServiceContext, cmd/app/types, API
  endpoints, login/register patterns, or Gin middleware wiring.
---

# Beehive-Blog 接口开发（路由层）

## 何时使用

在以下场景优先遵循本 Skill：

- 新增或修改 `cmd/app/routes/**` 下的 HTTP 接口
- 编写或调整 `handler.go`、`Init`、中间件挂载
- 定义或修改 `cmd/app/types/api/v1` 中的请求/响应体
- 讨论「接口分层」「统一返回」「ServiceContext」时

## 分层职责

| 层级 | 文件 | 职责 |
|------|------|------|
| HTTP 握手层 | `handler.go` | `ShouldBindJSON`、请求超时、`common.Success` / `common.Fail`、不写业务与数据访问 |
| 业务逻辑层 | `login.go` 等 | 接收 `context.Context`、`*v1.XxxRequest`、`*http.Request`；返回 `(*v1.XxxResponse, int, error)` |
| 路由注册 | 同包 `Init` | `router.V1().Group(...)`、注册路由与中间件 |

业务方法**不得**直接操作 `gin.Context` 写 JSON；由 handler 统一响应。

参考实现：`cmd/app/routes/auth/handler.go`、`cmd/app/routes/auth/login.go`。

## Service 与 ServiceContext

- 每个路由包内定义：`type XxxService struct { svc *svc.ServiceContext }` 与 `NewXxxService(svc *svc.ServiceContext)`.
- 通过 `s.svc` 使用配置与基础设施：`Config`、`DB`、`Redis` 等（见 `cmd/app/svc/serviceContext.go`）。
- **禁止**在 handler 中直接访问 DB/Redis/读配置做业务判断；保持 handler 薄。

## 请求/响应体（DTO）

- 放在 `cmd/app/types/api/v1`，命名 `XxxRequest` / `XxxResponse`。
- 使用 `json` 与 `binding` 标签做校验；规则须与业务语义一致（例如「用户名或邮箱」不要用 `alphanum` 误伤邮箱）。
- 统一响应封装：`cmd/app/types/common/response.go` 的 `Success`、`Fail`。
- **注意**：`Fail` 在 HTTP 状态码 ≥ 500 时会将对外 `message` 统一为 `internal server error`，避免泄漏内部细节；业务层仍应用 `klog` 记录真实错误。

## Handler 编写要点

1. 绑定：`c.ShouldBindJSON(&req)`；失败返回 `400`，`common.Fail(c, http.StatusBadRequest, err)`。
2. 超时：`ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)`（可按接口调整时长）。
3. 调用业务：`resp, statusCode, err := s.BusinessMethod(ctx, &req, c.Request)`。
4. 响应：`err != nil` 时 `common.Fail(c, statusCode, err)`，**不得**写死单一错误码（除非该接口永远只有一种失败语义）；成功用 `common.Success(c, resp)`。

## 业务层编写要点

1. 返回签名：`func (s *XxxService) Method(ctx context.Context, spec *v1.XxxRequest, request *http.Request) (*v1.XxxResponse, int, error)`。
2. `statusCode` 与 `error` 同时表达结果：handler 只信任这一对返回值。
3. 客户端可见错误信息：简短、稳定、英文（与现有 `login` 风格一致），例如 `invalid account or password`、`system error`。
4. 日志：使用 `klog`；可记录 `userID`、业务键、客户端 IP；**禁止**打印密码、明文 token、连接串、完整 SQL。
5. 依赖缺失（如 JWT Secret、Redis 未配置）：返回 `500` + 对外友好文案（如 `auth service unavailable`），详细原因写日志。

## 鉴权与 Redis 快照（与本项目中间件一致）

鉴权中间件会结合 JWT 与 Redis 中 `auth:user:{userId}` 的 `role` / `status` 校验（见 `cmd/app/middlewares/auth.go`）。

凡写入该快照的接口（如登录、注册）应：

- 使用 `HSet` 写入 `role`、`status`；
- 为 key 设置 **TTL**，与访问令牌生命周期对齐：使用 `JWTOptions.ExpireDuration`（`Expire` 在 `HSet` 之后调用）。

避免快照 key 永久存在导致策略与 JWT 过期时间不一致。

## 新增接口自检清单

复制并在 PR/提交前逐项确认：

- [ ] `cmd/app/types/api/v1` 已定义 Request/Response，`binding` 与真实业务一致
- [ ] `handler` 仅：绑定、超时、调 service、`Success`/`Fail`
- [ ] 业务方法返回 `(data, statusCode, err)`，且 handler 透传 `statusCode`
- [ ] 无密码/token/连接串进入响应或不当日志；5xx 不暴露堆栈与 SQL
- [ ] 需要鉴权的路由已在 `Init` 挂载 `middlewares.Auth` 或 `RequireRoles` 等
- [ ] 若写 `auth:user:{id}`，已设置与 access token 一致的 TTL

## 更多模板

见同目录 [reference.md](reference.md)。
