# Beehive-Blog v1：登录与注册约定

本文档描述 **HTTP API v1**（`/api/v1`）下注册、登录、会话探测、令牌刷新与登出的**契约与行为**，供产品、后端与前端（含 BFF）对齐。实现变更时须同步更新本文档。

**延伸阅读**：[产品设计原则](../product-principles.md)、[前端 SSR/SEO 与 BFF 约定](../frontend/react-ssr-seo-architecture.md)。

---

## 1. 端点一览

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/users/register` | 无（限流见 §2） | 本地账号注册；成功后签发 **access + refresh**（自动登录） |
| `GET` | `/api/v1/auth/github/authorize` | 无（限流见 §2） | 启动 GitHub OAuth：返回 `state` 与 `auth_url` |
| `POST` | `/api/v1/auth/login` | 无（限流见 §2） | `grant_type=local` 或 `github_oauth2` |
| `POST` | `/api/v1/auth/refresh` | 无（凭 `refresh_token` 请求体；限流见 §2） | 轮换 refresh 会话并签发新令牌对 |
| `GET` | `/api/v1/auth/session` | **Bearer（access）** | 校验 access，返回 `uid`、`role`、`exp`、`sid`（供 BFF / 客户端守卫） |
| `POST` | `/api/v1/auth/logout` | **Bearer（access）** | 吊销当前 access 绑定的服务端会话 |

`GET /auth/session` 与 `POST /auth/logout` 走 **AuthMiddleware**，**不在** §2 所列公共限流组内。

---

## 2. 限流（按客户端 IP）

实现：`cmd/app/middleware/ratelimit.go` 令牌桶（可持续速率 + 突发容量）。

| 路由 | 可持续速率 | 突发 |
|------|------------|------|
| `POST /users/register` | 约 **10 次/分钟**（`10/60` 事件/秒） | **12** |
| `GET /auth/github/authorize`、`POST /auth/login`、`POST /auth/refresh` | 约 **20 次/分钟**（`20/60` 事件/秒） | **25** |

---

## 3. 注册（本地）

- **路径**：`POST /api/v1/users/register`
- **请求体**：`RegisterRequest`（`cmd/app/types/api/v1/users.go`）

### 3.1 字段校验（Gin `binding`）

| 字段 | 规则 |
|------|------|
| `username` | 必填，最长 **64** |
| `password` | 必填，长度 **8–72**（bcrypt 上限 72 字节） |
| `email` | 可选；若填写须为合法 `email` 格式，最长 **320** |
| `nickname` | 可选，最长 **128** |
| `phone` | 可选，最长 **16** |

### 3.2 业务规则

- **用户名**：在「活跃行」中唯一；冲突 → **409**，文案如 `username is already taken`。
- **邮箱**：若提供，在活跃行中唯一；冲突 → **409**，如 `email is already registered`。
- **密码存储**：bcrypt，cost **12**（`pkg/auth/passwd`）。
- **新建用户默认值**：`role = "member"`，`status = "active"`。
- **事务**：同一事务内创建 `identity.users`、`UserCredential`（密码哈希），并签发 **access + refresh** JWT。
- **头像**：注册接口**不**绑定头像；附件归属在登录态后续流程中处理。

---

## 4. 登录

- **路径**：`POST /api/v1/auth/login`
- **请求体**：`LoginRequest`（`cmd/app/types/api/v1/auth.go`）
- **`grant_type`**：必填；仅支持 `local` 与 `github_oauth2`，否则 **400** `unsupported grant_type`。

### 4.1 `grant_type = local`（用户名 / 邮箱 + 密码）

- **必填**：`account`、`password`（处理器层校验；空则 **400**）。
- **`account`**：可为 **用户名或邮箱**（`WHERE username = ? OR email = ?`）。
- **安全**：用户不存在、无凭证、密码错误均返回 **401** `invalid credentials`（统一文案，降低用户枚举）。
- **成功后**：更新 `last_login_at`（失败仅记日志，不阻断登录），签发令牌对。

### 4.2 `grant_type = github_oauth2`

1. 客户端先调 `GET /api/v1/auth/github/authorize`，取得 **`state`**（Redis 一次性，TTL **15 分钟**）与 **`auth_url`**（scope：`read:user user:email`）。
2. 用户授权后，用回调中的 **`code`** 与**同一 `state`** 调 `POST /auth/login`。
3. **`state`**：须在 Redis 中存在且**一次性消费**；缺失 / 过期 / 无效 → **401** `invalid or expired oauth session`。
4. **`code`**：空 → **400**；与 GitHub 换 token 失败 → **401** 等。
5. 拉取 GitHub 用户与**主验证邮箱**逻辑见 `pkg/auth/oauth/github.go`（优先 primary+verified，否则 verified / primary；无可用邮箱则无法创建新用户）。
6. **用户解析**：先按 `provider=github` + `provider_subject`（GitHub 数字 ID）查找；否则按 **email** 查找已有用户并**绑定** GitHub 身份；否则用 `login` 为 username 创建用户（用户名冲突时最多重试 5 次，追加 `_` + 随机四位后缀）。

### 4.3 可登录账户状态（所有登录路径）

`assertUserMayLogin`：**仅** `status` 为 `active` 或 `pending` 的用户允许登录；否则 **403** `account is not allowed to login`。

---

## 5. 会话探测

- **路径**：`GET /api/v1/auth/session`
- **鉴权**：`Authorization: Bearer <access_token>`；由 **AuthMiddleware** 校验 JWT。
- **成功**：`200`，`data` 为 `AuthSessionResponse`：`uid`、`role`、`exp`（access 过期 Unix 秒）、`sid`（服务端会话 ID，可 `omitempty` 视实现）。
- **失败**：claims 无效或缺失关键字段 → **401** `invalid or expired access token`。

典型用途：BFF 或客户端用当前 access 换取**已签名的会话声明**，用于路由守卫；**不**替代 refresh 轮换逻辑。

---

## 6. 刷新令牌

- **路径**：`POST /api/v1/auth/refresh`
- **请求体**：`refresh_token` 必填。
- **行为**：解析 refresh JWT → 行级锁读取 `UserSession` → 校验未轮换、未吊销、未过期、JTI 与 refresh 哈希一致 → 再次 `assertUserMayLogin` → 轮换会话并返回新令牌对。
- **重用检测**：若会话已标记 `RotatedAt`（旧 refresh 再用）或哈希不匹配，会吊销会话族相关逻辑并 **401**（与过期统一对外文案）。
- **用户不存在**：→ **401**。

---

## 7. 登出

- **路径**：`POST /api/v1/auth/logout`
- **鉴权**：`Authorization: Bearer <access_token>`。
- **行为**：从 access claims 取 `SID` / `UID`，吊销该会话（`reason=logout`）；claims 无效 → **401**。重复调用在实现上按幂等安全场景处理。

---

## 8. 响应中的令牌形态

成功注册 / 登录 / 刷新均返回 `AuthToken`（或包在 `token` 字段内，以具体响应 DTO 为准）：

- `access_token`、`token_type`（如 `Bearer`）、`expires_in`（秒）
- `refresh_token`：有则返回；序列化标签 `omitempty` 时未签发可不输出

---

## 9. 与浏览器 / BFF 的协同约定

- **v1 API 契约不变**：注册 / 登录 / 刷新仍通过 JSON 返回令牌对；Go 不强制 Cookie 形态。
- **浏览器推荐形态**：由 **Next Route Handlers（BFF）** 将 refresh 写入 **HttpOnly** Cookie，业务 JS **不**长期持有 refresh；access 用于调用 `GET /auth/session` 或受 proxied 的 Go 接口。详见 [react-ssr-seo-architecture.md](../frontend/react-ssr-seo-architecture.md) §7、§11。
- **限流与 UX**：登录 / 注册 / 刷新路径须 **防连点**、友好展示 **429**，与 §2 一致。

---

## 10. 代码索引（便于审计）

| 主题 | 主要文件 |
|------|----------|
| 注册逻辑 | `cmd/app/routes/users/register.go` |
| 注册 DTO | `cmd/app/types/api/v1/users.go` |
| 登录 / OAuth / 状态门控 | `cmd/app/routes/auth/login.go` |
| 登录 DTO | `cmd/app/types/api/v1/auth.go` |
| GitHub OAuth 开始 | `cmd/app/routes/auth/github_oauth_begin.go` |
| 会话 | `cmd/app/routes/auth/session.go` |
| OAuth state / 用户查找 | `pkg/auth/oauth/github_oauth_state.go`、`pkg/auth/oauth/github.go` |
| 密码哈希 | `pkg/auth/passwd/passwd.go` |
| 刷新 / 登出 | `cmd/app/routes/auth/refresh.go`、`cmd/app/routes/auth/logout.go` |
| 路由与限流 | `cmd/app/routes/users/handler.go`、`cmd/app/routes/auth/handler.go` |

---

*若与 `docs/product-principles.md` 或其它 `docs/` 文档冲突，以实现或产品负责人裁定为准。*
