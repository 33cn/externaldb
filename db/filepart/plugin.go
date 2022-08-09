package filepart

import (
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	fpdb "github.com/33cn/externaldb/db/filepart/db"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
)

// 本文件将做成插件需要的接口和函数都提到这个文件, 作为参考, 方便写文档, 也方便其他人实现新的插件
// 实现插件需要至少实现下面的内容
// 1. 注册插件
// 2. 注册插件需要的插件创建函数
// 3. InitDB 创建db
// 4. ConvertTx 展开数据
// 5. SetDB 插件需要访问数据库 (可选)

// 1注册插件, NewConvert函数需要满足接口 db.ExecConvert
func init() {
	converts.Register("user.filepart", NewConvert)
}

// NewConvert 2注册插件需要的插件创建函数
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &Convert{symbol: symbol, title: paraTitle}
	return e
}

// InitDB 创建db (插件满足接口 db.ExecConvert)
func (c *Convert) InitDB(cli db.DBCreator) error {
	return fpdb.InitESDB(cli)
}

// ConvertTx 展开数据 (插件满足接口 db.ExecConvert)
func (c *Convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	return c.convertTx(env, op)
}

// SetDB SetDB (非必要: 插件需要访问数据库时需要实现, 一般使用于虚拟插件数据展开在es上包含执行逻辑)
func (c *Convert) SetDB(db db.WrapDB) error {
	c.ConfigDB = proofconfig.NewConfigDB(db)
	return nil
}
