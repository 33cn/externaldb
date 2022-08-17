package store

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
)

// SeqNum ...
type SeqNum struct {
	// 同协议中的seq
	Number int64
	// 对应的区块高度
	Height int64
	// 从哪里获得
	From string
}

// Seq ...
type Seq types.BlockSeqs

//
type SeqNumStore interface {
	LastSeq() (*SeqNum, error)
	UpdateLastSeq(s db.Record) error
}

// SeqStore SeqStore
type SeqStore interface {
	SaveSeqs(blockItems []db.Record) error
	GetSeq(num int64) (*block.Seq, error)
	CommitSeqAck(num int64) int64
}
