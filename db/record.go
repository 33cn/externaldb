package db

import (
	"fmt"

	"github.com/33cn/chain33/common"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
)

var version int32 = 6

// Key 获得相关的key
type Key interface {
	Index() string
	Type() string
	ID() string
	Key() string
}

// Operator for db Operator
type Operator interface {
	OpType() int
}

// Record 要保存到外部数据的数据结构都要满足这个接口
type Record interface {
	Key
	Operator
	Value() []byte
}

type SourceAbleRecord interface {
	Record
	Source() interface{}
}

// SeqType
const (
	SeqTypeAdd = 1
	SeqTypeDel = 2
)

// DB operator
const (
	OpAdd    = 1
	OpDel    = 2
	OpUpdate = 3
)

// DB Type
const (
	DatabaseTypeEs        = 1
	DatabaseTypeChainJRPC = 2
	DatabaseTypeChainGRPC = 3
)

// IKey impl Key
type IKey struct {
	index, typ, id string
}

func SetVersion(v int32) {
	version = v
}

// Index impl Key
func (d *IKey) Index() string {
	return d.index
}

// Type impl Key
func (d *IKey) Type() string {
	return d.typ
}

// ID impl Key
func (d *IKey) ID() string {
	return d.id
}

// Key impl Key
func (d *IKey) Key() string {
	return fmt.Sprintf("%s/%s/%s", d.index, d.typ, d.id)
}

// NewIKey create key
func NewIKey(index, typ, id string) *IKey {
	switch version {
	case 6:
		return &IKey{index: index, typ: typ, id: id}
	case 7:
		return &IKey{index: index, typ: "_doc", id: id}
	default:
		panic("not support es version" + fmt.Sprint(version) + "about NewIKey")
	}
}

// Op impl Op
type Op struct {
	opType int
}

// Op impl Op
func (d *Op) OpType() int {
	return d.opType
}

// NewOp create Op
func NewOp(t int) *Op {
	return &Op{opType: t}
}

// Block info
type Block struct {
	Height      int64  `json:"height"`
	Ts          int64  `json:"ts"`
	BlockHash   string `json:"block_hash"`
	Index       int64  `json:"index"`
	Send        string `json:"send"`
	TxHash      string `json:"tx_hash"`
	HeightIndex int64  `json:"height_index"`
}

// SetupBlock set into to block
func SetupBlock(env *TxEnv, from, txHash string) *Block {
	return &Block{
		Height:      env.Block.Block.Height,
		Ts:          env.Block.Block.BlockTime,
		BlockHash:   env.BlockHash,
		Index:       env.TxIndex,
		Send:        from,
		TxHash:      txHash,
		HeightIndex: HeightIndex(env.Block.Block.Height, env.TxIndex),
	}
}

// NewBlock New
func NewBlock(env *TxEnv) *Block {
	tx := env.Block.Block.Txs[env.TxIndex]
	return SetupBlock(env, tx.From(), common.ToHex(tx.Hash()))
}

// NewBlockByTxDetail New
func NewBlockByTxDetail(txd *rpctypes.TransactionDetail) *Block {
	return &Block{
		Height: txd.Height,
		Ts:     txd.Blocktime,
		//BlockHash:   common.ToHex(txd.Tx.Hash()),
		Index:       txd.Index,
		Send:        txd.Fromaddr,
		TxHash:      txd.Tx.Hash,
		HeightIndex: HeightIndex(txd.Height, txd.Index),
	}
}

// NewBlockByTxDetail2 New
func NewBlockByTxDetail2(txd *types.TransactionDetail) *Block {
	return &Block{
		Height: txd.Height,
		Ts:     txd.Blocktime,
		//BlockHash:   common.ToHex(txd.Tx.Hash()),
		Index:       txd.Index,
		Send:        txd.Fromaddr,
		TxHash:      common.ToHex(txd.Tx.Hash()),
		HeightIndex: HeightIndex(txd.Height, txd.Index),
	}
}
