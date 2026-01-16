package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleHolders 查询指定合约的Holder信息
// GET /api/holders/{address}?page=1&size=20&min_balance=0
func handleHolders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/holders/")
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
	minBalanceStr := r.URL.Query().Get("min_balance")

	// 构建查询
	offset := (page - 1) * size
	query := `SELECT tb.address, tb.balance, tb.last_tx_hash, tb.last_tx_block_number, 
	          tb.last_updated_at, c.decimals
	          FROM token_balances tb
	          LEFT JOIN contracts c ON tb.contract_address = c.contract_address
	          WHERE tb.contract_address = ?`
	args := []interface{}{path}

	if minBalanceStr != "" {
		query += " AND tb.balance >= ?"
		args = append(args, minBalanceStr)
	}

	query += " ORDER BY tb.balance DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.GetConn().Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var holders []HolderInfo
	for rows.Next() {
		var holder HolderInfo
		var balanceStr string
		var decimals uint8
		err := rows.Scan(
			&holder.Address,
			&balanceStr,
			&holder.LastTxHash,
			&holder.LastTxBlock,
			&holder.LastUpdated,
			&decimals,
		)
		if err != nil {
			continue
		}

		balance, _ := new(big.Int).SetString(balanceStr, 10)
		holder.Balance = balanceStr
		holder.BalanceFormatted = formatTokenAmount(balance, decimals)

		holders = append(holders, holder)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"holders": holders,
			"page":    page,
			"size":    size,
		},
	})
}

