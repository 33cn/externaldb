package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleTransactions 查询指定合约的交易记录
// GET /api/transactions/{address}?page=1&size=20&func_name=transfer
func handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "Contract address is required")
		return
	}

	// 规范化合约地址为小写
	path = normalizeAddress(path)

	// 解析查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	funcName := r.URL.Query().Get("func_name")

	// 构建查询（使用LOWER()确保大小写不敏感的比较）
	offset := (page - 1) * size
	query := `SELECT t.tx_hash, t.block_number, t.block_time, t.from_address, t.to_address,
	          t.func_name, t.value, t.gas_used, t.status, c.contract_symbol, c.decimals
	          FROM transactions t
	          LEFT JOIN contracts c ON LOWER(t.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(t.contract_address) = ?`
	args := []interface{}{path}

	if funcName != "" {
		query += " AND t.func_name = ?"
		args = append(args, funcName)
	}

	query += " ORDER BY t.block_number DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.GetConn().Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var transactions []TransactionRecord
	for rows.Next() {
		var record TransactionRecord
		var valueStr *string
		err := rows.Scan(
			&record.TxHash,
			&record.BlockNumber,
			&record.BlockTime,
			&record.From,
			&record.To,
			&record.FuncName,
			&valueStr,
			&record.GasUsed,
			&record.Status,
			&record.TokenSymbol,
			&record.TokenDecimals,
		)
		if err != nil {
			continue
		}

		if valueStr != nil {
			value, _ := new(big.Int).SetString(*valueStr, 10)
			record.Value = *valueStr
			record.ValueFormatted = formatTokenAmount(value, record.TokenDecimals)
		}

		transactions = append(transactions, record)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"transactions": transactions,
			"page":         page,
			"size":         size,
		},
	})
}

