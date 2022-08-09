package unfreeze

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// unfreeze db
const (
	UnfreezeTxDBX     = "unfreeze_tx"
	UnfreezeTxTableX  = "unfreeze"
	DefaultTable      = "_doc"
	UnfreezeSeqDBX    = "unfreeze_seq"
	UnfreezeSeqTableX = "seq"
	UnfreezeLastSeqX  = "last_seq"
)

// TxRecord 用于db 记录
type TxRecord struct {
	*db.IKey
	*db.Op
	tx *Tx
}

// Value impl
func (r *TxRecord) Value() []byte {
	v, _ := json.Marshal(r.tx)
	return v
}
