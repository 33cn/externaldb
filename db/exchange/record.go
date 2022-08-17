package exchange

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// exchange db
const (
	ExchangeTxDB   = "exchange_tx"   //exchange_tx/exchange_tx/{tx_hash}
	ExchangeInfoDB = "exchange_info" //exchange_info/exchangeinfo/{symbol}
	DefaultType    = "_doc"

	ExecExchangeX = "exchange"
)

// Record db exchange
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

// RecordExchange db exchange
type RecordExchange struct {
	*db.IKey
	*db.Op
	value *Exchange
}

// Value impl
func (r *RecordExchange) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}
