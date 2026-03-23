package code

// 用户相关错误码常量（User 服务 2xxx）
// 后三位：0=成功，1xx=请求/参数，2xx=认证/授权，3xx=业务逻辑，5xx=服务端
const (
	// 用户未授权
	CodeUserNotAuthorized = 2300
	// 用户已存在
	CodeUserAlreadyExists = 2301
	// 用户不存在
	CodeUserNotFound = 2400
)
