package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// handleTokensRouter 路由分发函数，处理 /evmapi/tokens 路径
func handleTokensRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 移除前缀 /evmapi/tokens
	path := strings.TrimPrefix(r.URL.Path, "/evmapi/tokens")

	// 如果路径为空或只有 /，则是列表接口
	if path == "" || path == "/" {
		handleTokenList(w, r)
		return
	}

	// 移除开头的 /
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// /evmapi/tokens/{address}
	if len(parts) == 1 {
		handleTokenDetail(w, r, parts[0])
		return
	}

	// /evmapi/tokens/{address}/transfers
	if len(parts) == 2 && parts[1] == "transfers" {
		handleTokenTransfers(w, r, parts[0])
		return
	}

	writeError(w, http.StatusNotFound, "Invalid path")
}

// handleTokenList 查询ERC20 token列表
// GET /evmapi/tokens?page=1&size=20&symbol=USDT&name=Token
func handleTokenList(w http.ResponseWriter, r *http.Request) {
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
	name := r.URL.Query().Get("name")

	// 构建查询，只查询ERC20类型的合约
	offset := (page - 1) * size
	query := `SELECT contract_address, contract_name, contract_symbol, contract_type, 
	          decimals, total_supply, deploy_block_time 
	          FROM contracts WHERE contract_type = 'ERC20'`
	args := []interface{}{}

	if symbol != "" {
		query += " AND contract_symbol LIKE ?"
		args = append(args, "%"+symbol+"%")
	}
	if name != "" {
		query += " AND contract_name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	query += " ORDER BY deploy_block_time DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.GetConn().Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer rows.Close()

	var tokens []ContractListItem
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

		tokens = append(tokens, item)
	}

	// 获取总数（用于分页）
	countQuery := `SELECT COUNT(*) FROM contracts WHERE contract_type = 'ERC20'`
	countArgs := []interface{}{}
	if symbol != "" {
		countQuery += " AND contract_symbol LIKE ?"
		countArgs = append(countArgs, "%"+symbol+"%")
	}
	if name != "" {
		countQuery += " AND contract_name LIKE ?"
		countArgs = append(countArgs, "%"+name+"%")
	}

	var total int
	err = db.GetConn().QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		total = len(tokens) // 如果查询总数失败，使用当前返回的数量
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"tokens": tokens,
			"page":   page,
			"size":   size,
			"total":  total,
		},
	})
}

// handleTokenDetail 查询指定ERC20代币的详细信息
// GET /evmapi/tokens/{address}
func handleTokenDetail(w http.ResponseWriter, r *http.Request, address string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(address), "0x") || len(address) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid token address format")
		return
	}

	// 规范化地址为小写
	address = normalizeAddress(address)

	contract, err := db.GetContractByAddress(address)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Token not found: %v", err))
		return
	}

	// 检查是否是ERC20代币
	if contract.ContractType != "ERC20" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Address is not an ERC20 token, contract type: %s", contract.ContractType))
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

// handleTokenTransfers 查询指定代币的转账列表
// GET /evmapi/tokens/{address}/transfers?page=1&size=20&from=0x...&to=0x...
func handleTokenTransfers(w http.ResponseWriter, r *http.Request, tokenAddress string) {
	// 验证地址格式
	if !strings.HasPrefix(strings.ToLower(tokenAddress), "0x") || len(tokenAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid token address format")
		return
	}

	// 规范化地址为小写
	tokenAddress = normalizeAddress(tokenAddress)

	// 先验证该地址是ERC20代币
	contract, err := db.GetContractByAddress(tokenAddress)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Token not found: %v", err))
		return
	}

	// 检查是否是ERC20代币
	if contract.ContractType != "ERC20" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Address is not an ERC20 token, contract type: %s", contract.ContractType))
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
	args := []interface{}{tokenAddress}

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

	// 获取总数（用于分页）
	countQuery := `SELECT COUNT(*) FROM events e WHERE LOWER(e.contract_address) = ? AND e.event_name = 'Transfer'`
	countArgs := []interface{}{tokenAddress}
	if fromAddr != "" {
		countQuery += " AND LOWER(e.from_address) = ?"
		countArgs = append(countArgs, fromAddr)
	}
	if toAddr != "" {
		countQuery += " AND LOWER(e.to_address) = ?"
		countArgs = append(countArgs, toAddr)
	}

	var total int
	err = db.GetConn().QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		total = len(transfers) // 如果查询总数失败，使用当前返回的数量
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"transfers": transfers,
			"page":      page,
			"size":      size,
			"total":     total,
		},
	})
}
