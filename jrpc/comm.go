package main

import (
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/util"
)

// Comm Comm
type Comm struct {
	syncDB    *DBRead
	convDB    *DBRead
	ConvertID string
	Version   int32
}

type LastSeq struct {
	LastSyncSeq    int64 `json:"lastSyncSeq"`
	LastConvertSeq int64 `json:"lastConvertSeq"`
}

func (c *Comm) LastSeq(out *interface{}) error {
	lastSeq := LastSeq{}
	lastSeq.LastSyncSeq = GetLastSeq(c.syncDB.Host, c.syncDB.Prefix, block.SyncSeq, c.Version, c.syncDB.Username, c.syncDB.Password)
	lastSeq.LastConvertSeq = GetLastSeq(c.convDB.Host, c.convDB.Prefix, c.ConvertID, c.Version, c.convDB.Username, c.convDB.Password)

	*out = lastSeq
	return nil
}

//GetLastSeq 获取已经同步或者解析的最新seq值
func GetLastSeq(host, prefix, id string, version int32, user, pwd string) int64 {
	cli, err := escli.NewESShortConnect(host, prefix, version, user, pwd)
	if err != nil {
		return -1
	}

	num, err := util.LastSyncSeq(cli, id)
	if err != nil {
		log.Error("lastSeq:LastSyncSeq", "err", err)
		return -1
	}
	return num
}
