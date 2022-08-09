package health

import (
	"github.com/33cn/externaldb/version"
)

// Health 服务健康状态
type Health struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// GetHealth 获取服务运行状态和版本
func GetHealth() *Health {
	return &Health{Status: "UP", Version: version.GetVersion()}
}
