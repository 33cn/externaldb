package multisig

import (
	"fmt"
)

// db
const (
	MSDBX       = "multisig"
	MSTxDBX     = "sig_tx"
	MSListDBX   = "sig_list"
	DefaultType = "_doc"

	MSTypeLimitX   = "limit"
	MSTypeOwnerX   = "owner"
	MSTypeAccountX = "account"
)

// 在这个合约中参数4类数据
// 通用account： multiSig 对应的资产集合， 为 account 数据类型： addr， asset
// 通用transaction： 用户行为： multiSigTx：tx-hash，multiSig-TxID(addr)， signer， action type
// multiSig 合约内部对象： addr， owner， assets
// 交易列表：促成多重签名交易生效

// 多重签名帐号结构体展开
// 地址信息 1：1
// owner对地址 n：1
// symbol limit对地址  n：1

// MS short for Multi Signature
type MS struct {
	CreateAddr     string `json:"create_address"`
	MultiSigAddr   string `json:"multi_signature_address"`
	TxCount        uint64 `json:"tx_count"`
	RequiredWeight uint64 `json:"required_weight"`
	Type           string `json:"type"`
}

// ID record
func (o *MS) ID() string {
	return fmt.Sprintf("ms-%s", o.MultiSigAddr)
}

// update tx count & required weight
func msID(address string) string {
	return fmt.Sprintf("ms-%s", address)
}

// MSOwner owner with MultiSigAddr
type MSOwner struct {
	MultiSigAddr string `json:"multi_signature_address"`
	Address      string `json:"address"`
	Weight       uint64 `json:"weight"`
	Type         string `json:"type"`
}

// ID record
func (o *MSOwner) ID() string {
	return fmt.Sprintf("owner-%s-%s", o.MultiSigAddr, o.Address)
}

// MSLimit limit with MS address
type MSLimit struct {
	MultiSigAddr string `json:"multi_signature_address"`
	Symbol       string `json:"symbol"`
	Execer       string `json:"execer"`
	DailyLimit   uint64 `json:"daily_limit"`
	SpentToday   uint64 `json:"spent_today"`
	LastDay      int64  `json:"last_day"`
	Type         string `json:"type"`
}

// ID record
func (l *MSLimit) ID() string {
	return fmt.Sprintf("limit-%s-%s-%s", l.MultiSigAddr, l.Execer, l.Symbol)
}
