package coinsx

import (
	"encoding/json"

	coinTy "github.com/33cn/plugin/plugin/dapp/coinsx/types"
	"github.com/33cn/externaldb/db"
)

const (
	CoinsxManagerDBX    = "coins_manager"
	CoinsxManagerTableX = "coins_manager"
	DefaultType         = "_doc"
)

type Manager struct {
	ManagerStatus *coinTy.ManagerStatus `json:"manager_status"`
	HeightIndex   int64                 `json:"height_index"`
}

type Record struct {
	*db.IKey
	*db.Op
	Manager Manager
}

func (r *Record) Value() []byte {
	v, _ := json.Marshal(r.Manager)
	return v
}

func newCoinsxKey(flag coinTy.TransferFlag) *db.IKey {
	return db.NewIKey(CoinsxManagerDBX, CoinsxManagerTableX, string(flag))
}
