package util

import (
	"strings"
)

// AddressConvert 对地址进行转化
// eth 格式地址, 都转化成小写格式
func AddressConvert(address string) string {
	if strings.HasPrefix(address, "0x") {
		return strings.ToLower(address)
	}
	return address
}
