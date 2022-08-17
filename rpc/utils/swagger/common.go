package swagger

// Health 服务健康状态
type Health struct {
	Status  string `json:"status" label:"状态"` // 状态
	Version string `json:"version"`           // 版本
}

type ClientRequest struct {
	Method string      `json:"method,omitempty"` // 方法
	Params interface{} `json:"params"`           // 参数
	ID     uint64      `json:"id"`               // 请求标识
}

// ClientRequestNil swagger:parameters
type ClientRequestNil struct {
	Method string `json:"method" label:"方法"` // 方法
	ID     uint64 `json:"id"`                // 请求标识
}

type ServerResponse struct {
	ID     uint64      `json:"id"`     // 请求标识
	Result interface{} `json:"result"` // 返回结果
	Error  interface{} `json:"error"`  // 错误描述
}
