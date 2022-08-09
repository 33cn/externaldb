package db

import (
	"encoding/json"
	"errors"

	"github.com/33cn/externaldb/escli/querypara"

	"github.com/33cn/chain33/types"
)

// TxEnv record block info
type TxEnv struct {
	Block     *types.BlockDetail
	TxIndex   int64
	BlockHash string
}

// HeightIndex HeightIndex
func HeightIndex(height, index int64) int64 {
	return height*types.MaxTxsPerBlock + index
}

// Convert for tx convert
type Convert interface {
	ConvertTx(env *TxEnv, op int) ([]Record, error)
}

// WrapDB 满足虚拟合约执行时访问数据库的要求.
// 虚拟合约 由于不在链上执行, 插件需要有一定的业务逻辑, 对业务逻辑的处理需要依赖于现有数据
// 所以需要插件能访问数据库, 目前的逻辑可以通过 Set/Get 来完成. 如果需要更复杂的方式可以扩展这个接口
type WrapDB interface {
	Get(k1, k2, id string) (*json.RawMessage, error)
	List(k1, k2 string, keyValue []*ListKV) ([]*json.RawMessage, error)
	Set(k1, k2, id string, r Record) error
	Search(idx, typ string, query *querypara.Query, decode func(x *json.RawMessage) (interface{}, error)) ([]interface{}, error)
}

// ListKV ListKV
type ListKV struct {
	Key   string
	Value interface{}
}

// 怎么解决有些插件不需要 WrapDB
// if c, ok := convert.(NeedWrapDB); ok {
//    c.SetDB(wrapDB)
// }
// 需要数据库访问的插件 实现 ConvertTx 时 判断 c.db == nil是否有设置, 这样不需要扩展 Convert 接口, 也不破坏现有的插件

// NeedWrapDB 看插件是否需要数据库
type NeedWrapDB interface {
	SetDB(w WrapDB) error
}

// ErrDBNotFound DB Not Found
var (
	ErrDBNotFound         = errors.New("DB Not Found")
	ErrDBBadParam         = errors.New("DB Bad Param")
	ErrDBInvalidOperation = errors.New("DB invalid operation")
)
