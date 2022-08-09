package token

import (
	"fmt"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

var log = l.New("module", "db.token")

// Convert tx convert
type Convert struct {
	symbol string
	title  string

	env     *db.TxEnv
	tx      *types.Transaction
	receipt *types.ReceiptData
}

func init() {
	converts.Register("token", NewConvert)
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

	err = util.InitIndex(cli, TokenInfoDB, TokenInfoDB, InfoRecordMapping)
	if err != nil {
		return err
	}

	return nil
}

func (e *Convert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "trade", e.env.Block.Block.Height, e.env.TxIndex)
}

// ConvertTx impl
func (e *Convert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	e.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	e.tx = tx
	e.receipt = env.Block.Receipts[env.TxIndex]

	tokenAction := pty.TokenAction{}
	err := types.Decode(tx.GetPayload(), &tokenAction)
	if err != nil || e.receipt.Ty != types.ExecOk {
		//解析payload出错，打印错误日志。跳过解析交易详情并把交易的基本信息存到数据库
		if e.receipt.Ty == types.ExecOk {
			//交易执行成功但解析payload出错，说明代码存在bug
			log.Error(e.positionID(), "info", "decode payload error and skip convert tx", "error", err)
		}
		//交易执行失败说明用户构造payload不合法
		tokenTx := transaction.ConvertTransaction(e.env)
		txRecord := transaction.TxRecord{
			IKey: transaction.NewTransactionKey(tokenTx.Hash),
			Op:   db.NewOp(op),
			Tx:   tokenTx,
		}
		return []db.Record{&txRecord}, nil
	}

	var records []db.Record
	switch tokenAction.Ty {
	case pty.TokenActionPreCreate:
		log.Debug("Convert", "action", "pre create")
		records = e.convertPreCreate(&tokenAction, op)
	case pty.TokenActionRevokeCreate:
		log.Debug("Convert", "action", "revoke create")
		records = e.convertRevokeCreate(&tokenAction, op)
	case pty.TokenActionFinishCreate:
		log.Debug("Convert", "action", "finish create")
		records = e.convertFinishCreate(&tokenAction, op)
	case pty.ActionTransfer:
		log.Debug("Convert", "action", "transfer")
		records = e.convertTransfer(&tokenAction, op)
	case pty.TokenActionTransferToExec:
		log.Debug("Convert", "action", "transfer to exec")
		records = e.convertTransferToExec(&tokenAction, op)
	case pty.ActionWithdraw:
		log.Debug("Convert", "action", "withdraw")
		records = e.convertWithdraw(&tokenAction, op)
	case pty.TokenActionMint:
		log.Debug("Convert", "action", "mint")
		records = e.convertMint(&tokenAction, op)
	case pty.TokenActionBurn:
		log.Debug("Convert", "action", "burn")
		records = e.convertBurn(&tokenAction, op)
	}

	return records, nil
}

func (e *Convert) convertPreCreate(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.TokenAction_TokenPreCreate).TokenPreCreate
	info := Token{
		Name:          v.Name,
		Symbol:        v.Symbol,
		Amount:        v.Total,
		Owner:         v.Owner,
		Creator:       e.env.Block.Block.Txs[e.env.TxIndex].From(),
		Introduction:  v.Introduction,
		Price:         v.Price,
		Category:      int64(v.Category),
		Status:        pty.TokenStatusPreCreated,
		PrepareHeight: e.env.Block.Block.Height,
	}
	infoRecord := RecordToken{
		IKey:  db.NewIKey(TokenInfoDB, TokenInfoDB, v.Symbol),
		Op:    db.NewOp(op),
		value: &info,
	}
	records = append(records, &infoRecord)

	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:       v.Symbol,
		Name:         v.Name,
		Introduction: v.Introduction,
		Total:        v.Total,
		Price:        v.Price,
		Owner:        v.Owner,
		Category:     int64(v.Category),
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
	}
	records = append(records, &txRecord)

	acc := account.Account{
		AssetSymbol: e.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogPreCreateToken {
			continue
		}
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records

}

func (e *Convert) convertFinishCreate(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.TokenAction_TokenFinishCreate).TokenFinishCreate
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Symbol,
		Owner:  v.Owner,
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
	}
	records = append(records, &txRecord)

	var info Token
	if op == db.SeqTypeAdd {
		info = Token{
			Status:       pty.TokenStatusCreated,
			CreateHeight: e.env.Block.Block.Height,
		}
	} else if op == db.SeqTypeDel {
		info = Token{
			Status:       pty.TokenStatusPreCreated,
			CreateHeight: -1,
		}
	}

	infoRecord := RecordToken{
		IKey:  db.NewIKey(TokenInfoDB, TokenInfoDB, v.Symbol),
		Op:    db.NewOp(db.OpUpdate),
		value: &info,
	}
	records = append(records, &infoRecord)

	acc := account.Account{
		AssetSymbol: e.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogFinishCreateToken {
			continue
		}
		if l.Ty == types.TyLogGenesisTransfer {
			acc.AssetSymbol = v.Symbol
			acc.AssetExec = ExecTokenX
		}
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertRevokeCreate(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.TokenAction_TokenRevokeCreate).TokenRevokeCreate
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Symbol,
		Owner:  v.Owner,
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
	}
	records = append(records, &txRecord)

	var info Token
	if op == db.SeqTypeAdd {
		info = Token{
			Status:       pty.TokenStatusCreateRevoked,
			RevokeHeight: e.env.Block.Block.Height,
		}
	} else if op == db.SeqTypeDel {
		info = Token{
			Status:       pty.TokenStatusPreCreated,
			RevokeHeight: -1,
		}
	}

	infoRecord := RecordToken{
		IKey:  db.NewIKey(TokenInfoDB, TokenInfoDB, v.Symbol),
		Op:    db.NewOp(db.OpUpdate),
		value: &info,
	}
	records = append(records, &infoRecord)

	acc := account.Account{
		AssetSymbol: e.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogRevokeCreateToken {
			continue
		}
		accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
		if err2 != nil {
			log.Info("convertTx", "logType", l.Ty, "err", err2)
			continue
		}
		acc.Detall = accountDetail
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertTransfer(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.TokenAction_Transfer).Transfer
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Cointoken,
		To:     v.To,
		Note:   string(v.Note),
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
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
			acc.AssetExec = ExecTokenX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertTransferToExec(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.TokenAction_TransferToExec).TransferToExec
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:   v.Cointoken,
		To:       v.To,
		ExecName: v.ExecName,
		Note:     string(v.Note),
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
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
			acc.AssetExec = ExecTokenX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertWithdraw(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record

	v := action.Value.(*pty.TokenAction_Withdraw).Withdraw
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol:   v.Cointoken,
		To:       v.To,
		ExecName: v.ExecName,
		Note:     string(v.Note),
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
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
			acc.AssetExec = ExecTokenX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertMint(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.TokenAction_TokenMint).TokenMint
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Symbol,
		Amount: v.Amount,
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
	}
	records = append(records, &txRecord)

	symbol := action.GetTokenMint().Symbol
	data := e.env.Block.Receipts[e.env.TxIndex]
	records = append(records, e.ParseMintAndBurnLog(data, symbol, op)...)
	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogTokenMint {
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
			acc.AssetExec = ExecTokenX
		}
		record := &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
		records = append(records, record)

		rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
		records = append(records, rrecord)
	}

	return records
}

func (e *Convert) convertBurn(action *pty.TokenAction, op int) []db.Record {
	var records []db.Record
	v := action.Value.(*pty.TokenAction_TokenBurn).TokenBurn
	tokenTx := transaction.ConvertTransaction(e.env)
	options := TxOption{
		Symbol: v.Symbol,
		Amount: v.Amount,
	}
	tokenTx.Options = &options
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tokenTx.Hash),
		Op:   db.NewOp(op),
		Tx:   tokenTx,
	}
	records = append(records, &txRecord)

	symbol := action.GetTokenBurn().Symbol
	data := e.env.Block.Receipts[e.env.TxIndex]
	records = append(records, e.ParseMintAndBurnLog(data, symbol, op)...)
	acc := account.Account{
		HeightIndex: db.HeightIndex(e.env.Block.Block.Height, e.env.TxIndex),
		Height:      e.env.Block.Block.Height,
		BlockTime:   e.env.Block.Block.BlockTime,
	}
	for _, l := range e.env.Block.Receipts[e.env.TxIndex].Logs {
		if l.Ty == pty.TyLogTokenBurn {
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
			acc.AssetExec = ExecTokenX
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
		case pty.TyLogTokenMint, pty.TyLogTokenBurn:
			receipt := pty.ReceiptTokenAmount{}
			err := types.Decode(rlog.Log, &receipt)
			if err != nil {
				log.Info("LogTokenBurnMint&Burn", "decode error", err)
				continue
			}
			var t *pty.Token
			if op == db.SeqTypeAdd {
				t = receipt.Current
			} else if op == db.SeqTypeDel {
				t = receipt.Prev
			}
			var info Token
			info.Amount = t.Total
			tokenRecord := Record{
				IKey:  db.NewIKey(TokenInfoDB, TokenInfoDB, t.Symbol),
				Op:    db.NewOp(db.OpUpdate),
				value: info,
			}
			records = append(records, &tokenRecord)
		}

	}

	return records
}
