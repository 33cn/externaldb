package multisig

import (
	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
	pty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/pkg/errors"
)

// status
const (
	StatusCreated = "created"
	StatusRevoked = "revoked"
	StatusDone    = "done"
)

var log = l.New("module", "db.multisig")

type msConvert struct {
	title  string
	symbol string

	// 上面是配置， 下面是对应每个交易处理的中间数据， 如何隔开
	env *db.TxEnv

	tx      *types.Transaction
	receipt *types.ReceiptData
	block   *db.Block

	accountIDBty   account.Account // for fee/bty
	accountIDAsset account.Account // for asset

	detail interface{}
}

// Support DB record keys
const (
	RAccountX        = "account"
	RTransactionX    = "transaction"
	RMultiSignatureX = "multi_signature"
)

func init() {
	converts.Register(pty.MultiSigX, NewConvert)
}

// NewConvert create
// TODO gen flags 暂时没有用上
func NewConvert(title, symbol string, supports []string) db.ExecConvert {
	e := &msConvert{title: title, symbol: symbol}

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
// TODO 定义mapping
func (t *msConvert) InitDB(cli db.DBCreator) error {
	err := util.InitIndex(cli, MSDBX, MSDBX, msMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, MSListDBX, MSListDBX, listMapping)
	if err != nil {
		return err
	}
	return util.InitIndex(cli, MSTxDBX, MSTxDBX, txMapping)
}

func (t *msConvert) InitConvert() {

}

func (t *msConvert) positionID(env *db.TxEnv) string {
	return util.PositionID("multisig", env.Block.Block.Height, env.TxIndex)
}

func (t *msConvert) setupEnv(env *db.TxEnv) {
	t.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	receipt := env.Block.Receipts[env.TxIndex]
	t.tx = tx
	t.receipt = receipt
	t.block = db.SetupBlock(env, tx.From(), common.ToHex(tx.Hash()))

	t.accountIDBty = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
	}
}

// ConveretTx impl
func (t *msConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	t.setupEnv(env)
	log.Info("convertTx", "position", t.positionID(env))

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}

	if t.receipt.Ty != types.ExecOk {
		records = append(records, &txRecord)
		return records, nil
	}

	var action pty.MultiSigAction
	err := types.Decode(t.tx.Payload, &action)
	if err != nil {
		return nil, errors.Wrapf(err, "decode tx action: %s", t.positionID(env))
	}

	switch action.Ty {
	case pty.ActionMultiSigAccCreate:
		log.Debug("Convert", "action", "ActionMultiSigAccCreate")
		return t.convertAccountCreate(action.GetMultiSigAccCreate(), op)
	case pty.ActionMultiSigOwnerOperate:
		log.Debug("Convert", "action", "ActionMultiSigOwnerOperate")
		return t.convertOwnerOperate(action.GetMultiSigOwnerOperate(), op)
	case pty.ActionMultiSigAccOperate:
		log.Debug("Convert", "action", "ActionMultiSigAccOperate")
		return t.convertAccountOperate(action.GetMultiSigAccOperate(), op)
	case pty.ActionMultiSigConfirmTx:
		log.Debug("Convert", "action", "ActionMultiSigConfirmTx")
		return t.convertConfirmTx(action.GetMultiSigConfirmTx(), op)
	case pty.ActionMultiSigExecTransferTo:
		log.Debug("Convert", "action", "ActionMultiSigExecTransferTo")
		return t.convertExecTransferTo(action.GetMultiSigExecTransferTo(), op)
	case pty.ActionMultiSigExecTransferFrom:
		log.Debug("Convert", "action", "ActionMultiSigExecTransferFrom")
		return t.convertExecTransferFrom(action.GetMultiSigExecTransferFrom(), op)
	}
	return nil, nil

}

// 1 order, 2 account
func (t *msConvert) convertAccountCreate(action *pty.MultiSigAccCreate, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	rs2, err := t.convertAccountCreateLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}

func (t *msConvert) convertOwnerOperate(action *pty.MultiSigOwnerOperate, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	t.detail = makeTxOwnerOp(action)

	rs2, err := t.convertOwnerOperateLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}

func makeTxOwnerOp(action *pty.MultiSigOwnerOperate) TxOwnerOperate {
	return TxOwnerOperate{
		Operate:   ownerOpStr(action.OperateFlag),
		OldOwner:  action.OldOwner,
		NewOwner:  action.NewOwner,
		NewWeight: action.NewWeight,
	}
}

func (t *msConvert) convertAccountOperate(action *pty.MultiSigAccOperate, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	t.detail = makeTxAccountOp(action)

	rs2, err := t.convertAccountOperateLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}

func makeTxAccountOp(action *pty.MultiSigAccOperate) TxAccountOperate {
	return TxAccountOperate{
		Operate:           accountOpStr(action.OperateFlag),
		DailyLimit:        makeSymbolLimit(action),
		NewRequiredWeight: action.NewRequiredWeight,
	}
}

func makeSymbolLimit(action *pty.MultiSigAccOperate) *SymbolLimit {
	if action.DailyLimit == nil {
		return nil
	}
	return &SymbolLimit{
		Symbol:     action.DailyLimit.Symbol,
		Execer:     action.DailyLimit.Execer,
		DailyLimit: action.DailyLimit.DailyLimit,
	}
}

func (t *msConvert) convertExecTransferTo(action *pty.MultiSigExecTransferTo, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	t.setupAsset(action.Execname, action.Symbol)

	rs2, err := t.convertExecTransferToLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}

func (t *msConvert) convertExecTransferFrom(action *pty.MultiSigExecTransferFrom, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	t.setupAsset(action.Execname, action.Symbol)
	t.detail = makeTransferOp(action)

	rs2, err := t.convertExecTransferFromLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}

func makeTransferOp(action *pty.MultiSigExecTransferFrom) TxTransferOperate {
	return TxTransferOperate{
		Symbol:   action.Symbol,
		Amount:   action.Amount,
		Note:     action.Note,
		Execname: action.Execname,
		To:       action.To,
		From:     action.From,
	}
}

func (t *msConvert) convertConfirmTx(action *pty.MultiSigConfirmTx, op int) ([]db.Record, error) {
	records := make([]db.Record, 0)

	tx := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	rs2, err := t.convertConfirmLogs(op)
	if err != nil {
		return records, err
	}
	records = append(records, rs2...)

	return records, nil
}
