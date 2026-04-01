# 模板与参考路径

本文件为 [SKILL.md](SKILL.md) 的补充，按需阅读以节省主 Skill 篇幅。

## 参考实现（本项目）

- HTTP 层：`cmd/app/routes/auth/handler.go`
- 业务层：`cmd/app/routes/auth/login.go`
- 统一响应：`cmd/app/types/common/response.go`
- 服务上下文：`cmd/app/svc/serviceContext.go`
- 鉴权中间件：`cmd/app/middlewares/auth.go`
- Swagger 路由与示例注释：`cmd/app/router/router.go`

## Swagger 注释模板（建议写在 handler 方法上方）

```go
// Login godoc
//
//	@Summary		用户登录
//	@Description	用户名或邮箱登录并返回 token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LoginRequest	true	"登录参数"
//	@Success		200		{object}	common.BaseResponse
//	@Failure		400		{object}	common.BaseResponse
//	@Failure		401		{object}	common.BaseResponse
//	@Failure		500		{object}	common.BaseResponse
//	@Router			/api/v1/auth/login [post]
func (s *AuthService) handleLogin(c *gin.Context) {}
```

## Swagger 维护清单

- 路由改动后，先对照 `Init` 中真实注册路径，再更新 `@Router`。
- `@Param` 使用 `cmd/app/types/api/v1` 下请求体类型，避免文档模型漂移。
- `@Success` / `@Failure` 使用 `common.BaseResponse`，并确保状态码与业务层返回一致。
- 确认文档访问入口 `/swagger/index.html` 可查看到最新接口项。

## Handler 骨架（伪代码）

```go
func (s *SomeService) handleXxx(c *gin.Context) {
	var req v1.XxxRequest
	// ShouldBindJSON 会自动按 DTO 上的 binding 标签执行参数校验
	// 例如：json:"account" binding:"required,min=3,max=50"
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.ErrorS(err, "Could not read request")
		common.Fail(c, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode, err := s.Xxx(ctx, &req, c.Request)
	if err != nil {
		common.Fail(c, statusCode, err)
		return
	}
	common.Success(c, resp)
}
```

## 业务方法骨架（伪代码）

```go
func (s *SomeService) Xxx(ctx context.Context, spec *v1.XxxRequest, request *http.Request) (*v1.XxxResponse, int, error) {
	if spec == nil {
		return nil, http.StatusBadRequest, errors.New("invalid request")
	}
	// 二次校验示例（业务层必做）：
	// 1) 跨字段规则：如 StartAt <= EndAt
	// 2) 数据一致性：唯一性/存在性检查
	// 3) 权限与状态：当前用户是否有权限、资源是否允许当前操作
	// 使用 s.svc.DB / s.svc.Redis / s.svc.Config
	// 返回 (*v1.XxxResponse, http.StatusOK, nil) 或 (nil, statusCode, err)
	return &v1.XxxResponse{}, http.StatusOK, nil
}
```

## 路由注册骨架（伪代码）

```go
func Init(svcCtx *svc.ServiceContext) {
	r := router.V1().Group("some")
	svc := NewSomeService(svcCtx)
	r.POST("/xxx", svc.handleXxx)
}
```

## JSON 响应形状

成功时：`common.Success` → HTTP 200，`data` 为业务体。

失败时：`common.Fail` → HTTP 状态码与 `code` 字段一致；`message` 为错误说明（5xx 时对外统一为 `internal server error`）。
