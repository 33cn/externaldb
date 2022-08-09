package stat

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli"
)

// Stat 统计
type Stat interface {
	db.DBSaver
	Stat(detail *types.BlockDetail, op int) ([]db.Record, error)
	Recover(client escli.ESClient, lastSeq int64) error
}
