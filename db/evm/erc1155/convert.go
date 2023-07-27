package erc1155

import (
	"fmt"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/evm"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/go-kit/convert"
)

const ERC1155 = "ERC1155"

var (
	log = l.New("module", "db.evm.erc1155")
)

func init() {
	// TransferSingle 0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62
	// event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value)
	evm.RegisterEventHandle(ERC1155+"TransferSingle", TransferSingle)

	// TransferBatch 0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb
	// event TransferBatch(address indexed operator, address indexed from, address indexed to, uint256[] ids, uint256[] values)
	evm.RegisterEventHandle(ERC1155+"TransferBatch", TransferBatch)
}

func TransferSingle(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	event, err := data.GetEvent("TransferSingle")
	if err != nil {
		log.Error("TransferSingle data.GetEvent(\"TransferBatch\")", "err", err)
		return nil, err
	}
	trans := &evm.Transfer{}
	trans.TokenID = convert.ToString(event["id"])
	trans.Amount = convert.ToInt64(event["value"])
	trans.To = util.AddressConvert(convert.ToString(event["to"]))
	trans.From = util.AddressConvert(convert.ToString(event["from"]))
	trans.Operator = convert.ToString(event["operator"])
	trans.TokenType = convert.ToString(ERC1155)
	trans.LoadBlockData(data)

	return transferToken(c, op, trans), nil
}

func transferToken(c *evm.Convert, op int, trans *evm.Transfer, dbTokens ...*evm.Token) []db.Record {

	records := make([]db.Record, 0)
	// 转账记录
	if op == db.SeqTypeAdd {
		records = append(records, evm.NewRecordTransfer(trans, db.OpAdd))
	} else { // 回滚
		records = append(records, evm.NewRecordTransfer(trans, db.OpDel))
	}

	// 移除原有token，或者减对应数量
	if trans.From != evm.EmptyAddr && trans.From != evm.EmptyAddrEth { // 等于则代表发币，不做处理
		oldT, err := getFromToken(c, trans, dbTokens...)
		switch err {
		case nil:
			if op == db.SeqTypeAdd {
				oldT.Amount -= trans.Amount
			} else {
				oldT.Amount += trans.Amount
			}
			records = append(records, evm.NewRecordToken(oldT, db.OpAdd))
		case db.ErrDBNotFound:
			log.Error("transferToken old token not found", "token", fmt.Sprintf("%s-%s-%v", trans.ContractAddr, trans.From, trans.TokenID))
		default:
			log.Error("transferToken get old token", "", err)
		}
	}

	// 新增token，或者增加对应数量
	newT, err := getToToken(c, trans, dbTokens...)
	switch err {
	case nil:
		if op == db.SeqTypeAdd {
			newT.Amount += trans.Amount
		} else {
			newT.Amount -= trans.Amount
		}
		records = append(records, evm.NewRecordToken(newT, db.OpAdd))
	case db.ErrDBNotFound:
		token := trans.GetNewToken()
		c.EvmTokenDb.UpdateCache(token)
		records = append(records, evm.NewRecordToken(token, db.OpAdd))
	default:
		log.Error("transferToken get newT token", "", err)
	}
	return records
}

func getFromToken(c *evm.Convert, trans *evm.Transfer, dbTokens ...*evm.Token) (*evm.Token, error) {
	var oldT *evm.Token
	var err error
	if len(dbTokens) >= 1 {
		oldT = dbTokens[0]
		if oldT == nil {
			err = db.ErrDBNotFound
		}
	} else {
		oldT, err = c.EvmTokenDb.GetToken(trans.ContractAddr, util.AddressConvert(trans.From), trans.TokenID)
	}
	return oldT, err
}

func getToToken(c *evm.Convert, trans *evm.Transfer, dbTokens ...*evm.Token) (*evm.Token, error) {
	var newT *evm.Token
	var err error
	if len(dbTokens) >= 2 {
		newT = dbTokens[1]
		if newT == nil {
			err = db.ErrDBNotFound
		}
	} else {
		newT, err = c.EvmTokenDb.GetToken(trans.ContractAddr, util.AddressConvert(trans.From), trans.TokenID)
	}
	return newT, err
}

func TransferBatch(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	event, err := data.GetEvent("TransferBatch")
	if err != nil {
		log.Error("TransferBatch data.GetEvent(\"TransferBatch\")", "err", err)
		return nil, err
	}
	records := make([]db.Record, 0)

	if ids, ok := event["ids"].([]interface{}); ok {
		if vals, ok := event["values"].([]interface{}); ok {
			if len(ids) != len(vals) {
				log.Error("len(ids) != len(vals)")
				return nil, fmt.Errorf("len(ids) != len(vals)")
			}
			to := convert.ToString(event["to"])
			from := convert.ToString(event["from"])
			operator := convert.ToString(event["operator"])

			idsStr := make([]string, 0, len(ids))
			for i := range ids {
				idsStr = append(idsStr, convert.ToString(ids[i]))
			}
			toTokens, err := c.EvmTokenDb.GetTokenMap(convert.ToString(data["contract_addr"]), to, idsStr)
			if err != nil {
				return nil, err
			}
			fromTokens, err := c.EvmTokenDb.GetTokenMap(convert.ToString(data["contract_addr"]), to, idsStr)
			if err != nil {
				return nil, err
			}

			for i := range ids {
				trans := &evm.Transfer{}
				trans.TokenID = idsStr[i]
				trans.Amount = convert.ToInt64(vals[i])
				trans.To = util.AddressConvert(to)
				trans.From = util.AddressConvert(from)
				trans.Operator = operator
				trans.TokenType = convert.ToString(ERC1155)

				trans.LoadBlockData(data)
				records = append(records, transferToken(c, op, trans, fromTokens[trans.TokenID], toTokens[trans.TokenID])...)
			}
		}
	}
	return records, nil
}
