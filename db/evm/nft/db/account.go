package db

import (
	"fmt"

	"github.com/33cn/externaldb/db"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

type Account struct {
	OwnerAddr string `json:"owner_addr"`
	LabelID   string `json:"label_id"`
	GoodsID   int64  `json:"goods_id"`
	Balance   int64  `json:"balance"`
	EvmState
}

// Key for account index id
func (acc *Account) Key() string {
	return fmt.Sprintf("account-%s-%s-%v", acc.OwnerAddr, acc.ContractAddr, acc.GoodsID)
}

// AccountBalance 账户余额event结构 由于使用时是数组形式，所以地址类型没有被解析成字符串
type AccountBalance struct {
	Account pcom.Hash160Address `json:"account"`
	GoodsID int64               `json:"goodsID"`
	Balance int64               `json:"balance"`
}

func (ab AccountBalance) GetAccount(info map[string]interface{}) *Account {
	acc := &Account{
		OwnerAddr: db.Hash160AddressToString(ab.Account),
		GoodsID:   ab.GoodsID,
		Balance:   ab.Balance,
	}
	acc.GetEvmState(info)
	return acc
}
