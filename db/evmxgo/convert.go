package evmxgo

import (
	"fmt"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
	pty "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

var log = l.New("module", "db.evmxgo")

// Convert tx convert
type Convert struct {
	symbol string
	title  string

	env     *db.TxEnv
	tx      *types.Transaction
	receipt *types.ReceiptData
}

func init() {
	converts.Register("evmxgo", NewConvert)
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

	err = util.InitIndex(cli, EvmxgoInfoDB, EvmxgoInfoDB, InfoRecordMapping)
	if err != nil {
		return err
	}

	return nil
}

func (e *Convert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "evmxgo", e.env.Block.Block.Height, e.env.TxIndex)
}

// ConvertTx impl
func (e *Convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	e.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	e.tx = tx
	e.receipt = env.Block.Receipts[env.TxIndex]

	evmxgoAction := pty.EvmxgoAction{}
	err := types.Decode(tx.GetPayload(), &evmxgoAction)
	if err != nil || e.receipt.Ty != types.ExecOk {
		//解析payload出错，打印错误日志。跳过解析交易详情并把交易的基本信息存到数据库
		if e.receipt.Ty == types.ExecOk {
			//交易执行成功但解析payload出错，说明代码存在bug
			log.Error(e.positionID(), "info", "decode payload error and skip convert tx", "error", err)
		}
		//交易执行失败说明用户构造payload不合法
		evmxgoTx := transaction.ConvertTransaction(e.env)
		txRecord := transaction.TxRecord{
			IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
			Op:   db.NewOp(op),
			Tx:   evmxgoTx,
		}
		return []db.Record{&txRecord}, nil
	}

	var records []db.Record
	switch evmxgoAction.Ty {
	case pty.ActionTransfer:
		log.Debug("Convert", "action", "transfer")
		records = e.convertTransfer(&evmxgoAction, op)
	case pty.EvmxgoActionTransferToExec:
		log.Debug("Convert", "action", "transfer to exec")
		records = e.convertTransferToExec(&evmxgoAction, op)
	case pty.ActionWithdraw:
		log.Debug("Convert", "action", "withdraw")
		records = e.convertWithdraw(&evmxgoAction, op)
	case pty.EvmxgoActionMint:
		log.Debug("Convert", "action", "mint")
		records = e.convertMint(&evmxgoAction, op)
	case pty.EvmxgoActionBurn:
		log.Debug("Convert", "action", "burn")
		records = e.convertBurn(&evmxgoAction, op)
	}

	return records, nil
}

func (e *Convert) convertTransfer(action *pty.EvmxgoAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.EvmxgoAction_Transfer).Transfer
	evmxgoTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Cointoken,
		To:     util.AddressConvert(v.To),
		Note:   string(v.Note),
	}
	evmxgoTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
		Op:   db.NewOp(op),
		Tx:   evmxgoTx,
	}
	records = append(records, &txRecord)

	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		if l.Ty == types.TyLogFee {
			acc.AssetSymbol = e.symbol
			acc.AssetExec = account.ExecCoinsX
		} else {
			acc.AssetSymbol = v.Cointoken
			acc.AssetExec = ExecEvmxgoX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertTransferToExec(action *pty.EvmxgoAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.EvmxgoAction_TransferToExec).TransferToExec
	evmxgoTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:   v.Cointoken,
		To:       util.AddressConvert(v.To),
		ExecName: v.ExecName,
		Note:     string(v.Note),
	}
	evmxgoTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
		Op:   db.NewOp(op),
		Tx:   evmxgoTx,
	}
	records = append(records, &txRecord)

	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		if l.Ty == types.TyLogFee {
			acc.AssetSymbol = e.symbol
			acc.AssetExec = account.ExecCoinsX
		} else {
			acc.AssetSymbol = v.Cointoken
			acc.AssetExec = ExecEvmxgoX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertWithdraw(action *pty.EvmxgoAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.EvmxgoAction_Withdraw).Withdraw
	evmxgoTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:   v.Cointoken,
		To:       util.AddressConvert(v.To),
		ExecName: v.ExecName,
		Note:     string(v.Note),
	}
	evmxgoTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
		Op:   db.NewOp(op),
		Tx:   evmxgoTx,
	}
	records = append(records, &txRecord)

	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		if l.Ty == types.TyLogFee {
			acc.AssetSymbol = e.symbol
			acc.AssetExec = account.ExecCoinsX
		} else {
			acc.AssetSymbol = v.Cointoken
			acc.AssetExec = ExecEvmxgoX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertMint(action *pty.EvmxgoAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.EvmxgoAction_Mint).Mint
	evmxgoTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:  v.Symbol,
		Amount:  v.Amount,
		Address: "",
	}
	evmxgoTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
		Op:   db.NewOp(op),
		Tx:   evmxgoTx,
	}
	records = append(records, &txRecord)

	symbol := action.GetMint().Symbol
	data := e.env.Block.Receipts[e.env.TxIndex]
	records = append(records, e.ParseMintAndBurnLog(data, symbol, op)...)
	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogEvmxgoMint {
			continue
		}
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		if l.Ty == types.TyLogFee {
			acc.AssetSymbol = e.symbol
			acc.AssetExec = account.ExecCoinsX
		} else {
			acc.AssetSymbol = symbol
			acc.AssetExec = ExecEvmxgoX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertBurn(action *pty.EvmxgoAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.EvmxgoAction_Burn).Burn
	evmxgoTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Symbol,
		Amount: v.Amount,
	}
	evmxgoTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(evmxgoTx.Hash),
		Op:   db.NewOp(op),
		Tx:   evmxgoTx,
	}
	records = append(records, &txRecord)

	symbol := action.GetBurn().Symbol
	data := e.env.Block.Receipts[e.env.TxIndex]
	records = append(records, e.ParseMintAndBurnLog(data, symbol, op)...)
	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogEvmxgoBurn {
			continue
		}
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		if l.Ty == types.TyLogFee {
			acc.AssetSymbol = e.symbol
			acc.AssetExec = account.ExecCoinsX
		} else {
			acc.AssetSymbol = symbol
			acc.AssetExec = ExecEvmxgoX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records

}

// ParseMintAndBurnLog ParseMintAndBurnLog
func (e *Convert) ParseMintAndBurnLog(data *types.ReceiptData, symbol string, op int) []db.Record {
	var records []db.Record
	for _, rlog := range data.Logs {
		switch rlog.Ty {
		case pty.TyLogEvmxgoMint, pty.TyLogEvmxgoBurn:
			receipt := pty.ReceiptEvmxgoAmount{}
			err := types.Decode(rlog.Log, &receipt)
			if err != nil {
				log.Info("LogEvmxgoBurnMint&Burn", "decode error", err)
				continue
			}
			var t *pty.Evmxgo
			if op == db.SeqTypeAdd {
				t = receipt.Current
			} else if op == db.SeqTypeDel {
				t = receipt.Prev
			}
			var info Evmxgo
			info.Amount = t.Total
			tokenRecord := Record{
				IKey:  db.NewIKey(EvmxgoInfoDB, EvmxgoInfoDB, t.Symbol),
				Op:    db.NewOp(db.OpAdd),
				value: info,
			}
			records = append(records, &tokenRecord)
		}

	}

	return records
}
