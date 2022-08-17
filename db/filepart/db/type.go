package db

import (
	"encoding/json"
	"time"

	"github.com/33cn/externaldb/db"
)

// FilePart 文件分片
type FilePart struct {
	Data       string `json:"data"`
	TxHash     string `json:"tx_hash"`
	InsertTime int64  `json:"insert_time"`
}

// Key 主键
func (f *FilePart) Key() string {
	return AddKeyPrefix(f.TxHash)
}

// RecordKey record key
func (f *FilePart) RecordKey() *db.IKey {
	return db.NewIKey(TableName, TableName, f.Key())
}

// RecordFilePart record
type RecordFilePart struct {
	*db.IKey
	*db.Op
	value *FilePart
}

// Value value
func (r *RecordFilePart) Value() []byte {
	if r.value.InsertTime == 0 {
		r.value.InsertTime = time.Now().Unix()
	}
	v, _ := json.Marshal(r.value)
	return v
}

// NewRecordFilePart new RecordFilePart
func NewRecordFilePart(op int, value *FilePart) *RecordFilePart {
	return &RecordFilePart{
		IKey:  value.RecordKey(),
		Op:    db.NewOp(op),
		value: value,
	}
}
