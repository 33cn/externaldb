package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleTransfers 查询指定合约的Transfer记录
// GET /api/transfers/{address}?page=1&size=20&from=0x...&to=0x...
func handleTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/transfers/")
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
	fromAddr := r.URL.Query().Get("from")
	toAddr := r.URL.Query().Get("to")

	// 规范化地址参数为小写
	if fromAddr != "" {
		fromAddr = normalizeAddress(fromAddr)
	}
	if toAddr != "" {
		toAddr = normalizeAddress(toAddr)
	}

	// 构建查询（使用LOWER()确保大小写不敏感的比较）
	offset := (page - 1) * size
	query := `SELECT e.tx_hash, e.block_number, e.block_time, e.from_address, e.to_address, 
	          e.value, c.contract_symbol, c.decimals
	          FROM events e
	          LEFT JOIN contracts c ON LOWER(e.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(e.contract_address) = ? AND e.event_name = 'Transfer'`
	args := []interface{}{path}

	if fromAddr != "" {
		query += " AND LOWER(e.from_address) = ?"
		args = append(args, fromAddr)
	}
	if toAddr != "" {
		query += " AND LOWER(e.to_address) = ?"
		args = append(args, toAddr)
	}

	query += " ORDER BY e.block_number DESC, e.log_index DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.GetConn().Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var transfers []TransferRecord
	for rows.Next() {
		var record TransferRecord
		var valueStr string
		err := rows.Scan(
			&record.TxHash,
			&record.BlockNumber,
			&record.BlockTime,
			&record.From,
			&record.To,
			&valueStr,
			&record.TokenSymbol,
			&record.TokenDecimals,
		)
		if err != nil {
			continue
		}

		value, _ := new(big.Int).SetString(valueStr, 10)
		record.Value = valueStr
		record.ValueFormatted = formatTokenAmount(value, record.TokenDecimals)

		transfers = append(transfers, record)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"transfers": transfers,
			"page":      page,
			"size":      size,
		},
	})
}

