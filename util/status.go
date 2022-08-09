package util

import (
	"encoding/json"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
)

var log = l.New("module", "util")

// dbStatus//lastHeader  hash
// dbStatus//Header -> keys[prev,current]

// GetLastSyncSeq 获得最后同步的seq
func LastSyncSeq(client escli.ESClient, id string) (int64, error) {
	result, err := client.Get(db.LastSeqDB, db.LastSeqDB, id)
	if err != nil {
		if err == db.ErrDBNotFound {
			// 第一个连上
			return -1, nil
		}
		return 0, err
	}
	var seq block.LastSyncSeq
	err = json.Unmarshal([]byte(*result), &seq)
	if err != nil {
		return 0, err
	}

	return seq.SyncSeq, nil
}

//func SaveLastSyncSeq(client *escli.ESClient, id string, seq int64) error {
//	record := newLastRecord(db.LastSeqDB, db.LastSeqDB, id, seq)
//	return client.Update(record.Index(), record.Type(), record.ID(), string(record.Value()))
//}

// NewLastRecord create LastSyncSeqRecord
func NewLastRecord(id string, seq int64) *block.LastSyncSeqRecord {
	s := &block.LastSyncSeq{SyncSeq: seq}
	return &block.LastSyncSeqRecord{
		IKey: db.NewIKey(db.LastSeqDB, db.LastSeqDB, id),
		Op:   db.NewOp(db.OpAdd),
		Seq:  s,
	}
}

// Save 保存convert后的结果到ES
func SaveToES(client escli.ESClient, blockItems []db.Record) error {
	if len(blockItems) == 0 {
		return nil
	}
	var rs []db.Record
	for i, v := range blockItems {
		rs = append(rs, v)
		log.Debug("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}
	return client.BulkUpdate(rs)
}

func SaveToESSelectBulk(client escli.ESClient, blockItems []db.Record, bulk bool) error {
	if len(blockItems) == 0 {
		return nil
	}
	if bulk {
		var rs []db.Record
		for i, v := range blockItems {
			rs = append(rs, v)
			log.Info("SaveToESSelectBulk bulk", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
		}
		return client.BulkUpdate(rs)
	}

	for i, v := range blockItems {
		err := client.Update(v.Index(), v.Type(), v.ID(), string(v.Value()))
		log.Info("SaveToESSelectBulk", "idx", i, "ID", v.Key(), "value", string(v.Value()))
		if err != nil {
			log.Error("SaveToESSelectBulk", "idx", i, "ID", v.Key(), "err", err)
			return err
		}
	}
	return nil
}
