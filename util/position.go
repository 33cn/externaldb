package util

import "fmt"

// PositionID 日志用， 定位问题
func PositionID(exec string, height, index int64) string {
	return fmt.Sprintf("%s:%d.%d", exec, height, index)
}
