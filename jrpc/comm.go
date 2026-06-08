package main

import (
	"strings"

	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
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

// addrKeys 地址类型字段，其值如果是 0x 开头需统一小写
var addrKeys = map[string]bool{
	"from":             true,
	"to":               true,
	"contract_addr":    true,
	"contract_address": true,
	"owner_addr":       true,
	"owner_address":    true,
	"creator":          true,
}

// normalizeAddrInQuery 将 query 中地址类型字段的 0x 值统一转为小写
// EVM 地址大小写不敏感，但 ES 查询大小写敏感，统一小写避免查询不到
func normalizeAddrInQuery(q *querypara.Query) {
	if q == nil {
		return
	}
	normalizeMatch := func(matches []*querypara.QMatch) {
		for _, m := range matches {
			if m == nil {
				continue
			}
			if addrKeys[m.Key] {
				if s, ok := m.Value.(string); ok && strings.HasPrefix(s, "0x") {
					m.Value = strings.ToLower(s)
				}
			}
			if m.SubQuery != nil {
				normalizeAddrInQuery(m.SubQuery)
			}
		}
	}
	normalizeMatch(q.Match)
	normalizeMatch(q.MatchOne)
	normalizeMatch(q.Filter)
	normalizeMatch(q.Not)
}
