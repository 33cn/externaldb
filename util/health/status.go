package health

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/status"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/version"
)

var log = log15.New("module", "health")

const (
	MethodChain33Version           = "Chain33.Version"
	MethodChain33GetPushSeqLastNum = "Chain33.GetPushSeqLastNum"
	MethodChain33GetCoinSymbol     = "Chain33.GetCoinSymbol"
)

// Status 服务状态详细信息
type Status struct {
	Server *ServerStatus  `json:"server"`
	Chain  *ChainStatus   `json:"chain"`
	ES     *status.Status `json:"es"`
}

// ServerStatus 当前服务状态信息
type ServerStatus struct {
	Version string `json:"version"`
	SyncSeq int64  `json:"sync_seq"` // 同步序列高度
	ConvSeq int64  `json:"conv_seq"` // 转换序列高度
	Title   string `json:"title"`
	Coin    string `json:"coin"`
}

// ChainStatus 区块链状态信息
type ChainStatus struct {
	Status  string            `json:"status"`
	PushSeq int64             `json:"push_seq"` // 推送高度
	Coin    string            `json:"coin"`     // 主代币信息
	Version *ChainVersionInfo `json:"version"`
}

// ChainVersionInfo 区块链版本信息
type ChainVersionInfo struct {
	Title   string `json:"title"`   // 区块链名，该节点 chain33.toml 中配置的 title 值
	App     string `json:"app"`     // 应用 app 的版本
	Chain33 string `json:"chain33"` // 版本信息，版本号-GitCommit（前八个字符）
	LocalDB string `json:"localDb"` // localdb 版本号
}

// ChainCommonStringMessage 区块链通用字符串消息
type ChainCommonStringMessage struct {
	Data string `json:"data"`
}

// ChainCommonIntMessage 区块链通用整形消息
type ChainCommonIntMessage struct {
	Data int64 `json:"data"`
}

// GetStatus 获取服务详细状态信息
func GetStatus(c *proto.ConfigNew) *Status {
	var s Status
	s.Server = GetServerStatus(c)
	s.Chain = GetChainStatus(c.Chain.Host, c.Sync.PushName)
	s.ES = GetElasticSearchStatus(c.ConvertEs, c.EsVersion)
	return &s
}

// GetServerStatus 获取当前服务状态信息
func GetServerStatus(c *proto.ConfigNew) *ServerStatus {
	var ss ServerStatus
	ss.Version = version.GetVersion()
	ss.SyncSeq = LastSyncSeq(c)
	ss.ConvSeq = LastConvertSeq(c)
	ss.Coin = c.Chain.Symbol
	ss.Title = c.Chain.Title
	return &ss
}

// GetChainStatus 获取区块链状态信息
func GetChainStatus(host, pushName string) *ChainStatus {
	var cs ChainStatus
	cs.Status = "UP"
	// 获取版本 Version
	// https://chain.33.cn/document/97#1.1%20%20%E8%8E%B7%E5%8F%96%E7%89%88%E6%9C%AC%20Version
	cs.Version = &ChainVersionInfo{}
	ctx := jsonclient.NewRPCCtx(host, MethodChain33Version, nil, &cs.Version)
	if _, err := ctx.RunResult(); err != nil {
		log.Error("Health, get Chain33.Version", "err", err)
		cs.Status = err.Error()
	}

	// 获取某推送服务最新序列号的值 GetPushSeqLastNum
	// https://chain.33.cn/document/97#1.12%20%E8%8E%B7%E5%8F%96%E6%9F%90%E6%8E%A8%E9%80%81%E6%9C%8D%E5%8A%A1%E6%9C%80%E6%96%B0%E5%BA%8F%E5%88%97%E5%8F%B7%E7%9A%84%E5%80%BC%20GetPushSeqLastNum
	req := ChainCommonStringMessage{Data: pushName}
	resp := ChainCommonIntMessage{}
	ctx = jsonclient.NewRPCCtx(host, MethodChain33GetPushSeqLastNum, req, &resp)
	if _, err := ctx.RunResult(); err != nil {
		log.Error("Health, get Chain33.GetPushSeqLastNum", "err", err)
		cs.Status = err.Error()
	}
	cs.PushSeq = resp.Data

	// 获取主代币信息 GetCoinSymbol
	// https://chain.33.cn/document/100#1.5%20%E8%8E%B7%E5%8F%96%E4%B8%BB%E4%BB%A3%E5%B8%81%E4%BF%A1%E6%81%AF%20GetCoinSymbol
	resp2 := ChainCommonStringMessage{}
	ctx = jsonclient.NewRPCCtx(host, MethodChain33GetCoinSymbol, nil, &resp2)
	if _, err := ctx.RunResult(); err != nil {
		log.Error("Health, get Chain33.GetPushSeqLastNum", "err", err)
		cs.Status = err.Error()
	}
	cs.Coin = resp2.Data
	return &cs
}

// GetElasticSearchStatus 获取elasticsearch状态信息
func GetElasticSearchStatus(es *proto.ESDB, version int32) *status.Status {
	client, err := escli.NewESShortConnect(es.Host, "", version, es.User, es.Pwd)
	if err != nil {
		log.Error("lastSeq:NewESShortConnect ", "err", err)
		return &status.Status{Status: err.Error()}
	}
	return client.Status()
}

// LastSyncSeq 获取当前sync序列号
func LastSyncSeq(c *proto.ConfigNew) int64 {
	// sync与convert合并后没有此项配置
	if c.SyncEs == nil {
		return -1
	}
	return LastSeq(c.SyncEs.Host, c.SyncEs.Prefix, block.SyncSeq, c.EsVersion, c.SyncEs.User, c.SyncEs.Pwd)
}

// LastConvertSeq 获取当前convert序列号
func LastConvertSeq(c *proto.ConfigNew) int64 {
	return LastSeq(c.ConvertEs.Host, c.ConvertEs.Prefix, c.Convert.AppName, c.EsVersion, c.ConvertEs.User, c.ConvertEs.Pwd)
}

// LastSeq 获取已经同步或者解析的最新seq值
func LastSeq(host, prefix, id string, version int32, username, password string) int64 {

	client, err := escli.NewESShortConnect(host, prefix, version, username, password)
	if err != nil {
		log.Error("lastSeq:NewESShortConnect ", "err", err)
		return -1
	}

	num, err := util.LastSyncSeq(client, id)
	if err != nil {
		log.Error("lastSeq:LastSyncSeq ", "err", err)
		return -1
	}
	return num
}
