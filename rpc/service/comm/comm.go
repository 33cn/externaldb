// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package comm

import (
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/rpc/middleware/protocol"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/util/health"
	"github.com/gin-gonic/gin"
)

var (
	log = l.New("module", "comm")
)

// InitRouter 初始化proofrpc接口的router路由表
func InitRouter(router gin.IRouter, cfg *proto.ConfigNew) {
	comm := Comm{}
	comm.SyncEs = cfg.SyncEs
	comm.ConvertEs = cfg.ConvertEs
	comm.ConvertID = cfg.Convert.AppName
	comm.Version = cfg.EsVersion
	comm.Config = cfg

	v1 := router.Group("/v1")
	v1.POST("/LastSeq", comm.LastSeq)
	v1.POST("/health", comm.GetHealth)
	v1.POST("/status", comm.GetStatus)

}

// Comm Comm
type Comm struct {
	SyncEs    *proto.ESDB
	ConvertEs *proto.ESDB
	ConvertID string
	Version   int32
	Config    *proto.ConfigNew
}

// LastSeq 获取当前最新同步以及解析的区块序列号
// @Summary 获取当前最新同步以及解析的区块序列号
// @Description get last sequence
// @Tags Comm
// @Produce  json
// @Param input body swagger.ClientRequestNil true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=rpcutils.RepLastSeq}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/LastSeq [post]
func (comm *Comm) LastSeq(c *gin.Context) {

	repLastSeq := rpcutils.RepLastSeq{}

	repLastSeq.LastConvertSeq = lastSeq(comm.ConvertEs.Host, comm.ConvertEs.Prefix, comm.ConvertID, comm.Version, comm.ConvertEs.User, comm.ConvertEs.Pwd)
	repLastSeq.LastSyncSeq = lastSeq(comm.SyncEs.Host, comm.SyncEs.Prefix, block.SyncSeq, comm.Version, comm.SyncEs.User, comm.SyncEs.Pwd)

	protocol.SetResult(c, repLastSeq, nil)

}

//lastSeq 获取已经同步或者解析的最新seq值
func lastSeq(host, prefix, id string, version int32, username, password string) int64 {
	log.Debug("lastSeq", "host", host, "prefix", prefix, "id", id)

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

// GetHealth 获取服务运行状态和版本
// @Summary 获取服务运行状态和版本
// @Description get server running state and version(这是一个说明)
// @Tags Comm
// @Produce  json
// @Param input body swagger.ClientRequestNil true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=swagger.Health}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/health [post]
func (comm *Comm) GetHealth(c *gin.Context) {
	protocol.SetResult(c, health.GetHealth(), nil)
}

// GetStatus 获取服务详细状态信息
// @Summary 获取服务详细状态信息
// @Description get server status in detail
// @Tags Comm
// @Produce  json
// @Param input body swagger.ClientRequestNil true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=swagger.Status}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/status [post]
func (comm *Comm) GetStatus(c *gin.Context) {
	protocol.SetResult(c, health.GetStatus(comm.Config), nil)
}
