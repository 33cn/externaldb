package transaction

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// unfreeze db
const (
	TransactionX  = "transaction"
	DefaultType   = "_doc"
	RTransactionX = "transaction"
)

// TxRecord 用于db 记录
type TxRecord struct {
	*db.IKey
	*db.Op
	Tx *Transaction
}

// Value impl
func (r *TxRecord) Value() []byte {
	v, _ := json.Marshal(r.Tx)
	return v
}

// NewTransactionKey NewTransactionKey
func NewTransactionKey(id string) *db.IKey {
	return db.NewIKey(TransactionX, TransactionX, id)
}
