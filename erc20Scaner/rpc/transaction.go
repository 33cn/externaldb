package main

import (
	"fmt"
	"net/http"
	"strings"
)

// handleTransactionsRouter 路由分发函数，处理 /api/transactions 路径
func handleTransactionsRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 移除前缀 /api/transactions
	path := strings.TrimPrefix(r.URL.Path, "/api/transactions")

	// 移除开头的 /
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "Transaction hash is required")
		return
	}

	parts := strings.Split(path, "/")

	// /api/transactions/{tx_hash}/analysis
	if len(parts) == 2 && parts[1] == "analysis" {
		handleTransactionAnalysis(w, r, parts[0])
		return
	}

	// /api/transactions/{tx_hash} - 如果需要查询交易详情，可以在这里实现
	if len(parts) == 1 {
		// 目前暂不实现，返回提示
		writeError(w, http.StatusNotImplemented, "Transaction detail query is not implemented yet")
		return
	}

	writeError(w, http.StatusNotFound, "Invalid path")
}

// handleTransactionAnalysis 解析并获取单笔 EVM 交易的详细动作
// GET /api/transactions/{tx_hash}/analysis
func handleTransactionAnalysis(w http.ResponseWriter, r *http.Request, txHash string) {
	// 验证交易哈希格式
	if !strings.HasPrefix(strings.ToLower(txHash), "0x") || len(txHash) != 66 {
		writeError(w, http.StatusBadRequest, "Invalid tx_hash format")
		return
	}

	// 从Chain33节点获取交易详情
	detail, err := getTxDetailFromChain33(globalChainGRPC, txHash)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Transaction not found: %v", err))
		return
	}

	// 解析EVM交易
	parsed := parseEvmTx(detail, getAbiFromES, globalChainSymbol)

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data:    parsed,
	})
}
