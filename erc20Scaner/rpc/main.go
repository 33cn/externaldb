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

	"github.com/33cn/externaldb/erc20Scaner/config"
	"github.com/33cn/externaldb/erc20Scaner/database"
)

var (
	port        = flag.String("port", "8080", "HTTP server port")
	dbDSN       = flag.String("dsn", "root:password@tcp(localhost:3306)/token_scanner?charset=utf8mb4&parseTime=True&loc=Local", "database DSN")
	configFile  = flag.String("c", "config.yaml", "config file path")
	chainGrpc   = flag.String("chain_grpc", "localhost:8802", "Chain33 gRPC host")
	chainSymbol = flag.String("chain_symbol", "bty", "Chain33 symbol")
	esHost      = flag.String("es_host", "http://localhost:9200/", "Elasticsearch host")
	esPrefix    = flag.String("es_prefix", "db01_", "Elasticsearch prefix")
	esVersion   = flag.Int("es_version", 7, "Elasticsearch version")
	esUser      = flag.String("es_user", "", "Elasticsearch username")
	esPassword  = flag.String("es_password", "", "Elasticsearch password")
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

	// 加载配置文件（如果存在）
	var cfg *config.Config
	if *configFile != "" {
		c, err := config.LoadConfig(*configFile)
		if err != nil {
			log.Printf("Warning: failed to load config file %s: %v, using defaults", *configFile, err)
			cfg = config.GetDefaultConfig()
		} else {
			log.Printf("Config loaded from %s", *configFile)
			cfg = c
		}
	} else {
		cfg = config.GetDefaultConfig()
	}

	// 记录哪些flag被显式设置，用于决定是否被配置文件覆盖
	var dsnFlag, chainGrpcFlag, esHostFlag, esPrefixFlag, esVersionFlag, esUserFlag, esPwdFlag bool
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "dsn":
			dsnFlag = true
		case "chain_grpc":
			chainGrpcFlag = true
		case "es_host":
			esHostFlag = true
		case "es_prefix":
			esPrefixFlag = true
		case "es_version":
			esVersionFlag = true
		case "es_user":
			esUserFlag = true
		case "es_password":
			esPwdFlag = true
		}
	})

	// 使用配置文件覆盖未显式设置的参数
	if cfg != nil {
		// 数据库 DSN
		if !dsnFlag && cfg.Database.DSN != "" {
			*dbDSN = cfg.Database.DSN
		}
		// Chain33 gRPC
		if !chainGrpcFlag && cfg.Node.GRPC != "" {
			*chainGrpc = cfg.Node.GRPC
		}
		// ES 配置（用于从 ES 查询 ABI）
		if cfg.ES.Host != "" && !esHostFlag {
			*esHost = cfg.ES.Host
		}
		if cfg.ES.Prefix != "" && !esPrefixFlag {
			*esPrefix = cfg.ES.Prefix
		}
		if cfg.ES.Version != 0 && !esVersionFlag {
			*esVersion = int(cfg.ES.Version)
		}
		if cfg.ES.User != "" && !esUserFlag {
			*esUser = cfg.ES.User
		}
		if cfg.ES.Password != "" && !esPwdFlag {
			*esPassword = cfg.ES.Password
		}
	}

	// 最终使用的 DSN
	dsn := *dbDSN

	// 连接数据库
	var err error
	db, err = database.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("Database connected successfully")
	log.Printf("Starting HTTP server on port %s", *port)

	// 注册路由（注意：handleContractAddressTransactions 和 handleContractAddressTransfers 会先检查路径，如果不是匹配的格式会调用 handleContractDetail）
	http.HandleFunc("/api/contract/", handleContractAddressTransfers)
	http.HandleFunc("/api/contracts", handleContractList)
	http.HandleFunc("/api/token/", handleTokenDetail)
	http.HandleFunc("/api/tokens", handleTokenList)
	http.HandleFunc("/api/transfers/", handleTransfers)
	http.HandleFunc("/api/transactions/", handleTransactions)
	http.HandleFunc("/api/holders/", handleHolders)
	http.HandleFunc("/api/parse_tx", handleParseTx)
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

// handleTokenDetail 查询指定ERC20代币的详细信息
// GET /api/token/{address}
func handleTokenDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 从URL路径提取代币地址
	path := strings.TrimPrefix(r.URL.Path, "/api/token/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "Token address is required")
		return
	}

	// 验证地址格式（简单检查）
	if !strings.HasPrefix(strings.ToLower(path), "0x") || len(path) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid token address format")
		return
	}

	// 规范化地址为小写
	path = normalizeAddress(path)

	contract, err := db.GetContractByAddress(path)
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

// handleTokenList 查询ERC20 token列表
// GET /api/tokens?page=1&size=20&symbol=USDT&name=Token
func handleTokenList(w http.ResponseWriter, r *http.Request) {
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
	detail, err := getTxDetailFromChain33(*chainGrpc, req.TxHash)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Transaction not found: %v", err))
		return
	}

	// 解析EVM交易
	parsed := parseEvmTx(detail, getAbiFromES, *chainSymbol)

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data:    parsed,
	})
}
