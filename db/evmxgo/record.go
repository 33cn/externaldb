package evmxgo

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// evmxgo db
const (
	EvmxgoTxDB   = "evmxgo_tx"   //evmxgo_tx/evmxgo_tx/{tx_hash}
	EvmxgoInfoDB = "evmxgo_info" //evmxgo_info/evmxgoinfo/{symbol}
	DefaultType  = "_doc"

	ExecEvmxgoX = "evmxgo"
)

// Record db evmxgo
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

// RecordEvmxgo db evmxgo
type RecordEvmxgo struct {
	*db.IKey
	*db.Op
	value *Evmxgo
}

// Value impl
func (r *RecordEvmxgo) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}
