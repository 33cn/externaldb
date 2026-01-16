package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// handleParseTx 解析EVM交易
// POST /api/parse_tx
// Request: {"tx_hash": "0x..."}
// Response: EvmTxInfo
func handleParseTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		TxHash string `json:"tx_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.TxHash == "" {
		writeError(w, http.StatusBadRequest, "tx_hash is required")
		return
	}

	// 从Chain33节点获取交易详情
	detail, err := getTxDetailFromChain33(globalChainGRPC, req.TxHash)
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

