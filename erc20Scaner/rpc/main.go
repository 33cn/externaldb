package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/externaldb/erc20Scaner/database"
)

var (
	port       = flag.String("port", "8080", "HTTP server port")
	dbDSN      = flag.String("dsn", "root:password@tcp(localhost:3306)/token_scanner?charset=utf8mb4&parseTime=True&loc=Local", "database DSN")
	configFile = flag.String("c", "config.yaml", "config file path")
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ContractDetail 合约详情响应
type ContractDetail struct {
	Address              string    `json:"address"`
	Name                 string    `json:"name"`
	Symbol               string    `json:"symbol"`
	Type                 string    `json:"type"`
	Decimals             uint8     `json:"decimals"`
	TotalSupply          string    `json:"total_supply"`
	TotalSupplyFormatted string    `json:"total_supply_formatted"`
	DeployTxHash         string    `json:"deploy_tx_hash"`
	DeployBlockNumber    uint64    `json:"deploy_block_number"`
	DeployTime           time.Time `json:"deploy_time"`
	Deployer             string    `json:"deployer"`
	VerificationStatus   int8      `json:"verification_status"`
}

// ContractListItem 合约列表项
type ContractListItem struct {
	Address              string    `json:"address"`
	Name                 string    `json:"name"`
	Symbol               string    `json:"symbol"`
	Type                 string    `json:"type"`
	Decimals             uint8     `json:"decimals"`
	TotalSupply          string    `json:"total_supply"`
	TotalSupplyFormatted string    `json:"total_supply_formatted"`
	DeployTime           time.Time `json:"deploy_time"`
}

// TransferRecord Transfer记录
type TransferRecord struct {
	TxHash         string    `json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number"`
	BlockTime      time.Time `json:"block_time"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	Value          string    `json:"value"`
	ValueFormatted string    `json:"value_formatted"`
	TokenSymbol    string    `json:"token_symbol"`
	TokenDecimals  uint8     `json:"token_decimals"`
}

// TransactionRecord 交易记录
type TransactionRecord struct {
	TxHash         string    `json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number"`
	BlockTime      time.Time `json:"block_time"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	FuncName       string    `json:"func_name"`
	Value          string    `json:"value"`
	ValueFormatted string    `json:"value_formatted"`
	GasUsed        uint64    `json:"gas_used"`
	Status         int8      `json:"status"`
	TokenSymbol    string    `json:"token_symbol"`
	TokenDecimals  uint8     `json:"token_decimals"`
}

// HolderInfo Holder信息
type HolderInfo struct {
	Address          string    `json:"address"`
	Balance          string    `json:"balance"`
	BalanceFormatted string    `json:"balance_formatted"`
	LastTxHash       string    `json:"last_tx_hash"`
	LastTxBlock      uint64    `json:"last_tx_block"`
	LastUpdated      time.Time `json:"last_updated"`
}

var db *database.DB

func main() {
	flag.Parse()

	// 加载配置（如果使用配置文件）
	var dsn string = *dbDSN
	if *configFile != "" {
		// 这里可以添加从配置文件读取DSN的逻辑
		// 为了简化，直接使用命令行参数
	}

	// 连接数据库
	var err error
	db, err = database.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("Database connected successfully")
	log.Printf("Starting HTTP server on port %s", *port)

	// 注册路由
	http.HandleFunc("/api/contract/", handleContractDetail)
	http.HandleFunc("/api/contracts", handleContractList)
	http.HandleFunc("/api/transfers/", handleTransfers)
	http.HandleFunc("/api/transactions/", handleTransactions)
	http.HandleFunc("/api/holders/", handleHolders)
	http.HandleFunc("/health", handleHealth)

	// 启动服务器
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleHealth 健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "OK",
		Data:    map[string]string{"status": "healthy"},
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
	if !strings.HasPrefix(path, "0x") || len(path) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid contract address format")
		return
	}

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

	// 构建查询
	offset := (page - 1) * size
	query := `SELECT e.tx_hash, e.block_number, e.block_time, e.from_address, e.to_address, 
	          e.value, c.contract_symbol, c.decimals
	          FROM events e
	          LEFT JOIN contracts c ON e.contract_address = c.contract_address
	          WHERE e.contract_address = ? AND e.event_name = 'Transfer'`
	args := []interface{}{path}

	if fromAddr != "" {
		query += " AND e.from_address = ?"
		args = append(args, fromAddr)
	}
	if toAddr != "" {
		query += " AND e.to_address = ?"
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

	// 构建查询
	offset := (page - 1) * size
	query := `SELECT t.tx_hash, t.block_number, t.block_time, t.from_address, t.to_address,
	          t.func_name, t.value, t.gas_used, t.status, c.contract_symbol, c.decimals
	          FROM transactions t
	          LEFT JOIN contracts c ON t.contract_address = c.contract_address
	          WHERE t.contract_address = ?`
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
