package erc721

import (
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/evm"
	"github.com/33cn/go-kit/convert"
)

const ERC721 = "ERC721"

var (
	log = l.New("module", "db.evm.erc721")
)

func init() {

	// ApprovalForAll 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31 event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
	// Transfer 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	// Approval 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)

	// Transfer 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	// event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	evm.RegisterEventHandle(ERC721+"Transfer", Transfer)
}

func Transfer(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	event, err := data.GetEvent("Transfer")
	if err != nil {
		log.Error("Transfer data.GetEvent(\"TransferBatch\")", "err", err)
		return nil, err
	}
	trans := &evm.Transfer{}
	trans.TokenID = convert.ToInt64(event["tokenId"])
	trans.From = convert.ToString(event["from"])
	trans.To = convert.ToString(event["to"])
	trans.Operator = convert.ToString(data["from_addr"])
	trans.Amount = 1
	trans.TokenType = convert.ToString(ERC721)
	trans.LoadBlockData(data)

	return transferToken(c, op, trans), nil
}

func transferToken(c *evm.Convert, op int, trans *evm.Transfer) []db.Record {
	records := make([]db.Record, 0)

	// 转账记录
	if op == db.SeqTypeAdd {
		records = append(records, evm.NewRecordTransfer(trans, db.OpAdd))
	} else { // 回滚
		records = append(records, evm.NewRecordTransfer(trans, db.OpDel))
	}

	// 维护token
	// 发币
	if trans.From == evm.EmptyAddr || trans.From == evm.EmptyAddrEth {
		if op == db.SeqTypeAdd {
			records = append(records, evm.NewRecordToken(trans.GetNewToken(), db.OpAdd))
		} else { // 回滚
			records = append(records, evm.NewRecordToken(trans.GetNewToken(), db.OpDel))
		}
		return records
	}

	// 转账
	if op == db.SeqTypeAdd {
		oldT, err := c.EvmTokenDb.GetToken(trans.ContractAddr, trans.From, trans.TokenID)
		if err != nil {
			log.Error("transferToken GetToken", "err", err)
			return records
		}
		records = append(records, evm.NewRecordToken(oldT, db.OpDel))
		oldT.Owner = trans.To
		records = append(records, evm.NewRecordToken(oldT, db.OpAdd))
	} else { // 回滚
		oldT, err := c.EvmTokenDb.GetToken(trans.ContractAddr, trans.To, trans.TokenID)
		if err != nil {
			log.Error("transferToken GetToken", "err", err)
			return records
		}
		records = append(records, evm.NewRecordToken(oldT, db.OpDel))
		oldT.Owner = trans.From
		records = append(records, evm.NewRecordToken(oldT, db.OpAdd))
	}
	return records
}
