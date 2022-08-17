package multisig

import (
	"encoding/json"
	"fmt"

	pty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/33cn/externaldb/db"
)

// MS path

// MSRecord db record
type MSRecord struct {
	*db.IKey
	*db.Op

	m *MS
}

// Value impl
func (r *MSRecord) Value() []byte {
	v, _ := json.Marshal(r.m)
	return v
}

// MSUpdateTxCount MultiSignature Update TxCount
type MSUpdateTxCount struct {
	Address string `json:"multi_signature_address"`
	TxCount uint64 `json:"tx_count"`
}

// MSUpdateTxCountRecord MSUpdateTxCountRecord
type MSUpdateTxCountRecord struct {
	*db.IKey
	*db.Op

	u *MSUpdateTxCount
}

// Value impl
func (r *MSUpdateTxCountRecord) Value() []byte {
	v, _ := json.Marshal(r.u)
	return v
}

// MSUpdateWeight MSUpdateWeight Update weight
type MSUpdateWeight struct {
	Address        string `json:"multi_signature_address"`
	RequiredWeight uint64 `json:"required_weight"`
}

// MSUpdateWeightRecord MSUpdateWeight
type MSUpdateWeightRecord struct {
	*db.IKey
	*db.Op

	u *MSUpdateWeight
}

// Value impl
func (r *MSUpdateWeightRecord) Value() []byte {
	v, _ := json.Marshal(r.u)
	return v
}

// MSUpdateLimitRecord MSUpdateLimitRecord
type MSUpdateLimitRecord struct {
	*db.IKey
	*db.Op

	u *MSLimit
}

// Value impl
func (r *MSUpdateLimitRecord) Value() []byte {
	v, _ := json.Marshal(r.u)
	return v
}

// MSUpdateOwnerRecord MSUpdateOwnerRecord
type MSUpdateOwnerRecord struct {
	*db.IKey
	*db.Op

	u *MSOwner
}

// Value impl
func (r *MSUpdateOwnerRecord) Value() []byte {
	v, _ := json.Marshal(r.u)
	return v
}

// Sig path

// SigTxRecord db record
type SigTxRecord struct {
	*db.IKey
	*db.Op

	m *SigTx
}

func toStringSigTxType(t uint64) string {
	switch t {
	case pty.OwnerOperate:
		return SigTxTypeOwner
	case pty.AccountOperate:
		return SigTxTypeAccount
	case pty.TransferOperate:
		return SigTxTypeTransfer
	}
	return fmt.Sprintf("SigTxType:%d", t)
}

// Value impl
func (r *SigTxRecord) Value() []byte {
	v, _ := json.Marshal(r.m)
	return v
}

// SigListRecord SigListRecord
type SigListRecord struct {
	*db.IKey
	*db.Op

	m *SigList
}

// Value impl
func (r *SigListRecord) Value() []byte {
	v, _ := json.Marshal(r.m)
	return v
}
