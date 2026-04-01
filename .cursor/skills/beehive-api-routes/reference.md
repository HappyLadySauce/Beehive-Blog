# 模板与参考路径

本文件为 [SKILL.md](SKILL.md) 的补充，按需阅读以节省主 Skill 篇幅。

## 参考实现（本项目）

- HTTP 层：`cmd/app/routes/auth/handler.go`
- 业务层：`cmd/app/routes/auth/login.go`
- 统一响应：`cmd/app/types/common/response.go`
- 服务上下文：`cmd/app/svc/serviceContext.go`
- 鉴权中间件：`cmd/app/middlewares/auth.go`

## Handler 骨架（伪代码）

```go
func (s *SomeService) handleXxx(c *gin.Context) {
	var req v1.XxxRequest
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
