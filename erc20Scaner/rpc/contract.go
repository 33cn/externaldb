package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleContractAddressTransfers 查询某个合约中指定地址的Transfer事件记录
// GET /api/contract/{contractAddress}/address/{address}/transfers?page=1&size=20&role=from|to|both
func handleContractAddressTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析路径：/api/contract/{contractAddress}/address/{address}/transfers 或 /transactions
	path := strings.TrimPrefix(r.URL.Path, "/api/contract/")
	parts := strings.Split(path, "/")

	// 检查是否是 address 相关的路径
	if len(parts) < 4 || parts[1] != "address" {
		// 不是这个接口，可能是 handleContractDetail
		handleContractDetail(w, r)
		return
	}

	// 判断是 transfers 还是 transactions
	if parts[3] == "transfers" {
		handleContractAddressTransfersImpl(w, r, parts)
	} else if parts[3] == "transactions" {
		handleContractAddressTransactions(w, r, parts)
	} else {
		// 不是这个接口，可能是 handleContractDetail
		handleContractDetail(w, r)
	}
}

// handleContractAddressTransfersImpl 实现查询某个合约中指定地址的Transfer事件记录
func handleContractAddressTransfersImpl(w http.ResponseWriter, r *http.Request, parts []string) {
	contractAddress := parts[0]
	address := parts[2]

	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}
	if !strings.HasPrefix(strings.ToLower(address), "0x") || len(address) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid address format")
		return
	}

	// 规范化地址为小写
	contractAddress = normalizeAddress(contractAddress)
	address = normalizeAddress(address)

	// 解析查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	role := strings.ToLower(r.URL.Query().Get("role"))
	if role != "from" && role != "to" && role != "" {
		writeError(w, http.StatusBadRequest, "Invalid role parameter, must be 'from', 'to', or empty (both)")
		return
	}
	if role == "" {
		role = "both" // 默认查询两种
	}

	// 构建查询（使用LOWER()确保大小写不敏感的比较）
	offset := (page - 1) * size
	query := `SELECT e.tx_hash, e.block_number, e.block_time, e.from_address, e.to_address, 
	          e.value, c.contract_symbol, c.decimals
	          FROM events e
	          LEFT JOIN contracts c ON LOWER(e.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(e.contract_address) = ? AND e.event_name = 'Transfer'`
	args := []interface{}{contractAddress}

	// 根据role参数添加地址筛选条件（使用LOWER()确保大小写不敏感）
	switch role {
	case "from":
		query += " AND LOWER(e.from_address) = ?"
		args = append(args, address)
	case "to":
		query += " AND LOWER(e.to_address) = ?"
		args = append(args, address)
	case "both":
		query += " AND (LOWER(e.from_address) = ? OR LOWER(e.to_address) = ?)"
		args = append(args, address, address)
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

// handleContractAddressTransactions 查询某个合约中指定地址的相关交易
// GET /api/contract/{contractAddress}/address/{address}/transactions?page=1&size=20&role=from|to|both&func_name=transfer
func handleContractAddressTransactions(w http.ResponseWriter, r *http.Request, parts []string) {

	contractAddress := parts[0]
	address := parts[2]

	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}
	if !strings.HasPrefix(strings.ToLower(address), "0x") || len(address) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid address format")
		return
	}

	// 规范化地址为小写
	contractAddress = normalizeAddress(contractAddress)
	address = normalizeAddress(address)

	// 解析查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	role := strings.ToLower(r.URL.Query().Get("role"))
	if role != "from" && role != "to" && role != "" {
		writeError(w, http.StatusBadRequest, "Invalid role parameter, must be 'from', 'to', or empty (both)")
		return
	}
	if role == "" {
		role = "both" // 默认查询两种
	}
	funcName := r.URL.Query().Get("func_name")

	// 构建查询（使用LOWER()确保大小写不敏感的比较）
	offset := (page - 1) * size
	query := `SELECT t.tx_hash, t.block_number, t.block_time, t.from_address, t.to_address,
	          t.func_name, t.value, t.gas_used, t.status, c.contract_symbol, c.decimals
	          FROM transactions t
	          LEFT JOIN contracts c ON LOWER(t.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(t.contract_address) = ?`
	args := []interface{}{contractAddress}

	// 根据role参数添加地址筛选条件（使用LOWER()确保大小写不敏感）
	switch role {
	case "from":
		query += " AND LOWER(t.from_address) = ?"
		args = append(args, address)
	case "to":
		query += " AND LOWER(t.to_address) = ?"
		args = append(args, address)
	case "both":
		query += " AND (LOWER(t.from_address) = ? OR LOWER(t.to_address) = ?)"
		args = append(args, address, address)
	}

	// 函数名称筛选
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

// handleContractDetail 查询指定合约的详细信息
// GET /api/contract/{address}
func handleContractDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 从URL路径提取合约地址
	path := strings.TrimPrefix(r.URL.Path, "/api/contract/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "Contract address is required")
		return
	}

	// 验证地址格式（简单检查）
	if !strings.HasPrefix(strings.ToLower(path), "0x") || len(path) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

	// 规范化地址为小写
	path = normalizeAddress(path)

	contract, err := db.GetContractByAddress(path)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Contract not found: %v", err))
		return
	}

	// 格式化总供应量
	totalSupplyFormatted := formatTokenAmount(contract.TotalSupply, contract.Decimals)

	detail := ContractDetail{
		Address:              contract.ContractAddress,
		Name:                 contract.ContractName,
		Symbol:               contract.ContractSymbol,
		Type:                 contract.ContractType,
		Decimals:             contract.Decimals,
		TotalSupply:          contract.TotalSupply.String(),
		TotalSupplyFormatted: totalSupplyFormatted,
		DeployTxHash:         contract.DeployTxHash,
		DeployBlockNumber:    contract.DeployBlockNumber,
		DeployTime:           contract.DeployBlockTime,
		Deployer:             contract.DeployerAddress,
		VerificationStatus:   contract.VerificationStatus,
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data:    detail,
	})
}

// handleContractList 查询合约列表
// GET /api/contracts?page=1&size=20&symbol=USDT
func handleContractList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	symbol := r.URL.Query().Get("symbol")

	// 构建查询
	offset := (page - 1) * size
	query := `SELECT contract_address, contract_name, contract_symbol, contract_type, 
	          decimals, total_supply, deploy_block_time 
	          FROM contracts WHERE 1=1`
	args := []interface{}{}

	if symbol != "" {
		query += " AND contract_symbol LIKE ?"
		args = append(args, "%"+symbol+"%")
	}

	query += " ORDER BY deploy_block_time DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.GetConn().Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var contracts []ContractListItem
	for rows.Next() {
		var item ContractListItem
		var totalSupplyStr string
		err := rows.Scan(
			&item.Address,
			&item.Name,
			&item.Symbol,
			&item.Type,
			&item.Decimals,
			&totalSupplyStr,
			&item.DeployTime,
		)
		if err != nil {
			continue
		}

		// 解析总供应量
		totalSupply, _ := new(big.Int).SetString(totalSupplyStr, 10)
		item.TotalSupply = totalSupplyStr
		item.TotalSupplyFormatted = formatTokenAmount(totalSupply, item.Decimals)

		contracts = append(contracts, item)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"contracts": contracts,
			"page":      page,
			"size":      size,
			"total":     len(contracts),
		},
	})
}

