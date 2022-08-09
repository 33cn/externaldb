package exchange

import (
	"fmt"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

var log = l.New("module", "db.exchange")

// Convert tx convert
type Convert struct {
	symbol string
	title  string

	env     *db.TxEnv
	tx      *types.Transaction
	receipt *types.ReceiptData

	// convertTx *transaction.Transaction
}

func init() {
	converts.Register("exchange", NewConvert)
}

// NewConvert NewConvert
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &Convert{symbol: symbol, title: paraTitle}

	return e
}

// InitDB init db
func (e *Convert) InitDB(cli db.DBCreator) error {
	var err error

	err = account.InitDB(cli)
	if err != nil {
		return err
	}

	err = transaction.InitDB(cli)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, ExchangeInfoDB, ExchangeInfoDB, InfoRecordMapping)
	if err != nil {
		return err
	}

	return nil
}

func (e *Convert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "exchange", e.env.Block.Block.Height, e.env.TxIndex)
}

func (e *Convert) setEnv(env *db.TxEnv) {
	e.env = env
	e.tx = env.Block.Block.Txs[env.TxIndex]
	e.receipt = env.Block.Receipts[env.TxIndex]
}

func (e *Convert) txError(op int, err error) ([]db.Record, error) {
	//解析payload出错，打印错误日志。跳过解析交易详情并把交易的基本信息存到数据库
	if e.receipt.Ty == types.ExecOk {
		//交易执行成功但解析payload出错，说明代码存在bug
		log.Error(e.positionID(), "info", "decode payload error and skip convert tx", "error", err)
	}
	//交易执行失败说明用户构造payload不合法
	exchangeTx := transaction.ConvertTransaction(e.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(exchangeTx.Hash),
		Op:   db.NewOp(op),
		Tx:   exchangeTx,
	}
	return []db.Record{&txRecord}, nil
}

// ConvertTx impl
func (e *Convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	e.setEnv(env)

	action := pty.ExchangeAction{}
	err := types.Decode(e.tx.GetPayload(), &action)
	if err != nil || e.receipt.Ty != types.ExecOk {
		return e.txError(op, err)
	}

	switch action.Ty {
	case pty.TyLimitOrderAction:
		log.Debug("Convert", "action", "LimitOrder")
		return e.convertLimitOrder(action.GetLimitOrder(), op)
	case pty.TyMarketOrderAction:
		log.Debug("Convert", "action", "MarketOrder")
		return e.convertMarketOrder(action.GetMarketOrder(), op)
	case pty.TyRevokeOrderAction:
		log.Debug("Convert", "action", "RevokeOrder")
		return e.convertRevokeOrder(action.GetRevokeOrder(), op)
	default:
		log.Error("Covert", "action", "type mismatch")
	}

	return nil, nil
}

func (e *Convert) convertLimitOrder(payload *pty.LimitOrder, op int) ([]db.Record, error) {
	var records []db.Record
	exchangeTx := transaction.ConvertTransaction(e.env)
	exchangeTx.ActionName = pty.NameLimitOrderAction
	exchangeTx.Options = payload
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(exchangeTx.Hash),
		Op:   db.NewOp(op),
		Tx:   exchangeTx,
	}
	records = append(records, &txRecord)
	return records, nil
}

func (e *Convert) convertMarketOrder(payload *pty.MarketOrder, op int) ([]db.Record, error) {
	var records []db.Record
	exchangeTx := transaction.ConvertTransaction(e.env)
	exchangeTx.ActionName = pty.NameMarketOrderAction
	exchangeTx.Options = payload
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(exchangeTx.Hash),
		Op:   db.NewOp(op),
		Tx:   exchangeTx,
	}
	records = append(records, &txRecord)
	return records, nil
}

func (e *Convert) convertRevokeOrder(payload *pty.RevokeOrder, op int) ([]db.Record, error) {
	var records []db.Record
	exchangeTx := transaction.ConvertTransaction(e.env)
	exchangeTx.ActionName = pty.NameRevokeOrderAction
	exchangeTx.Options = payload
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(exchangeTx.Hash),
		Op:   db.NewOp(op),
		Tx:   exchangeTx,
	}
	records = append(records, &txRecord)
	return records, nil
}
