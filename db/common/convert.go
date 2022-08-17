package common

import (
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

// consts
const (
	NameX         = "common"
	RAccountX     = "account"
	RTransactionX = "transaction"
)

var log = l.New("module", "db.common")

type convert struct {
	title  string
	symbol string

	// 上面是配置， 下面是对应每个交易处理的中间数据， 如何隔开
}

func init() {
	converts.Register(NameX, NewConvert)
}

// NewConvert create
func NewConvert(title, symbol string, supports []string) db.ExecConvert {
	e := &convert{title: title, symbol: symbol}

	return e
}

// 接口划分
// Init & Convert & Save
// Init 大写：需要配置
// Init 两部分： db & convert
// InitConvert & Convert -> tx convert
// InitDB & Save -> tx save, db part 可以配置不同的存储
// 数据项配置，是否展开， 可以在convert 阶段(x)， 也可以在save阶段丢弃
// InitConvert 在 newConvert 可以做掉

// InitDB 下阶段再处理配置存储(ES/MySQL/...)的问题
func (t *convert) InitDB(cli db.DBCreator) error {
	err := account.InitDB(cli)
	if err != nil {
		return err
	}

	return transaction.InitDB(cli)
}

// ConveretTx impl
func (t *convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	log.Info("convertTx", "position", util.PositionID(NameX, env.Block.Block.Height, env.TxIndex))

	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	return records, nil
}
