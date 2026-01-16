package main

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
)

// writeJSON 写入JSON响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
	})
}

// normalizeAddress 规范化地址，统一转换为小写
// 以太坊地址是大小写不敏感的，统一转换为小写便于比较和查询
func normalizeAddress(address string) string {
	if address == "" {
		return address
	}
	// 保持0x前缀，将后面的字符转换为小写
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		return "0x" + strings.ToLower(address[2:])
	}
	return strings.ToLower(address)
}

// formatTokenAmount 格式化代币金额
func formatTokenAmount(amount *big.Int, decimals uint8) string {
	if amount == nil {
		return "0"
	}
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	quotient := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)

	if remainder.Sign() == 0 {
		return quotient.String()
	}

	// 处理小数部分
	remainderFloat := new(big.Float).Quo(new(big.Float).SetInt(remainder), new(big.Float).SetInt(divisor))
	remainderStr := remainderFloat.Text('f', int(decimals))
	remainderStr = strings.TrimRight(strings.TrimRight(remainderStr, "0"), ".")

	if remainderStr == "" || remainderStr == "0" {
		return quotient.String()
	}

	return quotient.String() + "." + remainderStr
}

