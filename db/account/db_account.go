package account

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

// TODO 更新时， 开block-height 和 tx-index的逻辑

// 同一帐号存在多个合约对他进行改动， 即资产生成合约(如token， coins)和资产使用合约(如trade，unfreeze， game)
// 故在account 结构中添加 block-height 和 tx-index， 期望在不同合约的展开模块在不同进程中运行时， 老的数据不会覆盖新的数据

// Account 数据结构分开两部分， 是因为信息需要从两部分获得
// 一部分从执行日志， 一部分需要交易信息
// Debail 不指定 `json` tag， 可以在展开成json时， 不会有多级结构， 在外观上保存一致

var (
	// 可以通过 withdraw/transferToExec 收集执行器地址
	// 但不初始化的话， 执行器帐号地址在收集之前，会被标识为个人帐号
	execAddress = make(map[string]bool)
)

// Init 设置配置
func Init(title string, addresses []string) {
	for _, a := range addresses {
		if strings.HasPrefix(title, "user.p.") {
			a = title + a
		}
		addr := db.ExecAddress(a)
		execAddress[addr] = true
	}
}

// AddExecAddress add exec
func AddExecAddress(address string) {
	execAddress[address] = true
}

// account ypte
const (
	AccountPersonage        = "personage"
	AccountContract         = "contract"
	AccountContractInternal = "contractInternal"
)

// Detall  id: address-exec-asset
type Detall struct {
	Address  string `json:"address"`
	Exec     string `json:"exec"` // 执行器名 和 执行器地址 等效， 可能不知道执行器名， 这时填地址
	Frozen   int64  `json:"frozen"`
	Balance  int64  `json:"balance"`
	Total    int64  `json:"total"`
	Type     string `json:"type"` // contract, personage
	AddrType string `json:"addr_type"`
}

// Account 记录帐号信息， Detail 解析自执行日志， 其他项在交易解析时获得
type Account struct {
	*Detall
	HeightIndex int64  `json:"height_index"`
	AssetSymbol string `json:"asset_symbol"`
	AssetExec   string `json:"asset_exec"`
	Height      int64  `json:"height"`
	BlockTime   int64  `json:"block_time"`
}

// Record 同时负责 log解析后的状态保存和状态合并后的状态
type Record struct {
	*db.IKey
	*db.Op
	Acc Account
}

// Value impl
func (r *Record) Value() []byte {
	v, _ := json.Marshal(r.Acc)
	return v
}

// default asset db
const (
	DBX                 = "account"
	TableX              = "account"
	RAccountX           = "account"
	AccountRecordDBX    = "account_record"
	AccountRecordTableX = "account_record"
	DefaultType         = "_doc"
)

// asset
const (
	ExecCoinsX  = "coins"
	ExecCoinsxX = "coinsx"
)

// NewAccountKey NewAccountKey
func NewAccountKey(id string) *db.IKey {
	return db.NewIKey(DBX, TableX, id)
}

// NewAccountRecordKey New AccountRecordKey
func NewAccountRecordKey(id string) *db.IKey {
	return db.NewIKey(AccountRecordDBX, AccountRecordTableX, id)
}

// Key for index id
func (acc *Account) Key() string {
	acc.Address = util.AddressConvert(acc.Address)
	if acc.Exec == "" {
		return fmt.Sprintf("%s-%s-%s:%s", acc.Address, acc.AssetExec, acc.AssetExec, acc.AssetSymbol)
	}
	return fmt.Sprintf("%s-%s-%s:%s", acc.Address, acc.Exec, acc.AssetExec, acc.AssetSymbol)
}

// RecordKey for index id
func (acc *Account) RecordKey() string {
	acc.Address = util.AddressConvert(acc.Address)
	if acc.Exec == "" {
		return fmt.Sprintf("%s-%s-%s-%d:%s", acc.Address, acc.AssetExec, acc.AssetExec, acc.Height, acc.AssetSymbol)
	}
	return fmt.Sprintf("%s-%s-%s-%d:%s", acc.Address, acc.Exec, acc.AssetExec, acc.Height, acc.AssetSymbol)
}
