package util

import "runtime"

func GetStack() string {
	buf := make([]byte, 1<<12)
	runtime.Stack(buf, false)
	return string(buf)
}
