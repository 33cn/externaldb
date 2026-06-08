package main

import "time"

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

// AccountERC20TokenBalance 某地址持有的一笔 ERC20（按地址查询持仓列表项）
type AccountERC20TokenBalance struct {
	ContractAddress  string    `json:"contract_address"`
	Name               string    `json:"name"`
	Symbol             string    `json:"symbol"`
	Decimals           uint8     `json:"decimals"`
	Balance            string    `json:"balance"`
	BalanceFormatted   string    `json:"balance_formatted"`
	LastTxHash         string    `json:"last_tx_hash"`
	LastTxBlock        uint64    `json:"last_tx_block"`
	LastUpdated        time.Time `json:"last_updated"`
}

// ApprovalInfo 授权信息（按 owner 查询时返回的列表项）
type ApprovalInfo struct {
	ContractAddress  string    `json:"contract_address"`
	ContractName     string    `json:"contract_name"`
	ContractSymbol   string    `json:"contract_symbol"`
	Decimals         uint8     `json:"decimals"`
	Spender          string    `json:"spender"`
	Amount           string    `json:"amount"`
	AmountFormatted  string    `json:"amount_formatted"`
	LastTxHash       string    `json:"last_tx_hash"`
	LastBlockNumber  uint64    `json:"last_block_number"`
	LastUpdated      time.Time `json:"last_updated"`
}

