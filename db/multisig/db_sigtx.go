package multisig

import (
	"fmt"
)

// Sig 分两部分：逻辑上的交易， 每个交易  上的贡献weight 的列表

// Sig Tx Type
const (
	SigTxTypeOwner    = "owner"
	SigTxTypeAccount  = "account"
	SigTxTypeTransfer = "transfer"
)

// SigTx SigTx
type SigTx struct {
	TxHash string `json:"tx_hash"`
	// 不同类型的多重签名
	Type    string      `json:"type"`
	Address string      `json:"multi_signature_address"`
	TxID    uint64      `json:"tx_id"`
	Detail  interface{} `json:"detail"`
}

// TxOwnerOperate OwnerOperate
type TxOwnerOperate struct {
	Operate   string `json:"operate"`
	OldOwner  string `json:"old_owner"`
	NewOwner  string `json:"new_owner"`
	NewWeight uint64 `json:"new_weight"`
}

func ownerOpStr(op uint64) string {
	/*
		OwnerAdd     uint64 = 1
		OwnerDel     uint64 = 2
		OwnerModify  uint64 = 3
		OwnerReplace uint64 = 4
	*/
	switch op {
	case 1:
		return "add"
	case 2:
		return "delete"
	case 3:
		return "modify"
	case 4:
		return "replace"
	}
	return "unknown"
}

// SymbolLimit SymbolLimit
type SymbolLimit struct {
	Symbol     string `json:"symbol"`
	Execer     string `json:"execer"`
	DailyLimit uint64 `json:"daily_limit"`
}

// TxAccountOperate TxAccountOperate
type TxAccountOperate struct {
	DailyLimit        *SymbolLimit `json:"dailyLimit,omitempty"`
	NewRequiredWeight uint64       `json:"new_required_weight,omitempty"`
	Operate           string       `json:"operate"`
}

func accountOpStr(op bool) string {
	/*
		AccWeightOp     = true
		AccDailyLimitOp = false
	*/
	if op {
		return "weight_operate"
	}
	return "limit_operate"
}

// TxTransferOperate TxTransferOperate
type TxTransferOperate struct {
	Symbol   string `json:"symbol"`
	Amount   int64  `json:"amount"`
	Note     string `json:"note"`
	Execname string `json:"execname"`
	To       string `json:"to"`
	From     string `json:"from"`
}

// ID record id
func (s *SigTx) ID() string {
	return fmt.Sprintf("tx-%s:%08d", s.Address, s.TxID)
}

// SigList 不包括撤销的
type SigList struct {
	Address string `json:"multi_signature_address"`
	TxID    uint64 `json:"tx_id"`
	Owner   string `json:"address"`
	Weight  uint64 `json:"weight"`
	// 角色和状态: 是否为创建者，是否是促成交易执行的
	Creator  bool   `json:"creator"`
	Executed bool   `json:"executed"`
	TxHash   string `json:"tx_hash"`
}

// ID record id
func (s *SigList) ID() string {
	return fmt.Sprintf("sig_list-%s:%08d:%s", s.Address, s.TxID, s.Owner)
}

// 现在暂时没有需求
// TxOption  multi signature tx
