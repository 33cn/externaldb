package main

import (
	"testing"

	"github.com/33cn/externaldb/db/block"
)

func getLastSeq(prefix, id string) int64 {
	syncSeq := GetLastSeq("http://172.16.101.87:9200", prefix, id, 7, "elastic", "elastic")
	return syncSeq
}

func TestGetLastSeq(t *testing.T) {
	lastSeq := LastSeq{}
	lastSeq.LastSyncSeq = getLastSeq("v12seq01_", block.SyncSeq)
	lastSeq.LastConvertSeq = getLastSeq("v12db02_", "convert_bty3")

	t.Log(lastSeq)
}
