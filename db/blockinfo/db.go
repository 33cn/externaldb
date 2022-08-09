package blockinfo

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

const (
	TableName = "block_info"
)

type BlockInfo struct {
	Height    int64  `json:"height"`
	TxCount   int    `json:"tx_count"`
	Hash      string `json:"hash"`
	BlockTime int64  `json:"block_time"`
	From      string `json:"from"`
}

// BlockRecord 用于db记录区块基本信息
type BlockRecord struct {
	*db.IKey
	*db.Op
	Block *BlockInfo
}

// Value impl
func (r *BlockRecord) Value() []byte {
	v, _ := json.Marshal(r.Block)
	return v
}

// NewBlockRecord create BlockRecord
func NewBlockRecord(block *BlockInfo, op int) *BlockRecord {
	return &BlockRecord{
		IKey:  db.NewIKey(TableName, TableName, block.Hash),
		Op:    db.NewOp(op),
		Block: block,
	}
}
