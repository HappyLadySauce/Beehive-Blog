package ws

// 定义 Envelope 为 WebSocket 提供统一 JSON 消息格式
type Envelope struct {
	Type	string `json:"type"`
	Tid 	string `json:"tid,omitempty"`
	PayLoad interface{} `json:"payload,omitempty"`
	Error 	*Error `json:"error,omitempty"`
}

// 定义错误结构的结构体
type Error struct {
	Code 	int `json:"code"`
	Message string `json:"message"`
}