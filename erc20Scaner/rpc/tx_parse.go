package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	etypes "github.com/ethereum/go-ethereum/core/types"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"google.golang.org/grpc"
)

const (
	EvmActionNameCreate    = "createEvmContract"
	EvmActionNameCall      = "callEvmContract"
	EvmCallGoAddr          = "0x0000000000000000000000000000000000200005"
	ContractAbiDB          = "contractabi"
	ContractAbiDefaultType = "_doc"
)

type Asset struct {
	Exec   string `json:"exec"`
	Symbol string `json:"symbol"`
	Amount int64  `json:"amount"`
}

// EvmTxInfo EVM交易信息
type EvmTxInfo struct {
	IsEvmTx      bool
	ParseSuccess bool
	Error        string

	EvmTxId     string
	Chain33TxId string

	ContractAddress string
	CallAddress     string
	Amount          uint64
	Asset           Asset
	GasLimit        uint64

	ExecSuccess      bool
	GasUsed          uint64
	IsCreateContract bool

	Func   EvmFunctionCall
	Events []EvmEvent
}

// EvmFunctionCall 函数调用信息
type EvmFunctionCall struct {
	FuncName string
	Args     string
}

// EvmEvent 事件信息
type EvmEvent struct {
	Name string
	Args map[string]interface{}
}

// Abi ABI结构
type Abi struct {
	Address string `json:"address"`
	Abi     string `json:"abi"` // hex 格式
}

// ContractAbiID 生成合约ABI的ID
func ContractAbiID(contractAddress string) string {
	return fmt.Sprintf("contract-abi-%s", contractAddress)
}

// getTxDetailFromChain33 从Chain33节点获取交易详情
func getTxDetailFromChain33(host string, txHash string) (txDetail *types.TransactionDetail, err error) {
	hash, err := FromHex(txHash)
	if err != nil {
		return
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*100)))
	if err != nil {
		return
	}
	defer conn.Close()

	client := types.NewChain33Client(conn)
	txDetail, err = client.QueryTransaction(context.TODO(), &types.ReqHash{Hash: hash})
	if err != nil {
		return
	}
	return
}

// isEvmTx 判断是否是EVM交易
func isEvmTx(execer string) bool {
	return (execer == "evm" || strings.HasSuffix(string(execer), ".evm"))
}

// getAbiFromES 从ES获取ABI
func getAbiFromES(address string) (string, error) {
	cli, err := escli.NewESShortConnect(globalESHost, globalESPrefix, int32(globalESVersion), globalESUser, globalESPassword)
	if err != nil {
		log.Printf("ParseTx: NewESShortConnect failed: %v", err)
		return "", err
	}

	raw, err := cli.Get(ContractAbiDB, ContractAbiDefaultType, ContractAbiID(address))
	if err != nil {
		log.Printf("ParseTx: Get ABI failed: %v", err)
		return "", err
	}

	abi := Abi{}
	err = json.Unmarshal([]byte(*raw), &abi)
	if err != nil {
		log.Printf("ParseTx: decodeAbi failed: %v", err)
		return "", err
	}

	abiByte, err := FromHex(abi.Abi)
	if err != nil {
		log.Printf("ParseTx: FromHex failed: %v", err)
		return "", err
	}
	return string(abiByte), err
}

// parseEvmTx 解析EVM交易
// evm 合约有3中情况
// 1. 部署合约, 执行器为 evm, 地址为 evm地址, 类型为部署
// 2. 合约功能, 执行器为 evm, 地址为 合约地址,  类型为合约功能
// 3. 转账功能, 执行器为 evm, 地址为 evm地址, 类型为合约功能
func parseEvmTx(txDetail *types.TransactionDetail, getabi func(string) (string, error), symbol string) *EvmTxInfo {
	var info EvmTxInfo
	isEvm := isEvmTx(string(txDetail.Tx.Execer))
	if !isEvm {
		info.IsEvmTx = false
		return &info
	}
	info.IsEvmTx = true

	var payload proto.EVMContractAction
	err := types.Decode(txDetail.Tx.Payload, &payload)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "parse payload failed: " + err.Error()
		log.Printf("ParseTx: Decode Tx Payload failed: %v", err)
		return &info
	}
	info.ParseSuccess = true

	info.ContractAddress = payload.ContractAddr
	info.CallAddress = txDetail.Tx.From()
	info.GasLimit = payload.GasLimit
	info.Amount = payload.Amount
	info.Asset.Amount = int64(payload.Amount)
	info.Asset.Exec = string(txDetail.Tx.Execer)
	if strings.HasSuffix(string(info.Asset.Exec), ".evm") {
		info.Asset.Symbol = "Para"
	} else {
		info.Asset.Symbol = symbol
	}

	// 处理特殊的合约: 使用evm 合约调用go合约
	// 由于 没有solidity编译产生的abi, 所以这里插入处理
	if info.ContractAddress == EvmCallGoAddr {
		// 简化处理，直接返回
		info.Error = "EvmCallGo contract not fully supported"
		return &info
	}

	// note 作为交易evm交易的内容
	ntx := new(etypes.Transaction)
	ntxRaw, err := FromHex(payload.Note)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "decode note to eth failed: " + err.Error()
		log.Printf("ParseTx: Decode Tx Note failed: %v", err)
		return &info
	}

	err = ntx.UnmarshalBinary(ntxRaw)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "parse eth-th failed: " + err.Error()
		log.Printf("ParseTx: Tx parse failed: %v, hash: %s", err, ToHex(txDetail.Tx.Tx().Hash()))
		return &info
	}

	evmId := ntx.Hash().Bytes()
	evmIdStr := ToHex(evmId)
	info.EvmTxId = evmIdStr
	info.Chain33TxId = ToHex(txDetail.Tx.Tx().Hash())

	// coins 转账
	if len(ntx.Data()) == 0 {
		info.Func.FuncName = "transfer"
		info.Func.Args = fmt.Sprintf("{\"to\": \"%v\",\"amount\": \"%v\"}", ToHex(payload.Para), info.Amount)
		if txDetail.Receipt.Ty == 2 {
			info.ExecSuccess = true
		}
		return &info
	}
	// 合约操作 : 部署合约
	if txDetail.ActionName == EvmActionNameCreate {
		info.IsCreateContract = true
		info.Func.FuncName = "deploy_contract"
		_ = parseLogs("", txDetail, &info)
		info.Func.Args = fmt.Sprintf("{\"contract\": \"%v\"}", info.ContractAddress)
		return &info
	}
	// 调用合约功能
	// 1. 解析调用功能的参数
	signedFunc := ToHex(payload.Para[:4])
	info.Func.FuncName = signedFunc
	// 需要abi才能解析参数
	info.Func.Args = "args" // ToHex(payload.Para[4:])

	// 2. 解析调用功能产生的事件
	amount := parseLogs("", txDetail, &info)
	if info.Amount == 0 {
		info.Amount = amount
		info.Asset.Amount = int64(amount)
	}

	// 3. 通过evmHash 去数据库查询 transfer相关的log
	transferLogs, err := getTransferLogsFromDB(info.EvmTxId)
	if err != nil {
		log.Printf("ParseTx: get transfer logs failed: %v", err)
		// 不返回错误，继续处理其他信息
	} else {
		info.Events = append(info.Events, transferLogs...)
	}

	return &info
}

// parseLogs 解析日志
func parseLogs(abi string, tx *types.TransactionDetail, info *EvmTxInfo) uint64 {
	if tx.Receipt.Ty == 2 {
		info.ExecSuccess = true
	}

	amount := uint64(0)
	/*
		for _, log1 := range tx.Receipt.Logs {
			switch log1.Ty {
			case 603: // logtype.TyLogCallContract: // 603: // LogCallContract
				var l logtype.ReceiptEVMContract
				err := types.Decode(log1.Log, &l)
				if err != nil {
					info.Error = "decode log failed: " + err.Error()
					log.Printf("decode log failed: %v", err)
					return amount
				}
				info.GasUsed = l.UsedGas
				info.ContractAddress = l.ContractAddr
				//case logtype.TyLogEVMEventData: //  605: // LogEVMEventData
				//	continue
			}
		}
	*/
	return amount
}

// getTransferLogsFromDB 从MySQL数据库查询Transfer事件
func getTransferLogsFromDB(evmHash string) ([]EvmEvent, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// 规范化evmHash（统一转换为小写，保留0x前缀）
	evmHash = strings.ToLower(evmHash)
	if !strings.HasPrefix(evmHash, "0x") {
		evmHash = "0x" + evmHash
	}
	if len(evmHash) != 66 { // 0x + 64 hex chars
		return nil, fmt.Errorf("invalid evmHash format: %s", evmHash)
	}

	// 构建查询，查询指定交易的所有Transfer事件
	query := `SELECT e.tx_hash, e.block_number, e.block_time, e.from_address, e.to_address, 
	          e.value, c.contract_symbol, c.decimals, e.contract_address
	          FROM events e
	          LEFT JOIN contracts c ON LOWER(e.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(e.tx_hash) = LOWER(?) AND e.event_name = 'Transfer'
	          ORDER BY e.log_index ASC`

	rows, err := db.GetConn().Query(query, evmHash)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var events []EvmEvent
	for rows.Next() {
		var txHash string
		var blockNumber uint64
		var blockTime string
		var fromAddr string
		var toAddr string
		var valueStr sql.NullString
		var tokenSymbol sql.NullString
		var tokenDecimals sql.NullInt16
		var contractAddr string

		err := rows.Scan(
			&txHash,
			&blockNumber,
			&blockTime,
			&fromAddr,
			&toAddr,
			&valueStr,
			&tokenSymbol,
			&tokenDecimals,
			&contractAddr,
		)
		if err != nil {
			log.Printf("getTransferLogsFromDB: scan error: %v", err)
			continue
		}

		// 解析value
		var value *big.Int
		if valueStr.Valid && valueStr.String != "" {
			value, _ = new(big.Int).SetString(valueStr.String, 10)
		}
		if value == nil {
			value = big.NewInt(0)
		}

		// 构建事件参数
		args := make(map[string]interface{})
		args["from"] = fromAddr
		args["to"] = toAddr
		args["value"] = value
		args["amount"] = value // 兼容性，同时提供value和amount

		// 如果有代币信息，添加到参数中
		if tokenSymbol.Valid {
			args["token_symbol"] = tokenSymbol.String
		}
		if tokenDecimals.Valid {
			args["token_decimals"] = uint8(tokenDecimals.Int16)
		}
		args["contract_address"] = contractAddr

		// 创建EvmEvent
		event := EvmEvent{
			Name: "Transfer",
			Args: args,
		}
		log.Printf("getTransferLogsFromDB: event: %v", event)

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}

// FromHex hex -> []byte
func FromHex(s string) ([]byte, error) {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
		return hex.DecodeString(s)
	}
	return []byte{}, nil
}

// ToHex []byte -> hex
func ToHex(b []byte) string {
	hex := hex.EncodeToString(b)
	// Prefer output of "0x0" instead of "0x"
	if len(hex) == 0 {
		return ""
	}
	return "0x" + hex
}
