# Beehive-Blog v1：登录与注册规则（实现快照）

本文档根据当前仓库实现整理，描述 **HTTP API v1**（`/api/v1`）下的注册、登录、令牌刷新与登出行为，供产品与前后端对齐。若实现变更，请同步更新本文档。

---

## 1. 端点一览

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/users/register` | 无 | 本地账号注册（自动登录，返回令牌对） |
| `GET` | `/api/v1/auth/github/authorize` | 无 | 启动 GitHub OAuth：返回 `state` 与 `auth_url` |
| `POST` | `/api/v1/auth/login` | 无 | 本地密码或 GitHub 授权码登录 |
| `POST` | `/api/v1/auth/refresh` | 无（凭 `refresh_token`） | 轮换 refresh 会话并签发新令牌对 |
| `POST` | `/api/v1/auth/logout` | Bearer（access） | 吊销当前 access 绑定的服务端会话 |

---

## 2. 限流（按客户端 IP）

实现：`cmd/app/middleware/ratelimit.go` 的令牌桶（可持续速率 + 突发容量）。

| 路由组 | 可持续速率 | 突发 |
|--------|------------|------|
| `POST /users/register` | 约 **10 次/分钟**（`10/60` 事件/秒） | **12** |
| `GET /auth/github/authorize`、`POST /auth/login`、`POST /auth/refresh` | 约 **20 次/分钟**（`20/60` 事件/秒） | **25** |

`POST /auth/logout` 走鉴权中间件，**不在**上述公共限流组内。

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

- **用户名**：在「活跃行」中唯一；冲突时 **409**，文案如 `username is already taken`。
- **邮箱**：若提供，在活跃行中唯一；冲突时 **409**，如 `email is already registered`。
- **密码存储**：bcrypt，cost **12**（`pkg/auth/passwd`）。
- **新建用户默认值**：`role = "member"`，`status = "active"`。
- **事务**：同一事务内创建 `identity.users`、`UserCredential`（密码哈希），并签发 **access + refresh** JWT（自动登录）。
- **头像**：注册接口**不**绑定头像；注释说明附件无归属列，需在登录态流程后再绑定。

---

## 4. 登录

- **路径**：`POST /api/v1/auth/login`
- **请求体**：`LoginRequest`（`cmd/app/types/api/v1/auth.go`）
- **`grant_type`**：必填；仅支持 `local` 与 `github_oauth2`，否则 **400** `unsupported grant_type`。

### 4.1 `grant_type = local`（用户名/邮箱 + 密码）

- **必填**：`account`、`password`（处理器层校验；空则 **400**）。
- **`account`**：可为 **用户名或邮箱**（`WHERE username = ? OR email = ?`）。
- **安全**：用户不存在、无凭证、密码错误均返回 **401** `invalid credentials`（统一文案，降低用户枚举风险）。
- **成功后**：更新 `last_login_at`（失败仅记日志，不阻断登录），签发令牌对。

### 4.2 `grant_type = github_oauth2`

1. 客户端先调 `GET /api/v1/auth/github/authorize`，取得 **`state`**（Redis 一次性，TTL **15 分钟**）与 **`auth_url`**（scope：`read:user user:email`）。
2. 用户授权后，用回调中的 **`code`** 与**同一 `state`** 调 `POST /login`。
3. **`state`**：须在 Redis 中存在且**一次性消费**；缺失/过期/无效 → **401** `invalid or expired oauth session`。
4. **`code`**：空 → **400**；与 GitHub 换 token 失败 → **401** 等。
5. 拉取 GitHub 用户与**主验证邮箱**逻辑见 `pkg/auth/oauth/github.go`（优先 primary+verified，否则 verified / primary；无可用邮箱则无法创建新用户）。
6. **用户解析**：先按 `provider=github` + `provider_subject`（GitHub 数字 ID）查找；否则按 **email** 查找已有用户并**绑定** GitHub 身份；否则用 `login` 为 username 创建用户（用户名冲突时最多重试 5 次，追加 `_随机四位后缀`）。

### 4.3 可登录账户状态（所有登录路径）

`assertUserMayLogin`：**仅** `status` 为 `active` 或 `pending` 的用户允许登录；否则 **403** `account is not allowed to login`。

---

## 5. 刷新令牌

- **路径**：`POST /api/v1/auth/refresh`
- **请求体**：`refresh_token` 必填。
- **行为**：解析 refresh JWT → 行级锁读取 `UserSession` → 校验未轮换、未吊销、未过期、JTI 与 refresh 哈希一致 → **再次** `assertUserMayLogin` → 轮换会话并返回新令牌对。
- **重用检测**：若会话已标记 `RotatedAt`（旧 refresh 再用）或哈希不匹配，会吊销会话族相关逻辑并 **401**（与过期统一对外文案）。
- **用户删除**：用户不存在 → **401**。

---

## 6. 登出

- **路径**：`POST /api/v1/auth/logout`
- **鉴权**：`Authorization: Bearer <access_token>`。
- **行为**：从 access claims 取 `SID`/`UID`，吊销该会话（`reason=logout`）；claims 无效 → **401**。重复调用视为幂等安全场景由实现保证。

---

## 7. 响应中的令牌形态

成功注册/登录/刷新均返回 `AuthToken`：

- `access_token`、`token_type`（如 `Bearer`）、`expires_in`（秒）
- `refresh_token`（有则返回；省略策略以序列化标签为准）

---

## 8. 与代码的对应关系（便于审计）

| 主题 | 主要文件 |
|------|----------|
| 注册逻辑 | `cmd/app/routes/users/register.go` |
| 注册 DTO | `cmd/app/types/api/v1/users.go` |
| 登录 / OAuth / 状态门控 | `cmd/app/routes/auth/login.go` |
| 登录 DTO | `cmd/app/types/api/v1/auth.go` |
| GitHub OAuth 开始 | `cmd/app/routes/auth/github_oauth_begin.go` |
| OAuth state / 用户查找 | `pkg/auth/oauth/github_oauth_state.go`、`pkg/auth/oauth/github.go` |
| 密码哈希 | `pkg/auth/passwd/passwd.go` |
| 刷新 / 登出 | `cmd/app/routes/auth/refresh.go`、`cmd/app/routes/auth/logout.go` |
| 路由与限流 | `cmd/app/routes/users/handler.go`、`cmd/app/routes/auth/handler.go` |

---

*文档生成说明：由仓库当前实现反推，非独立产品规格书。若与 `docs/product-principles.md` 或其它版本文档冲突，以实现或产品负责人裁定为准。*
