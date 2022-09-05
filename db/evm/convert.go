package evm

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/contract"
	"github.com/33cn/externaldb/db/contractverify"
	ndb "github.com/33cn/externaldb/db/evm/nft/db"
	"github.com/33cn/go-kit/convert"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	dbcom "github.com/33cn/externaldb/db/common"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
)

var log = l.New("module", "db.evm")

// Convert tx convert
type Convert struct {
	symbol string
	title  string

	env     *db.TxEnv
	tx      *types.Transaction
	receipt *types.ReceiptData

	ConfigDB      proofconfig.ConfigDB
	ContractDB    contract.DB
	ContractVrfDB contractverify.DB
	NftTokenDb    ndb.TokenDB
	EvmTokenDb    TokenDB
}

func init() {
	converts.Register("evm", NewConvert)
}

// NewConvert new convert
func NewConvert(paraTitle, symbol string, _ []string) db.ExecConvert {
	e := &Convert{symbol: symbol, title: paraTitle}
	return e
}

// InitDB init db
func (c *Convert) InitDB(cli db.DBCreator) error {
	var err error
	err = contract.InitESDB(cli)
	if err != nil {
		return err
	}
	err = contractverify.InitESDB(cli)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, EVMX, EVMX, EVMMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, EVMTokenX, EVMTokenX, TokenMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, EVMTransferX, EVMTransferX, EVMTransferMapping)
	if err != nil {
		return err
	}
	for _, initDB := range initDBs {
		if err := initDB(cli); err != nil {
			return err
		}
	}
	return nil
}

func (c *Convert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "evm.token", c.env.Block.Block.Height, c.env.TxIndex)
}

// SetDB 访问权限控制index
func (c *Convert) SetDB(db db.WrapDB) error {
	c.ConfigDB = proofconfig.NewConfigDB(db)
	c.ContractDB = contract.NewEsDB(db)
	c.ContractVrfDB = contractverify.NewEsDB(db)
	c.NftTokenDb = ndb.NewTokenDB(db)
	c.EvmTokenDb = NewTokenDB(db)
	return nil
}

// ConvertTx impl
func (c *Convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	var err error
	c.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	c.tx = tx
	c.receipt = env.Block.Receipts[env.TxIndex]
	height := env.Block.Block.Height

	var records []db.Record
	var payload proto.EVMContractAction
	trans := transaction.ConvertTransaction(env)
	txOption := TxOption{}
	trans.Options = &txOption
	records = append(records, &transaction.TxRecord{
		IKey: transaction.NewTransactionKey(trans.Hash),
		Op:   db.NewOp(op),
		Tx:   trans,
	})
	log := log.New("tx_hash", trans.Hash)

	err = types.Decode(tx.Payload, &payload)
	if err != nil || c.receipt.Ty != types.ExecOk {
		//解析payload出错，打印错误日志。跳过解析交易详情并把交易的基本信息存到数据库
		if c.receipt.Ty == types.ExecOk {
			//交易执行成功但解析payload出错，说明代码存在bug
			log.Error(c.positionID(), "info", "decode payload error and skip convert tx", "error", err)
		}
		log.Error(c.positionID(), "info", "decode payload error convert tx", "error", err)
		return records, nil
	}
	txOption.ContractAddr = payload.ContractAddr

	// 签名地址是否有权限,没有直接返回
	fromAddr := tx.From()
	if !c.ConfigDB.IsHaveProofPermission(fromAddr) {
		log.Error("ConvertTx.evm:IsHaveProofPermission", "err", errors.New("ErrNoPermission"))
		return records, nil
	}

	_, err = c.ConfigDB.GetOrganizationName(fromAddr)
	if err != nil {
		log.Error("ConvertTx.evm:GetOrganizationName", "err", err, "addr", fromAddr)
		return records, nil
	}

	mapinfo := make(map[string]interface{})
	mapinfo, err = parseNote(mapinfo, payload.Note, payload.ContractAddr)
	if err != nil {
		log.Warn("parseNote", "err", err)
	}
	eventLogs := c.parseLog(mapinfo)
	mapinfo["to_addr"] = tx.To
	mapinfo["from_addr"] = fromAddr
	mapinfo["evm_tx_hash"] = common.ToHex(tx.Hash())
	mapinfo["evm_height_index"] = db.HeightIndex(height, env.TxIndex)
	mapinfo["evm_height"] = height
	mapinfo["evm_block_hash"] = env.BlockHash
	mapinfo["evm_block_time"] = env.Block.Block.BlockTime

	// 合约处理
	log = log.New("contract_addr", payload.ContractAddr)
	// 合约部署
	if db.ExecAddress(string(tx.Execer)) == payload.ContractAddr {
		mapinfo["contract_bin"] = hex.EncodeToString(payload.Code)
		ct := contract.NewContractByMap(mapinfo)

		vrf, err := c.ContractVrfDB.Get(ct.GetContractBinHash())
		switch err {
		case nil:
			ct.ContractAbi = vrf.ContractAbi
			ct.ContractType = vrf.ContractType
		case db.ErrDBNotFound:
			if ct.ContractVerify.ContractAbi != "" {
				records = append(records, contractverify.NewRecord(op, ct.ContractVerify))
			}
		default:
			log.Error("ContractVrfDB.Get", "err", err, "bin-hash", ct.GetContractBinHash())
		}
		if ct.ContractType == "" {
			contract.Detector.Detect(ct)
		}
		txOption.ContractAddr = ct.Address
		txOption.ContractType = ct.ContractType
		txOption.CallFunName = "DeployContract"
		ct.TxCount++
		c.ContractDB.UpdateCache(ct)
		records = append(records, contract.NewRecord(op, ct))
		return records, nil
	}
	//records = append(records, &RecordEVM{
	//	IKey:  NewEVMKey(EVM(mapinfo).Key()),
	//	Op:    db.NewOp(op),
	//	value: mapinfo,
	//})

	// 合约执行
	ct, err := c.ContractDB.Get(payload.ContractAddr)
	if err != nil {
		log.Error("ConvertTx.evm:ContractDB.Get", "err", err)
		return records, nil
	}
	mapinfo["contract_type"] = ct.ContractType
	if ct.ContractAbi == "" {
		if ct.ContractType != "" {
			ct.ParsedAbi = contract.DefaultABIs[ct.ContractType]
		} else {
			return records, nil
		}
	}

	params, err := c.parseParam(ct.ParsedAbi, payload.Para, mapinfo)
	if err != nil {
		log.Warn("ConvertTx.evm:parseParam", "err", err)
	}
	txOption.Params = params

	events := c.parseEvent(ct.ParsedAbi, mapinfo, eventLogs)
	txOption.Events = events

	// 指定合约调用处理
	callFuncName := convert.ToString(mapinfo["call_func_name"])
	if callFuncName == "" {
		callFuncName = hex.EncodeToString(payload.Para[:4])
	}
	if fcv, ok := GetFuncConvert(callFuncName); ok {
		rd, err := fcv(c, op, mapinfo)
		if err == nil {
			records = append(records, rd...)
		} else {
			log.Error("evm func convert", "err", err, "func_name", callFuncName)
		}
	} else {
		log.Warn("no evm func convert", "func_name", callFuncName)
	}
	// 指定event处理
	for key := range events {
		if handle, ok := GetEventHandle(ct.ContractType + key); ok {
			rd, err := handle(c, op, mapinfo)
			if err == nil {
				records = append(records, rd...)
			} else {
				log.Error("evm event handle", "err", err, "event_name", key)
			}
		}
	}
	// 数据收集
	transfers := make([]*Transfer, 0)
	tokenID := make([]string, 0)
	contractType := ct.ContractType
	mintCount := int64(0)
	for _, record := range records {
		if record.Index() == EVMTransferX {
			if r, ok := record.(db.SourceAbleRecord); ok {
				if t, ok := r.Source().(*Transfer); ok {
					transfers = append(transfers, t)
					tokenID = append(tokenID, t.TokenID)
					contractType = t.TokenType
					if t.IsMint() {
						mintCount += t.Amount
					}
				}
			}
		}
	}

	txOption.ContractType = contractType
	txOption.TokenID = tokenID
	txOption.Transfers = transfers
	txOption.CallFunName = callFuncName

	// 合约数量统计
	if op == db.OpAdd {
		ct.PublishCount += mintCount
		ct.TxCount++
	} else {
		ct.PublishCount -= mintCount
		ct.TxCount--
	}
	records = append(records, contract.NewRecord(db.OpUpdate, ct))
	return records, nil
}

func (c *Convert) parseLog(mapinfo map[string]interface{}) []*types.EVMLog {
	eLogs := make([]*types.EVMLog, 0)
	for _, lg := range c.receipt.Logs {
		switch lg.Ty {
		case evmtypes.TyLogContractData:
			var cd evmtypes.EVMContractData
			if err := types.Decode(lg.Log, &cd); err != nil {
				log.Error("parseLog: Decode EVMContractData", "err", err)
				continue
			}
			mapinfo["contract_addr"] = cd.Addr
			mapinfo["contract_name"] = cd.Name
			mapinfo["contract_creator"] = cd.Creator
		case evmtypes.TyLogCallContract:
			var ec evmtypes.ReceiptEVMContract
			if err := types.Decode(lg.Log, &ec); err != nil {
				log.Error("parseLog: Decode ReceiptEVMContract", "err", err)
				continue
			}
			mapinfo["contract_addr"] = ec.ContractAddr
			mapinfo["contract_used_gas"] = ec.UsedGas
		case evmtypes.TyLogEVMEventData:
			var el types.EVMLog
			if err := types.Decode(lg.Log, &el); err != nil {
				log.Error("parseLog: Decode EVMLog", "err", err)
				continue
			}
			if len(el.Topic) <= 0 {
				log.Error("parseLog: no topic", "len(topic)", len(el.Topic))
				continue
			}
			eLogs = append(eLogs, &el)
		}
	}
	return eLogs
}

func (c *Convert) parseEvent(abi *pabi.ABI, mapinfo map[string]interface{}, eLogs []*types.EVMLog) (events map[string]interface{}) {
	// parse event
	if len(eLogs) <= 0 {
		return
	}
	events = make(map[string]interface{})
	errEvents := make([]interface{}, 0)
	for _, el := range eLogs {
		var topics []pcom.Hash
		for _, topic := range el.Topic {
			topics = append(topics, pcom.BytesToHash(topic))
		}
		name, m, err := UnpackEvent(el.Data, topics, abi)
		if err != nil {
			log.Error("parseLog: UnpackEventLog", "err", err)
			m = make(map[string]interface{})
			m["data"] = el.Data
			m["topic"] = el.Topic
			m["err"] = err.Error()
			errEvents = append(errEvents, m)
			continue
		}
		events[name] = m
	}
	if len(events) > 0 {
		buf, err := json.Marshal(events)
		if err != nil {
			log.Error("parseLog json.Marshal(events)", "err", err)
			return
		}
		mapinfo["evm_events"] = string(buf)
	}
	if len(errEvents) > 0 {
		buf, err := json.Marshal(errEvents)
		if err != nil {
			log.Error("parseLog json.Marshal(errEvents)", "err", err)
			return
		}
		mapinfo["evm_err_parse_events"] = string(buf)
	}
	return
}

func (c *Convert) parseParam(abi *pabi.ABI, data []byte, m map[string]interface{}) (map[string]interface{}, error) {
	if m == nil {
		return nil, fmt.Errorf("parseParam: map is nil")
	}
	pm, err := UnpackParam(data, abi)
	if err != nil {
		log.Error("parseParam: UnpackParam", "err", err)
		return pm, err
	}
	buf, err := json.Marshal(pm)
	if err != nil {
		log.Error("parseParam: json.Marshal(pm)", "err", err)
		return pm, err
	}
	if m["call_func_name"] == nil || m["call_func_name"] == "" {
		m["call_func_name"] = pm["call_func_name"]
	}
	m["evm_param"] = string(buf)
	return pm, nil
}

// parseNote 若是部署合约，获取abi；若是调用合约，获取调用方法
func parseNote(mapinfo map[string]interface{}, payload string, addr string) (map[string]interface{}, error) {
	mapNote := make(map[string]interface{})
	note, err := dbcom.DecodeString(payload)
	if err != nil {
		log.Warn("parseNote:DecodeString payload.Note", "err", err)
		return mapinfo, err
	}

	err = json.Unmarshal(note, &mapNote)
	if err != nil {
		log.Warn("Unmarshal note failed", "err", err)
	}
	mapinfo["evm_note"] = note

	if abi, ok := mapNote["abi"]; ok {
		mapinfo["contract_abi"] = abi
	} else {
		mapinfo["call_func_name"] = mapNote["call_func_name"]
		mapinfo["contract_addr"] = addr
	}

	return mapinfo, nil
}
