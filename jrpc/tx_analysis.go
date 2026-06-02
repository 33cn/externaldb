package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/proto"
	"google.golang.org/grpc"

	dbevm "github.com/33cn/externaldb/db/evm"
	"github.com/33cn/externaldb/db/transaction"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	logtype "github.com/33cn/plugin/plugin/dapp/evm/types"
	etypes "github.com/ethereum/go-ethereum/core/types"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Chain33.QueryTransaction TransactionDetail
func getTxDetailFromChain33(host string, txHash string) (txDetail *types.TransactionDetail, err error) {
	hash, err := common.FromHex(txHash)
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

const (
	EvmActionNameCreate = "createEvmContract"
	EvmActionNameCall   = "callEvmContract"
)

type EvmTxInfo struct {
	IsEvmTx      bool
	ParseSuccess bool
	Error        string

	EvmTxId     string
	Chain33TxId string

	ContractAddress string
	CallAddress     string
	Amount          uint64
	Asset           transaction.Asset
	GasLimit        uint64

	ExecSuccess      bool
	GasUsed          uint64
	IsCreateContract bool

	Func   EvmFunctionCall
	Events []EvmEvent
}

type EvmFunctionCall struct {
	FuncName string
	Args     string
}

type EvmEvent struct {
	Name string
	Args map[string]interface{}
}

// create contract logs
// 601, LogContractData log.addr =  contract.address
// 603, LogCallContract   caller = 0x8387505d1571ee2b2d7339addb3f5dcf9f32c389 deployer
//  contractAddr 0x639874a1978065ea394a444f032400655ed55e7b
//	usedGas 2791275
// 605,LogEVMEventData , log.topic
/*  [
        "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000008387505d1571ee2b2d7339addb3f5dcf9f32c389"
    ]
*/
// 		604, 		LogEVMStateChangeItem 		*N
func isEvmTx(execer string) bool {
	return (execer == "evm" || strings.HasSuffix(string(execer), ".evm"))
}

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
		log.Error("ParseTx", "Decode Tx Payload", err.Error())
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
		return parseEvmCallGoTx(&info, payload.Para)
	}

	// note 作为交易evm交易的内容
	ntx := new(etypes.Transaction)
	ntxRaw, err := common.FromHex(payload.Note)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "decode note to eth failed: " + err.Error()
		log.Error("ParseTx", "Decode Tx Note", err.Error())
		return &info
	}

	err = ntx.UnmarshalBinary(ntxRaw)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "parse eth-th failed: " + err.Error()
		log.Error("ParseTx", " Tx ", info.Error, "hash", common.ToHex(txDetail.Tx.Tx().Hash()))
		return &info
	}

	evmId := ntx.Hash().Bytes()
	evmIdStr := common.ToHex(evmId)
	info.EvmTxId = evmIdStr
	info.Chain33TxId = common.ToHex(txDetail.Tx.Tx().Hash())

	// coins 转账
	if len(ntx.Data()) == 0 {
		info.Func.FuncName = "transfer"
		info.Func.Args = fmt.Sprintf("{\"to\": \"%v\",\"amount\", \"%v\"}", common.ToHex(payload.Para), info.Amount)
		if txDetail.Receipt.Ty == 2 {
			info.ExecSuccess = true
		}
		return &info
	}
	// 合约操作 : 部署合约
	// if len(ntx.Data()) != 0 {
	if txDetail.ActionName == EvmActionNameCreate {
		info.IsCreateContract = true
		// parsy log 603 for detail
		//str := fmt.Sprintf("deploy contract: ")
		info.Func.FuncName = "deploy_contract"
		_ = parseLogs("", txDetail, &info)
		info.Func.Args = fmt.Sprintf("{\"contract\": \"%v\"}", info.ContractAddress)
		return &info
	}
	// 调用合约功能
	abi, err := getabi(info.ContractAddress)
	if err != nil {
		info.ParseSuccess = false
		log.Error("ParseTx", " get abi failed ", err.Error())
		return &info
	}
	// 1. 解析调用功能的参数
	fun, arg, err := parseParam(abi, payload.Para, nil)
	if err != nil {
		info.ParseSuccess = false
		log.Error("ParseTx", " abi parse parseParam ", err.Error())
		return &info
	}
	info.Func.FuncName = fun
	info.Func.Args = arg

	// 2. 解析调用功能产生的事件
	amount := parseLogs(abi, txDetail, &info)
	if info.Amount == 0 {
		info.Amount = amount
		info.Asset.Amount = int64(amount)
	}

	// 3. 通过evmHash 去数据库查询 transfer相关的log
	transferLogs, err := getTransferLogs(info.EvmTxId)
	if err != nil {
		log.Error("ParseTx", " get transfer logs failed ", err.Error())
		// 不返回错误，继续处理其他信息
	} else {
		info.Events = append(info.Events, transferLogs...)
	}

	return &info
}

func getEvent(data string) []string {
	return []string{""}
}

func parseParam(abiStr string, data []byte, m map[string]interface{}) (string, string, error) {
	log.Debug("parseParam  start")
	abi, err := pabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", "", fmt.Errorf("parseParam: map is nil")
	}

	pm, err := dbevm.UnpackParam(data, &abi)
	if err != nil {
		log.Error("parseParam: UnpackParam", "err", err)
		return "", "", err
	}
	buf, err := json.Marshal(pm)
	if err != nil {
		log.Error("parseParam: json.Marshal(pm)", "err", err)
		return "", "", err
	}
	log.Debug("parseParam  end")
	return pm["call_func_name"].(string), string(buf), nil
}

func parseLogs(abi string, tx *types.TransactionDetail, info *EvmTxInfo) uint64 {
	if tx.Receipt.Ty == 2 {
		info.ExecSuccess = true
	}

	amount := uint64(0)

	for i, log1 := range tx.Receipt.Logs {
		switch log1.Ty {
		case logtype.TyLogCallContract: // 603: // LogCallContract
			log.Debug("LogCallContract event start:")
			var l logtype.ReceiptEVMContract
			err := types.Decode(log1.Log, &l)
			if err != nil {
				info.Error = "decode log failed: " + err.Error()
				log.Error("decode log failed:", "err", err)
				return amount
			}
			info.GasUsed = l.UsedGas
			info.ContractAddress = l.ContractAddr
			log.Debug("LogCallContract event end:")
		case logtype.TyLogEVMEventData: //  605: // LogEVMEventData
			if "" == abi {
				continue
			}
			log.Debug("UnpackEvent event start:")
			var e types.EVMLog
			err := types.Decode(log1.Log, &e)
			if err != nil {
				info.Error = "decode event failed: " + err.Error()
				log.Error("decode event failed:", "idx", i, "err", err)
				return amount
			}
			log.Debug("UnpackEvent event start:", "log size", len(log1.Log), "topic size", len(e.Topic), "data size", len(e.Data))
			name, args, err := UnpackEvent(abi, e.Topic, e.Data)
			if err != nil {
				info.Error = "UnpackEvent event failed: " + err.Error()
				log.Error("UnpackEvent event failed:", "idx", i, "err", err, "name", name, "args", args)
				return amount
			}
			log.Debug("UnpackEvent event  :", "name", name)
			if name != "" {
				info.Events = append(info.Events, EvmEvent{Name: name, Args: args})
			}
			for k1, v1 := range args {
				if k1 == "amount" {
					b1 := v1.(*big.Int)
					divisor := big.NewInt(1e10)
					b1 = b1.Div(b1, divisor)
					amount = b1.Uint64()
					//t := reflect.TypeOf(v1)
					log.Debug("UnpackEvent event  :", "amount", amount, "value", v1)

				}
			}
		}
	}
	return amount
}

// 日志对象的数组，包含了由交易执行过程中触发的合约事件生成的日志条目
// event : Transfer(address indexed from , address indexed to, uint amount)
// topics: [event-by-keccak256, indexed-f1-from, indexed-f2-to]
// data: amount
// keccak256("Transfer(address,address,uint256)")
// 每个 topic 的大小为 32 字节

// return event-name event-args
func UnpackEvent(abiStr string, topics [][]byte, data []byte) (string, map[string]interface{}, error) {
	if len(topics) <= 0 {
		return "", nil, nil
	}
	eData := data //pcom.FromHex(data)

	var hashs []pcom.Hash
	for _, topic := range topics {
		hashs = append(hashs, pcom.BytesToHash(topic))
	}
	log.Debug("event topic", "event", common.ToHex(topics[0][:8]), "topic size", len(topics), "data-size", len(eData))

	contractABI, err := pabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", nil, err
	}

	//for i, x := range contractABI.Events {
	//	log.Debug("event info", "i", i, "e", x.String())
	//}

	name, args, err := dbevm.UnpackEvent(eData, hashs, &contractABI)
	return name, args, err
}

// getTransferLogs 根据evmHash（交易哈希）从MySQL数据库查询Transfer事件
// 参考 handleContractAddressTransfersImpl，但不需要合约地址参数，只列出对应交易的transfer event
func getTransferLogs(evmHash string) ([]EvmEvent, error) {
	if mysqlDB == nil {
		return nil, fmt.Errorf("MySQL database not initialized")
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
	// 不需要合约地址参数，只根据tx_hash查询
	query := `SELECT e.tx_hash, e.block_number, e.block_time, e.from_address, e.to_address, 
	          e.value, c.contract_symbol, c.decimals, e.contract_address
	          FROM events e
	          LEFT JOIN contracts c ON LOWER(e.contract_address) = LOWER(c.contract_address)
	          WHERE LOWER(e.tx_hash) = LOWER(?) AND e.event_name = 'Transfer'
	          ORDER BY e.log_index ASC`

	rows, err := mysqlDB.Query(query, evmHash)
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
		var tokenDecimals sql.NullInt16 // MySQL driver 不支持 NullUint8，使用 NullInt16
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
			log.Error("getTransferLogs", "scan error", err)
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

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}
