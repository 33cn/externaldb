package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleContractsRouter 路由分发函数，处理 /evmapi/contracts 路径
func handleContractsRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 移除前缀 /evmapi/contracts
	path := strings.TrimPrefix(r.URL.Path, "/evmapi/contracts")

	// 如果路径为空或只有 /，则是列表接口
	if path == "" || path == "/" {
		handleContractList(w, r)
		return
	}

	// 移除开头的 /
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// /evmapi/contracts/{address}
	if len(parts) == 1 {
		handleContractDetail(w, r, parts[0])
		return
	}

	// /evmapi/contracts/{address}/transfers
	if len(parts) == 2 && parts[1] == "transfers" {
		handleContractTransfers(w, r, parts[0])
		return
	}

	// /evmapi/contracts/{address}/transactions
	if len(parts) == 2 && parts[1] == "transactions" {
		handleContractTransactions(w, r, parts[0])
		return
	}

	// /evmapi/contracts/{address}/holders
	if len(parts) == 2 && parts[1] == "holders" {
		handleContractHolders(w, r, parts[0])
		return
	}

	// /evmapi/contracts/{contractAddress}/holders/{holderAddress}/transfers
	if len(parts) == 4 && parts[1] == "holders" && parts[3] == "transfers" {
		handleHolderTransfers(w, r, parts[0], parts[2])
		return
	}

	// /evmapi/contracts/{contractAddress}/holders/{holderAddress}/transactions
	if len(parts) == 4 && parts[1] == "holders" && parts[3] == "transactions" {
		handleHolderTransactions(w, r, parts[0], parts[2])
		return
	}

	writeError(w, http.StatusNotFound, "Invalid path")
}

// handleContractList 查询合约列表
// GET /evmapi/contracts?page=1&size=20&symbol=USDT
func handleContractList(w http.ResponseWriter, r *http.Request) {
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

// handleContractDetail 查询指定合约的详细信息
// GET /evmapi/contracts/{address}
func handleContractDetail(w http.ResponseWriter, r *http.Request, address string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(address), "0x") || len(address) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

	// 规范化地址为小写
	address = normalizeAddress(address)

	contract, err := db.GetContractByAddress(address)
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

// handleContractTransfers 查询指定合约的所有Transfer记录
// GET /evmapi/contracts/{address}/transfers?page=1&size=20&from=0x...&to=0x...
func handleContractTransfers(w http.ResponseWriter, r *http.Request, contractAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

	// 规范化合约地址为小写
	contractAddress = normalizeAddress(contractAddress)

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
	args := []interface{}{contractAddress}

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

// handleContractTransactions 查询指定合约的所有交易记录
// GET /evmapi/contracts/{address}/transactions?page=1&size=20&func_name=transfer
func handleContractTransactions(w http.ResponseWriter, r *http.Request, contractAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

	// 规范化合约地址为小写
	contractAddress = normalizeAddress(contractAddress)

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
	args := []interface{}{contractAddress}

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

// handleContractHolders 查询指定合约的Holder列表
// GET /evmapi/contracts/{address}/holders?page=1&size=20&min_balance=0
func handleContractHolders(w http.ResponseWriter, r *http.Request, contractAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

	// 规范化合约地址为小写
	contractAddress = normalizeAddress(contractAddress)

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
	args := []interface{}{contractAddress}

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

// handleHolderTransfers 查询特定合约中，特定持有者的转账记录
// GET /evmapi/contracts/{contractAddress}/holders/{holderAddress}/transfers?page=1&size=20&role=from|to|both
func handleHolderTransfers(w http.ResponseWriter, r *http.Request, contractAddress, holderAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}
	if !strings.HasPrefix(strings.ToLower(holderAddress), "0x") || len(holderAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid holder address format")
		return
	}

	// 规范化地址为小写
	contractAddress = normalizeAddress(contractAddress)
	holderAddress = normalizeAddress(holderAddress)

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
		args = append(args, holderAddress)
	case "to":
		query += " AND LOWER(e.to_address) = ?"
		args = append(args, holderAddress)
	case "both":
		query += " AND (LOWER(e.from_address) = ? OR LOWER(e.to_address) = ?)"
		args = append(args, holderAddress, holderAddress)
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

// handleHolderTransactions 查询特定合约中，特定持有者的相关交易
// GET /evmapi/contracts/{contractAddress}/holders/{holderAddress}/transactions?page=1&size=20&role=from|to|both&func_name=transfer
func handleHolderTransactions(w http.ResponseWriter, r *http.Request, contractAddress, holderAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(contractAddress), "0x") || len(contractAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}
	if !strings.HasPrefix(strings.ToLower(holderAddress), "0x") || len(holderAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid holder address format")
		return
	}

	// 规范化地址为小写
	contractAddress = normalizeAddress(contractAddress)
	holderAddress = normalizeAddress(holderAddress)

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
		args = append(args, holderAddress)
	case "to":
		query += " AND LOWER(t.to_address) = ?"
		args = append(args, holderAddress)
	case "both":
		query += " AND (LOWER(t.from_address) = ? OR LOWER(t.to_address) = ?)"
		args = append(args, holderAddress, holderAddress)
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
