package coins

import (
	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	pty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

type coinsConvert struct {
	symbol string
	title  string

	env *db.TxEnv

	tx         *types.Transaction
	receipt    *types.ReceiptData
	block      *db.Block
	accountBty account.Account
}

var log = l.New("module", "db.coins")

func init() {
	converts.Register("coins", NewConvert)
}

// NewConvert NewConvert
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &coinsConvert{symbol: symbol, title: paraTitle}

	return e
}

// InitDB init db
func (t *coinsConvert) InitDB(cli db.DBCreator) error {
	var err error

	err = account.InitDB(cli)
	if err != nil {
		return err
	}

	err = transaction.InitDB(cli)
	if err != nil {
		return err
	}

	return nil
}

// Convert Convert
func (t *coinsConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	log.Info("convertTX", "position", t.positionID(env))

	t.env = env
	receipt := env.Block.Receipts[env.TxIndex]
	t.tx = env.Block.Block.Txs[env.TxIndex]
	t.receipt = receipt
	t.block = db.SetupBlock(env, t.tx.From(), common.ToHex(t.tx.Hash()))

	var records []db.Record

	tx := transaction.ConvertTransaction(env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	if t.receipt.Ty != types.ExecOk {
		records = append(records, &txRecord)
		return records, nil
	}

	var action pty.CoinsAction
	err := types.Decode(t.tx.Payload, &action)
	if err != nil {
		return nil, err
	}

	// tx option setup here, if needed
	records = append(records, &txRecord)

	t.accountBty = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}

	for _, l := range receipt.Logs {
		rs, err2 := account.RecordHelper(l, op, t.accountBty)
		if err2 != nil {
			log.Info("convertTX", "position", t.positionID(env), "logType", l.Ty, "err", err2)
			continue
		}
		records = append(records, rs...)
	}
	return records, nil
}

func (t *coinsConvert) positionID(env *db.TxEnv) string {
	return util.PositionID("coins", env.Block.Block.Height, env.TxIndex)
}
