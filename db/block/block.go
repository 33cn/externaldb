package block

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/externaldb/db"
)

// sync status
const (
	StatusDB    = "seq"
	SyncSeq     = "sync_seq"
	DefaultType = "_doc"
)

// LastSyncSeq 同于记录同步的seq
type LastSyncSeq struct {
	SyncSeq int64 `json:"sync_seq"`
}

// Seq 记录同步好的block 信息: id = sync_seq
type Seq struct {
	// 同于记录同步的seq, 如果切换节点可以用上， 不切换节点一直和 Number 一致
	SyncSeq int `json:"sync_seq"`
	// 设置从何处同步的
	From   string `json:"from"`
	Number int    `json:"number"`
	// Seq 的具体信息
	Hash        string `json:"hash"`
	Type        int    `json:"type"`
	BlockDetail []byte `json:"block_detall"`
}

// LastSyncSeqRecord 用于db 记录 LastSyncSeq
type LastSyncSeqRecord struct {
	*db.IKey
	*db.Op
	Seq *LastSyncSeq
}

// SeqRecord 用于db 记录 BlockSeq
type SeqRecord struct {
	*db.IKey
	*db.Op
	seq *Seq
}

// Value impl
func (r *LastSyncSeqRecord) Value() []byte {
	v, _ := json.Marshal(r.Seq)
	return v
}

// Value impl
func (r *SeqRecord) Value() []byte {
	v, _ := json.Marshal(r.seq)
	return v
}

// NewSeqRecord create SeqRecord
func NewSeqRecord(blockSeq *Seq) *SeqRecord {
	return &SeqRecord{
		IKey: db.NewIKey(StatusDB, StatusDB, fmt.Sprintf("%d", blockSeq.SyncSeq)),
		Op:   db.NewOp(db.OpAdd),
		seq:  blockSeq,
	}
}

// NewLastRecord create LastSyncSeqRecord
func NewLastRecord(seq int64) *LastSyncSeqRecord {
	s := &LastSyncSeq{SyncSeq: seq}
	return &LastSyncSeqRecord{
		IKey: db.NewIKey(db.LastSeqDB, db.LastSeqDB, SyncSeq),
		Op:   db.NewOp(db.OpAdd),
		Seq:  s,
	}
}
