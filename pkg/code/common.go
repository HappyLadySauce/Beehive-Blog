package code

// 业务响应/错误码常量（通用服务 1xxx）
// 后三位：0=成功，1xx=请求/参数，2xx=认证/授权，3xx=业务逻辑，5xx=服务端
const (
	// 成功
	CodeSuccess = 200
	// 请求/参数错误
	CodeInvalidParam = 300
	// 解析失败
	CodeUnmarshalFailed = 302
	// Marshal 失败
	CodeMarshalFailed = 301
	// 服务端错误
	CodeInternal = 500
)

// 数据库相关错误
const (
	CodeDBQueryFailed = 321
	// 数据库添加失败
	CodeDBAddFailed = 322
	// 数据库删除失败
	CodeDBDeleteFailed = 324
)

// 缓存相关错误
const (
	// 缓存设置失败
	CodeCacheSetFailed = 331
	// 缓存获取失败
	CodeCacheGetFailed = 332
	// 缓存删除失败
	CodeCacheDeleteFailed = 333
)

// 分布式锁相关错误
const (
	// 分布式锁获取失败
	CodeDistributedLockGetFailed = 343
	// 分布式锁释放失败
	CodeDistributedLockUnlockFailed = 344
	// 分布式锁已存在
	CodeDistributedLockAlreadyExists = 341
	// 分布式锁获取超时
	CodeDistributedLockGetTimeout = 342
)

// 消息队列相关错误
const (
	// 消息队列声明失败
	CodeMQDeclareFailed   = 351
	// 消息队列绑定失败
	CodeMQBindQueueFailed = 352
	// 消息队列消费失败
	CodeMQConsumeFailed   = 355
	// 消息队列推送失败
	CodeMQPushFailed      = 353
	// 消息队列设置QoS失败
	CodeMQSetQosFailed    = 354
)

// WebSocket 连接相关错误
const (
	// 连接不存在或已关闭
	CodeConnNotFound = 361
	// 连接写入失败
	CodeConnWriteFailed = 362
)
