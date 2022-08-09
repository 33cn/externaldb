package token

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// token db
const (
	TokenTxDB   = "token_tx"   //token_tx/token_tx/{tx_hash}
	TokenInfoDB = "token_info" //token_info/tokeninfo/{symbol}
	DefaultType = "_doc"

	ExecTokenX = "token"
)

// Record db token
type Record struct {
	*db.IKey
	*db.Op
	value interface{}
}

// Value impl
func (r *Record) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}

// RecordToken db token
type RecordToken struct {
	*db.IKey
	*db.Op
	value *Token
}

// Value impl
func (r *RecordToken) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}
