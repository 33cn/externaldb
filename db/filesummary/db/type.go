package db

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// FileInfo block tx payload
type FileInfo struct {
	FileType  string `json:"file_type"`
	FileHash  string `json:"file_hash"`
	FileSize  int64  `json:"file_size"`
	PartType  int32  `json:"part_type"` // 1:base64
	PartHashs []byte `json:"part_hashs"`
}

// BlacklistInfo block tx payload
type BlacklistInfo struct {
	FileHash string `json:"file_hash"`
	Note     string `json:"note"`
	Operate  int64  `json:"operate"`
}

type Blacklist struct {
	FileBlacklistFlag bool   `json:"file_blacklist_flag"`
	FileBlacklist     string `json:"file_blacklist"`
	FileBlacklistNote string `json:"file_blacklist_note"`
}

// FileSummary 文件汇总
type FileSummary struct {
	FileInfo
	Blacklist
	*db.Block
}

// Key 主键
func (s *FileSummary) Key() string {
	return AddKeyPrefix(s.FileHash)
}

// RecordKey 文件汇总
func (s *FileSummary) RecordKey() *db.IKey {
	return db.NewIKey(TableName, TableName, s.Key())
}

// NewSummary new summary
func NewSummary(p FileInfo, env *db.TxEnv) *FileSummary {
	var sum FileSummary
	sum.FileInfo = p
	sum.Block = db.NewBlock(env)
	sum.FileBlacklistFlag = false
	sum.FileBlacklist = ""
	sum.FileBlacklistNote = ""
	return &sum
}

// RecordSummary record
type RecordSummary struct {
	*db.IKey
	*db.Op
	value *FileSummary
}

// Value value
func (r *RecordSummary) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}

// NewRecordSummary new record
func NewRecordSummary(op int, value *FileSummary) *RecordSummary {
	return &RecordSummary{
		IKey:  value.RecordKey(),
		Op:    db.NewOp(op),
		value: value,
	}
}
