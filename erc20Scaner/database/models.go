package database

import (
	"database/sql"
	"math/big"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// FunctionSignature 函数签名模型
type FunctionSignature struct {
	ID                uint64    `db:"id"`
	Selector          string    `db:"selector"`
	FunctionName      string    `db:"function_name"`
	FunctionSignature string    `db:"function_signature"`
	StandardType      string    `db:"standard_type"`
	FunctionType      string    `db:"function_type"`
	Description       string    `db:"description"`
	IsStandard        bool      `db:"is_standard"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

// Contract 合约模型
type Contract struct {
	ID                 uint64    `db:"id"`
	ContractAddress    string    `db:"contract_address"`
	ContractName       string    `db:"contract_name"`
	ContractSymbol     string    `db:"contract_symbol"`
	ContractType       string    `db:"contract_type"` // ERC20/ERC721/ERC1155/UNKNOWN等
	Decimals           uint8     `db:"decimals"`
	TotalSupply        *big.Int  `db:"total_supply"`
	DeployTxHash       string    `db:"deploy_tx_hash"`
	DeployBlockNumber  uint64    `db:"deploy_block_number"`
	DeployBlockTime    time.Time `db:"deploy_block_time"`
	DeployerAddress    string    `db:"deployer_address"`
	VerificationStatus int8      `db:"verification_status"`
	VerifiedFunctions  string    `db:"verified_functions"` // JSON格式的函数选择器列表
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// Transaction 交易模型
type Transaction struct {
	ID              uint64    `db:"id"`
	TxHash          string    `db:"tx_hash"`
	BlockNumber     uint64    `db:"block_number"`
	BlockHash       string    `db:"block_hash"`
	BlockTime       time.Time `db:"block_time"`
	TxIndex         uint      `db:"tx_index"`
	FromAddress     string    `db:"from_address"`
	ToAddress       string    `db:"to_address"`
	ContractAddress string    `db:"contract_address"`
	FuncSelector    string    `db:"func_selector"`
	FuncName        string    `db:"func_name"`
	Value           *big.Int  `db:"value"`
	GasLimit        uint64    `db:"gas_limit"`
	GasUsed         uint64    `db:"gas_used"`
	GasPrice        *big.Int  `db:"gas_price"`
	TxFee           *big.Int  `db:"tx_fee"`
	Status          int8      `db:"status"`
	TxData          string    `db:"tx_data"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// TokenBalance 代币余额模型
type TokenBalance struct {
	ID                uint64     `db:"id"`
	Address           string     `db:"address"`
	ContractAddress   string     `db:"contract_address"`
	Balance           *big.Int   `db:"balance"`
	BalanceFormatted  *big.Float `db:"balance_formatted"`
	LastTxHash        string     `db:"last_tx_hash"`
	LastTxBlockNumber uint64     `db:"last_tx_block_number"`
	LastUpdatedAt     time.Time  `db:"last_updated_at"`
	CreatedAt         time.Time  `db:"created_at"`
}

// Event 事件模型
type Event struct {
	ID              uint64     `db:"id"`
	TxHash          string     `db:"tx_hash"`
	BlockNumber     uint64     `db:"block_number"`
	BlockTime       time.Time  `db:"block_time"`
	LogIndex        uint       `db:"log_index"`
	ContractAddress string     `db:"contract_address"`
	EventName       string     `db:"event_name"`
	EventSignature  string     `db:"event_signature"`
	FromAddress     string     `db:"from_address"`
	ToAddress       string     `db:"to_address"`
	Value           *big.Int   `db:"value"`
	ValueFormatted  *big.Float `db:"value_formatted"`
	OwnerAddress    string     `db:"owner_address"`
	SpenderAddress  string     `db:"spender_address"`
	Amount          *big.Int   `db:"amount"`
	Topic0          string     `db:"topic0"`
	Topic1          string     `db:"topic1"`
	Topic2          string     `db:"topic2"`
	Topic3          string     `db:"topic3"`
	Data            string     `db:"data"`
	CreatedAt       time.Time  `db:"created_at"`
}

// DB 数据库操作接口
type DB struct {
	conn *sql.DB
}

// GetConn 获取数据库连接（用于复杂查询）
func (db *DB) GetConn() *sql.DB {
	return db.conn
}

// NewDB 创建数据库连接
func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetFunctionSignature 根据选择器获取函数签名
func (db *DB) GetFunctionSignature(selector string) (*FunctionSignature, error) {
	query := `SELECT * FROM function_signatures WHERE selector = ?`

	sig := &FunctionSignature{}
	err := db.conn.QueryRow(query, selector).Scan(
		&sig.ID,
		&sig.Selector,
		&sig.FunctionName,
		&sig.FunctionSignature,
		&sig.StandardType,
		&sig.FunctionType,
		&sig.Description,
		&sig.IsStandard,
		&sig.CreatedAt,
		&sig.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// GetFunctionSignaturesByStandard 根据标准类型获取函数签名列表
func (db *DB) GetFunctionSignaturesByStandard(standardType string) ([]FunctionSignature, error) {
	query := `SELECT * FROM function_signatures WHERE standard_type = ? ORDER BY function_name`

	rows, err := db.conn.Query(query, standardType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signatures []FunctionSignature
	for rows.Next() {
		var sig FunctionSignature
		err := rows.Scan(
			&sig.ID,
			&sig.Selector,
			&sig.FunctionName,
			&sig.FunctionSignature,
			&sig.StandardType,
			&sig.FunctionType,
			&sig.Description,
			&sig.IsStandard,
			&sig.CreatedAt,
			&sig.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, sig)
	}

	return signatures, nil
}

// SaveContract 保存合约信息
func (db *DB) SaveContract(contract *Contract) error {
	query := `INSERT INTO contracts (
		contract_address, contract_name, contract_symbol, contract_type, decimals, 
		total_supply, deploy_tx_hash, deploy_block_number, deploy_block_time,
		deployer_address, verification_status, verified_functions
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		contract_name = VALUES(contract_name),
		contract_symbol = VALUES(contract_symbol),
		contract_type = VALUES(contract_type),
		decimals = VALUES(decimals),
		total_supply = VALUES(total_supply),
		verification_status = VALUES(verification_status),
		verified_functions = VALUES(verified_functions),
		updated_at = CURRENT_TIMESTAMP`

	// 处理 total_supply 字段：如果为 nil，传递 nil（SQL NULL），否则传递字符串
	var totalSupplyStr interface{}
	if contract.TotalSupply != nil {
		totalSupplyStr = contract.TotalSupply.String()
	} else {
		totalSupplyStr = nil
	}

	_, err := db.conn.Exec(query,
		contract.ContractAddress,
		contract.ContractName,
		contract.ContractSymbol,
		contract.ContractType,
		contract.Decimals,
		totalSupplyStr,
		contract.DeployTxHash,
		contract.DeployBlockNumber,
		contract.DeployBlockTime,
		contract.DeployerAddress,
		contract.VerificationStatus,
		contract.VerifiedFunctions,
	)
	return err
}

// SaveTransaction 保存交易信息
func (db *DB) SaveTransaction(tx *Transaction) error {
	query := `INSERT INTO transactions (
		tx_hash, block_number, block_hash, block_time, tx_index,
		from_address, to_address, contract_address, func_selector, func_name,
		value, gas_limit, gas_used, gas_price, tx_fee, status, tx_data
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		status = VALUES(status),
		gas_used = VALUES(gas_used),
		updated_at = CURRENT_TIMESTAMP`

	// 处理 DECIMAL 字段：如果为 nil，传递 nil（SQL NULL），否则传递字符串
	var valueStr interface{}
	if tx.Value != nil {
		valueStr = tx.Value.String()
	} else {
		valueStr = nil
	}

	var gasPriceStr interface{}
	if tx.GasPrice != nil {
		gasPriceStr = tx.GasPrice.String()
	} else {
		gasPriceStr = nil
	}

	var txFeeStr interface{}
	if tx.TxFee != nil {
		txFeeStr = tx.TxFee.String()
	} else {
		txFeeStr = nil
	}

	_, err := db.conn.Exec(query,
		tx.TxHash,
		tx.BlockNumber,
		tx.BlockHash,
		tx.BlockTime,
		tx.TxIndex,
		tx.FromAddress,
		tx.ToAddress,
		tx.ContractAddress,
		tx.FuncSelector,
		tx.FuncName,
		valueStr,
		tx.GasLimit,
		tx.GasUsed,
		gasPriceStr,
		txFeeStr,
		tx.Status,
		tx.TxData,
	)
	return err
}

// UpdateTokenBalance 更新代币余额
func (db *DB) UpdateTokenBalance(address, contractAddress string, balance *big.Int, txHash string, blockNumber uint64) error {
	query := `INSERT INTO token_balances (
		address, contract_address, balance, last_tx_hash, last_tx_block_number
	) VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		balance = VALUES(balance),
		last_tx_hash = VALUES(last_tx_hash),
		last_tx_block_number = VALUES(last_tx_block_number),
		last_updated_at = CURRENT_TIMESTAMP`

	_, err := db.conn.Exec(query,
		address,
		contractAddress,
		balance.String(),
		txHash,
		blockNumber,
	)
	return err
}

// SaveEvent 保存事件信息
func (db *DB) SaveEvent(event *Event) error {
	query := `INSERT INTO events (
		tx_hash, block_number, block_time, log_index, contract_address,
		event_name, event_signature, from_address, to_address, value,
		owner_address, spender_address, amount, topic0, topic1, topic2, topic3, data
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		value = VALUES(value),
		amount = VALUES(amount)`

	// 处理 value 字段：如果为 nil，传递 nil（SQL NULL），否则传递字符串
	var valueStr interface{}
	if event.Value != nil {
		valueStr = event.Value.String()
	} else {
		valueStr = nil
	}

	// 处理 amount 字段：如果为 nil，传递 nil（SQL NULL），否则传递字符串
	var amountStr interface{}
	if event.Amount != nil {
		amountStr = event.Amount.String()
	} else {
		amountStr = nil
	}

	_, err := db.conn.Exec(query,
		event.TxHash,
		event.BlockNumber,
		event.BlockTime,
		event.LogIndex,
		event.ContractAddress,
		event.EventName,
		event.EventSignature,
		event.FromAddress,
		event.ToAddress,
		valueStr,
		event.OwnerAddress,
		event.SpenderAddress,
		amountStr,
		event.Topic0,
		event.Topic1,
		event.Topic2,
		event.Topic3,
		event.Data,
	)
	return err
}

// GetContractByAddress 根据地址获取合约信息
func (db *DB) GetContractByAddress(address string) (*Contract, error) {
	query := `SELECT * FROM contracts WHERE contract_address = ?`

	contract := &Contract{}
	err := db.conn.QueryRow(query, address).Scan(
		&contract.ID,
		&contract.ContractAddress,
		&contract.ContractName,
		&contract.ContractSymbol,
		&contract.ContractType,
		&contract.Decimals,
		&contract.TotalSupply,
		&contract.DeployTxHash,
		&contract.DeployBlockNumber,
		&contract.DeployBlockTime,
		&contract.DeployerAddress,
		&contract.VerificationStatus,
		&contract.VerifiedFunctions,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

// GetTokenBalancesByAddress 获取地址的所有代币余额
func (db *DB) GetTokenBalancesByAddress(address string) ([]TokenBalance, error) {
	query := `SELECT * FROM token_balances WHERE address = ? ORDER BY balance DESC`

	rows, err := db.conn.Query(query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []TokenBalance
	for rows.Next() {
		var balance TokenBalance
		err := rows.Scan(
			&balance.ID,
			&balance.Address,
			&balance.ContractAddress,
			&balance.Balance,
			&balance.BalanceFormatted,
			&balance.LastTxHash,
			&balance.LastTxBlockNumber,
			&balance.LastUpdatedAt,
			&balance.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// GetEventsByContract 获取合约的所有事件
func (db *DB) GetEventsByContract(contractAddress string, limit int) ([]Event, error) {
	query := `SELECT * FROM events 
		WHERE contract_address = ? 
		ORDER BY block_number DESC, log_index DESC 
		LIMIT ?`

	rows, err := db.conn.Query(query, contractAddress, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.TxHash,
			&event.BlockNumber,
			&event.BlockTime,
			&event.LogIndex,
			&event.ContractAddress,
			&event.EventName,
			&event.EventSignature,
			&event.FromAddress,
			&event.ToAddress,
			&event.Value,
			&event.ValueFormatted,
			&event.OwnerAddress,
			&event.SpenderAddress,
			&event.Amount,
			&event.Topic0,
			&event.Topic1,
			&event.Topic2,
			&event.Topic3,
			&event.Data,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
