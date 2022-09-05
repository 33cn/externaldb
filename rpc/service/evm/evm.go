package evm

import (
	"github.com/33cn/externaldb/db/evm/nft/db"
	"github.com/33cn/externaldb/rpc/middleware/protocol"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/gin-gonic/gin"
)

// EVM EVM合约
type EVM struct {
	*rpcutils.DBRead
}

// InitRouter 初始化proofrpc接口的router路由表
func InitRouter(router gin.IRouter, dbread *rpcutils.DBRead) {
	evm := EVM{DBRead: dbread}

	v1 := router.Group("/v1")
	v1.POST("/evm/List", evm.ListEVM)
	v1.POST("/evm/Count", evm.CountEVM)
	v1.POST("/evm/nft/account/List", evm.ListNTFAccount)
	v1.POST("/evm/nft/account/Count", evm.CountNTFAccount)
	v1.POST("/evm/nft/transfer/List", evm.ListNTFTransfer)
	v1.POST("/evm/nft/transfer/Count", evm.CountNTFTransfer)
}

// ListEVM 查询通证列表
// @Summary 查询通证列表
// @Description list evm token
// @Tags EVM
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListEVMResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/List [post]
func (t *EVM) ListEVM(c *gin.Context) {
	protocol.List(c, db.TokenX, db.TokenX, rpcutils.DecodeJSONToMap, t.DBRead)
}

// CountEVM 查询通证数量
// @Summary 查询通证数量
// @Description get evm count of organization/sender
// @Tags EVM
// @Produce  json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/Count [post]
func (t *EVM) CountEVM(c *gin.Context) {
	protocol.Count(c, db.TokenX, db.TokenX, t.DBRead)
}

// ListNTFAccount 查询账户信息列表
// @Summary 查询账户信息列表
// @Description list evm account
// @Tags EVM
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListEVMResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/nft/account/List [post]
func (t *EVM) ListNTFAccount(c *gin.Context) {
	protocol.List(c, db.AccountX, db.AccountX, rpcutils.DecodeJSONToMap, t.DBRead)
}

// CountNTFAccount 查询账户信息数量
// @Summary 查询账户信息数量
// @Description get account count
// @Tags EVM
// @Produce  json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/nft/account/Count [post]
func (t *EVM) CountNTFAccount(c *gin.Context) {
	protocol.Count(c, db.AccountX, db.AccountX, t.DBRead)
}

// ListNTFTransfer 查询转账列表
// @Summary 查询转账列表
// @Description list evm transfer
// @Tags EVM
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListEVMResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/nft/transfer/List [post]
func (t *EVM) ListNTFTransfer(c *gin.Context) {
	protocol.List(c, db.TransferX, db.TransferX, rpcutils.DecodeJSONToMap, t.DBRead)
}

// CountNTFTransfer 查询转账数量
// @Summary 查询转账数量
// @Description get transfer count
// @Tags EVM
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/evm/nft/transfer/Count [post]
func (t *EVM) CountNTFTransfer(c *gin.Context) {
	protocol.Count(c, db.TransferX, db.TransferX, t.DBRead)
}
