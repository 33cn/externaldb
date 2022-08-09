package filepart

import (
	"errors"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	fpdb "github.com/33cn/externaldb/db/filepart/db"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
)

var log = l.New("module", "db.file_part")

// Convert tx convert
type Convert struct {
	symbol string
	title  string

	ConfigDB proofconfig.PrivilegeDB
}

// ConvertTx impl
func (c *Convert) convertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	tx := env.Block.Block.Txs[env.TxIndex]

	records := make([]db.Record, 0)

	// 签名地址是否有权限,没有直接返回
	if !c.ConfigDB.IsHaveProofPermission(tx.From()) {
		log.Error("IsHaveProofPermission", "err", errors.New("ErrNoPermission"))
		return records, nil
	}

	rdFilePart := fpdb.NewRecordFilePart(op, &fpdb.FilePart{Data: string(tx.Payload), TxHash: common.ToHex(tx.Hash())})
	records = append(records, rdFilePart)
	return records, nil
}
