package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/proto"
	"google.golang.org/grpc"

	dbevm "github.com/33cn/externaldb/db/evm"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	logtype "github.com/33cn/plugin/plugin/dapp/evm/types"
	etypes "github.com/ethereum/go-ethereum/core/types"
)

func analysisTx(txHash string) string {

	tx, err := getTxDetailFromChain33("", "")
	if err != nil {
		return ""
	}

	abi := "" // todo get abi

	args := parseEvmTx(tx, abi)

	events := getEvent("")

	_, _ = args, events
	return ""
}

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

	ContractAddress string
	CallAddress     string
	Amount          uint64
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
func parseEvmTx(txDetail *types.TransactionDetail, abi string) *EvmTxInfo {
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
		return &info
	}
	info.ParseSuccess = true

	info.ContractAddress = payload.ContractAddr
	info.CallAddress = txDetail.Tx.From()
	info.Amount = payload.Amount
	info.GasLimit = payload.GasLimit

	// note 作为交易evm交易的内容
	ntx := new(etypes.Transaction)
	ntxRaw, err := common.FromHex(payload.Note)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "decode note to eth failed: " + err.Error()
	}
	err = ntx.UnmarshalBinary(ntxRaw)
	if err != nil {
		info.ParseSuccess = false
		info.Error = "parse eth-th failed: " + err.Error()
		return &info
	}

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
		parseLogs("", txDetail, &info)
		info.Func.Args = fmt.Sprintf("{\"contract\": \"%v\"}", info.ContractAddress)
		return &info
	}
	// 调用合约功能

	fun, arg, err := parseParam(abi, payload.Para, nil)
	if err != nil {
		info.ParseSuccess = false
		return &info
	}
	info.Func.FuncName = fun
	info.Func.Args = arg

	parseLogs(abi, txDetail, &info)

	return &info
}

func getEvent(data string) []string {
	return []string{""}
}

func parseParam(abiStr string, data []byte, m map[string]interface{}) (string, string, error) {
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
	return pm["call_func_name"].(string), string(buf), nil

}

func parseLogs(abi string, tx *types.TransactionDetail, info *EvmTxInfo) {
	if tx.Receipt.Ty == 2 {
		info.ExecSuccess = true
	}

	for _, log := range tx.Receipt.Logs {
		switch log.Ty {
		case logtype.TyLogCallContract: // 603: // LogCallContract
			var l logtype.ReceiptEVMContract
			err := types.Decode(log.Log, &l)
			if err != nil {
				info.Error = "decode log failed: " + err.Error()
				return
			}
			info.GasUsed = l.UsedGas
			info.ContractAddress = l.ContractAddr
		case logtype.TyLogEVMEventData: //  605: // LogEVMEventData
			if "" == abi {
				continue
			}
			var e types.EVMLog
			err := types.Decode(log.Log, &e)
			if err != nil {
				info.Error = "decode log failed: " + err.Error()
				return
			}
			name, args, err := UnpackEvent(abi, e.Topic, string(e.Data))
			if err != nil {
				info.Error = "decode event failed: " + err.Error()
				return
			}
			info.Events = append(info.Events, EvmEvent{Name: name, Args: args})
		}
	}
}

// 日志对象的数组，包含了由交易执行过程中触发的合约事件生成的日志条目
// event : Transfer(address indexed from , address indexed to, uint amount)
// topics: [event-by-keccak256, indexed-f1-from, indexed-f2-to]
// data: amount
// keccak256("Transfer(address,address,uint256)")
// 每个 topic 的大小为 32 字节

func UnpackEvent(abiStr string, topics [][]byte, data string) (string, map[string]interface{}, error) {
	eData := pcom.FromHex(data)
	var hashs []pcom.Hash
	for _, topic := range topics {
		hashs = append(hashs, pcom.BytesToHash(topic))
	}
	contractABI, err := pabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", nil, err
	}

	name, args, err := dbevm.UnpackEvent(eData, hashs, &contractABI)
	return name, args, err
}
