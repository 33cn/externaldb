package trade

import (
	"encoding/json"
)

// Value impl
func (r *dbTxOrder) Value() []byte {
	v, _ := json.Marshal(r.tx)
	return v
}

// Value impl
func (r *dbOrder) Value() []byte {
	v, _ := json.Marshal(r.order)
	return v
}

// Value impl
func (r *dbAsset) Value() []byte {
	v, _ := json.Marshal(r.asset)
	return v
}
