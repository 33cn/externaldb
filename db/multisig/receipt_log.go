package multisig

// 只能通过日志顺序判断 帐号receipt对应的资产类型

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/util"
	pty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

// create: logs := []int{types.TyLogFee, TyLogMultiSigAccCreate}
// owner: logs := []int{types.TyLogFee, 0~1/4, TyLogTxCountUpdate, TyLogMultiSigTx}
//	4={TyLogMultiSigOwnerAdd/TyLogMultiSigOwnerDel TyLogMultiSigOwnerModify/TyLogMultiSigOwnerReplace}
//                     ReceiptOwnerAddOrDel                      ReceiptOwnerModOrRep
// acc op: logs := []int{types.TyLogFee, 0~1/2, TyLogTxCountUpdate, TyLogMultiSigTx}
//  2={TyLogMultiSigAccWeightModify/TyLogMultiSigAccDailyLimitModify}
// transferFrom: fee,  0~1/1 ExecTransferFrozen, TyLogTxCountUpdate, 0~1/1 TyLogDailyLimitUpdate,TyLogMultiSigTx
// ConfirmTx: 1/3  owner/acc op/transferFrom 没有 TyLogTxCountUpdate
// transferTo: Fee ExecTransfer, ExecFrozen
func (t *msConvert) convertAccountCreateLogs(op int) ([]db.Record, error) {
	//logs := []int{types.TyLogFee, TyLogMultiSigAccCreate}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertAccountCreateLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "AccountCreateLogs e:1~2.a:%d", cnt)
	}

	rs, err := t.TyLogMultiSigAccCreate(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "AccountCreateLogs log[-1]")
	}
	records = append(records, rs...)

	if cnt == 2 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "AccountCreateLogs: log-fee")
		}

		records = append(records, fee...)
	}
	return records, nil
}

func (t *msConvert) conventLogs(log *types.ReceiptLog, op int) ([]db.Record, error) {
	switch log.Ty {
	case pty.TyLogMultiSigAccCreate:
		return t.TyLogMultiSigAccCreate(log.Log, op)
	// owner log
	case pty.TyLogMultiSigOwnerAdd:
		return t.TyLogMultiSigOwnerAdd(log.Log, op)
	case pty.TyLogMultiSigOwnerDel:
		return t.TyLogMultiSigOwnerDel(log.Log, op)
	case pty.TyLogMultiSigOwnerModify:
		return t.TyLogMultiSigOwnerModify(log.Log, op)
	case pty.TyLogMultiSigOwnerReplace:
		return t.TyLogMultiSigOwnerReplace(log.Log, op)
	// limit log
	case pty.TyLogMultiSigAccWeightModify:
		return t.TyLogMultiSigAccWeightModify(log.Log, op)
	case pty.TyLogMultiSigAccDailyLimitAdd:
		return t.TyLogMultiSigAccDailyLimitAdd(log.Log, op)
	case pty.TyLogMultiSigAccDailyLimitModify:
		return t.TyLogMultiSigAccDailyLimitModify(log.Log, op)
	case pty.TyLogDailyLimitUpdate:
		return t.TyLogDailyLimitUpdate(log.Log, op)
	// confirm
	case pty.TyLogMultiSigConfirmTx:
		return t.TyLogMultiSigConfirmTx(log.Log, op)
	case pty.TyLogMultiSigConfirmTxRevoke:
		return t.TyLogMultiSigConfirmTxRevoke(log.Log, op)
	// tx
	case pty.TyLogMultiSigTx:
		return t.TyLogMultiSigTx(log.Log, op)
	case pty.TyLogTxCountUpdate:
		return t.TyLogTxCountUpdate(log.Log, op)
	case types.TyLogFee:
		records, err := account.RecordHelper(log, op, t.accountIDBty)
		return records, err
	default:
		records, err := account.RecordHelper(log, op, t.accountIDAsset)
		return records, err
	}
}

//
// owner: logs := []int{types.TyLogFee, 0~1/4, 0~1 TyLogTxCountUpdate, TyLogMultiSigTx}
//	4={TyLogMultiSigOwnerAdd/TyLogMultiSigOwnerDel TyLogMultiSigOwnerModify/TyLogMultiSigOwnerReplace}
//                     ReceiptOwnerAddOrDel                      ReceiptOwnerModOrRep
// 0~1/4: 0 处于提交状态， 1 已经执行了
// 0~1: 第一次有
func (t *msConvert) convertOwnerOperateLogs(op int) ([]db.Record, error) {
	var records []db.Record

	cnt := len(t.receipt.Logs)
	log.Debug("convertOwnerOperateLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "OwnerOperateLog e:1~4.a:%d", cnt)
	}

	for i, log := range t.receipt.Logs {
		rs, err := t.conventLogs(log, op)
		if err != nil {
			return nil, errors.Wrapf(err, "OwnerOperateLog %d/%d", i, cnt)
		}
		records = append(records, rs...)
	}
	return records, nil
}

// transferTo: Fee ExecTransfer, ExecFrozen
func (t *msConvert) convertExecTransferToLogs(op int) ([]db.Record, error) {
	var records []db.Record

	cnt := len(t.receipt.Logs)
	log.Debug("convertExecTransferToLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "ExecTransferToLogs e:1~4.a:%d", cnt)
	}

	for i, log := range t.receipt.Logs {
		rs, err := t.conventLogs(log, op)
		if err != nil {
			return nil, errors.Wrapf(err, "ExecTransferToLogs %d/%d", i, cnt)
		}
		records = append(records, rs...)
	}
	return records, nil
}

// acc op: logs := []int{types.TyLogFee, 0~1/2, TyLogTxCountUpdate, TyLogMultiSigTx}
//  2={TyLogMultiSigAccWeightModify/TyLogMultiSigAccDailyLimitModify}
func (t *msConvert) convertAccountOperateLogs(op int) ([]db.Record, error) {
	var records []db.Record

	cnt := len(t.receipt.Logs)
	log.Debug("convertAccountOperateLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "AccountOperateLogs e:1~4.a:%d", cnt)
	}

	for i, log := range t.receipt.Logs {
		rs, err := t.conventLogs(log, op)
		if err != nil {
			return nil, errors.Wrapf(err, "AccountOperateLogs %d/%d", i, cnt)
		}
		records = append(records, rs...)
	}
	return records, nil
}

// transferFrom: fee,  0~1/1 ExecTransferFrozen, 0~1 TyLogTxCountUpdate, 0~1/1 TyLogDailyLimitUpdate,TyLogMultiSigTx
func (t *msConvert) convertExecTransferFromLogs(op int) ([]db.Record, error) {
	var records []db.Record

	// limit update
	cnt := len(t.receipt.Logs)
	log.Debug("convertExecTransferFromLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "ExecTransferFromLogs e:1~4.a:%d", cnt)
	}

	for i, log := range t.receipt.Logs {
		rs, err := t.conventLogs(log, op)
		if err != nil {
			return nil, errors.Wrapf(err, "ExecTransferFromLogs %d/%d", i, cnt)
		}
		records = append(records, rs...)
	}
	return records, nil
}

func (t *msConvert) findSetAsset(logs *types.ReceiptData, op int) error {
	for _, log := range logs.Logs {
		if log.Ty == pty.TyLogDailyLimitUpdate {
			records, err := t.TyLogDailyLimitUpdate(log.Log, op)
			if err != nil {
				return err
			}
			if len(records) != 1 {
				return errors.Wrapf(errors.New("Not Found"), "find asset from TyLogDailyLimitUpdate failed")
			}
			limit, ok := records[0].(*MSUpdateLimitRecord)
			if !ok {
				return errors.Wrapf(errors.New("Not Found"), "get asset from TyLogDailyLimitUpdate failed")
			}
			exec, symbol := limit.u.Execer, limit.u.Symbol
			t.setupAsset(exec, symbol)
		}
	}
	return nil
}

func (t *msConvert) convertConfirmLogs(op int) ([]db.Record, error) {
	var records []db.Record

	cnt := len(t.receipt.Logs)
	log.Debug("convertConfirmLogs", "log_cnt", cnt)
	if cnt < 1 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "convertConfirmLogs e:1~4.a:%d", cnt)
	}

	// confirm 分三类，在转账分类需要先找出资产类型
	err := t.findSetAsset(t.receipt, op)
	if err != nil {
		return nil, errors.Wrap(err, "convertConfirmLogs findSetAsset")
	}

	for i, log := range t.receipt.Logs {
		rs, err := t.conventLogs(log, op)
		if err != nil {
			return nil, errors.Wrapf(err, "convertConfirmLogs %d/%d", i, cnt)
		}
		records = append(records, rs...)
	}
	return records, nil
}

func newTxKey(id string) *db.IKey {
	return db.NewIKey(MSTxDBX, MSTxDBX, id)
}

func newListKey(id string) *db.IKey {
	return db.NewIKey(MSListDBX, MSListDBX, id)
}

func newKey(id string) *db.IKey {
	return db.NewIKey(MSDBX, MSDBX, id)
}

func makeLimit(l *pty.DailyLimit, msAddress string) *MSLimit {
	return &MSLimit{
		MultiSigAddr: msAddress,
		Symbol:       l.Symbol,
		Execer:       l.Execer,
		DailyLimit:   l.DailyLimit,
		SpentToday:   l.SpentToday,
		LastDay:      l.LastDay,
		Type:         MSTypeLimitX,
	}
}

// TyLogMultiSigAccCreate 只输出多重签名的账户地址
func (t *msConvert) TyLogMultiSigAccCreate(v []byte, op int) ([]db.Record, error) {
	var l pty.MultiSig
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigAccCreate log")
	}

	var records []db.Record
	ms := &MS{
		CreateAddr:     util.AddressConvert(t.tx.From()),
		MultiSigAddr:   l.MultiSigAddr,
		TxCount:        l.TxCount,
		RequiredWeight: l.RequiredWeight,
		Type:           MSTypeAccountX,
	}
	t2 := &MSRecord{
		IKey: newKey(ms.ID()),
		Op:   db.NewOp(op),
		m:    ms,
	}
	records = append(records, t2)

	for _, o := range l.Owners {
		msOwner := t.makeOwner(o, l.MultiSigAddr)
		r := &MSUpdateOwnerRecord{
			IKey: newKey(msOwner.ID()),
			Op:   db.NewOp(op),
			u:    msOwner,
		}
		records = append(records, r)
	}

	for _, limit := range l.DailyLimits {
		msLimit := makeLimit(limit, l.MultiSigAddr)
		r := &MSUpdateLimitRecord{
			IKey: newKey(msLimit.ID()),
			Op:   db.NewOp(op),
			u:    msLimit,
		}
		records = append(records, r)
	}

	return records, nil
}

//输出add的owner：addr和weight
func (t *msConvert) TyLogMultiSigOwnerAdd(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptOwnerAddOrDel
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigOwnerAdd log")
	}

	current := t.makeOwner(l.Owner, l.MultiSigAddr)

	t2 := &MSUpdateOwnerRecord{
		IKey: newKey(current.ID()),
		Op:   db.NewOp(op),
		u:    current,
	}

	return []db.Record{t2}, nil
}

//输出del的owner：addr和weight
func (t *msConvert) TyLogMultiSigOwnerDel(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptOwnerAddOrDel
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigOwnerDel log")
	}

	Op := db.OpDel
	if op == db.SeqTypeDel {
		Op = db.OpAdd
	}

	current := t.makeOwner(l.Owner, l.MultiSigAddr)
	t2 := &MSUpdateOwnerRecord{
		IKey: newKey(current.ID()),
		Op:   db.NewOp(Op),
		u:    current,
	}

	return []db.Record{t2}, nil
}

//输出modify的owner：preweight以及currentweight
func (t *msConvert) TyLogMultiSigOwnerModify(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptOwnerModOrRep
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigOwnerModify log")
	}

	current := t.makeOwner(l.CurrentOwner, l.MultiSigAddr)
	prev := t.makeOwner(l.PrevOwner, l.MultiSigAddr)
	if op == db.SeqTypeDel {
		prev, current = current, prev
	}

	var records []db.Record
	t2 := &MSUpdateOwnerRecord{
		IKey: newKey(current.ID()),
		Op:   db.NewOp(db.OpAdd),
		u:    current,
	}
	records = append(records, t2)
	if prev.Address != current.Address {
		r := &MSUpdateOwnerRecord{
			IKey: newKey(current.ID()),
			Op:   db.NewOp(db.OpDel),
			u:    current,
		}
		records = append(records, r)
	}

	return records, nil
}

func (t *msConvert) makeOwner(owner *pty.Owner, msAddress string) *MSOwner {
	return &MSOwner{
		MultiSigAddr: msAddress,
		Address:      owner.OwnerAddr,
		Weight:       owner.Weight,
		Type:         MSTypeOwnerX,
	}
}

//输出old的owner的信息：以及当前的owner信息：addr+weight
func (t *msConvert) TyLogMultiSigOwnerReplace(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptOwnerModOrRep
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigOwnerReplace log")
	}

	current := t.makeOwner(l.CurrentOwner, l.MultiSigAddr)
	prev := t.makeOwner(l.PrevOwner, l.MultiSigAddr)
	if op == db.SeqTypeDel {
		prev, current = current, prev
	}

	var records []db.Record
	t2 := &MSUpdateOwnerRecord{
		IKey: newKey(current.ID()),
		Op:   db.NewOp(db.OpAdd),
		u:    current,
	}
	records = append(records, t2)
	if prev.Address != current.Address {
		r := &MSUpdateOwnerRecord{
			IKey: newKey(current.ID()),
			Op:   db.NewOp(db.OpDel),
			u:    current,
		}
		records = append(records, r)
	}

	return records, nil
}

//输出修改前后确认权重的值：preReqWeight和curReqWeight
func (t *msConvert) TyLogMultiSigAccWeightModify(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptWeightModify
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigOwnerReplace log")
	}

	weight := l.CurrentWeight
	if op == db.SeqTypeDel {
		weight = l.PrevWeight
	}

	// update
	u := MSUpdateWeight{
		Address:        l.MultiSigAddr,
		RequiredWeight: weight,
	}
	t2 := &MSUpdateWeightRecord{
		IKey: newKey(msID(l.MultiSigAddr)),
		Op:   db.NewOp(db.OpUpdate),
		u:    &u,
	}

	return []db.Record{t2}, nil
}

//输出add的DailyLimit：Symbol和DailyLimit
func (t *msConvert) TyLogMultiSigAccDailyLimitAdd(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptDailyLimitOperate
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigAccDailyLimitAdd log")
	}

	limit := makeLimit(l.CurDailyLimit, l.MultiSigAddr)
	t2 := &MSUpdateLimitRecord{
		IKey: newKey(limit.ID()),
		Op:   db.NewOp(op),
		u:    limit,
	}

	return []db.Record{t2}, nil
}

//输出modify的DailyLimit：preDailyLimit以及currentDailyLimit
func (t *msConvert) TyLogMultiSigAccDailyLimitModify(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptDailyLimitOperate
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigAccDailyLimitModify log")
	}

	limit := makeLimit(l.CurDailyLimit, l.MultiSigAddr)
	if op == db.SeqTypeDel {
		limit = makeLimit(l.PrevDailyLimit, l.MultiSigAddr)
	}

	t2 := &MSUpdateLimitRecord{
		IKey: newKey(limit.ID()),
		Op:   db.NewOp(db.OpUpdate),
		u:    limit,
	}

	return []db.Record{t2}, nil
}

//对某笔未执行交易的确认
func (t *msConvert) TyLogMultiSigConfirmTx(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptConfirmTx
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigConfirmTx log")
	}

	sigList := t.makeSigList(l.MultiSigTxOwner, false, false, common.ToHex(t.tx.Hash()))
	t2 := &SigListRecord{
		IKey: newListKey(sigList.ID()),
		Op:   db.NewOp(op),
		m:    sigList,
	}

	return []db.Record{t2}, nil
}

// 已经确认交易的撤销只针对还未执行的交易
func (t *msConvert) TyLogMultiSigConfirmTxRevoke(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptConfirmTx
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigConfirmTxRevoke log")
	}

	Op := db.OpDel
	if op == db.SeqTypeDel {
		Op = db.OpAdd
	}

	sigList := t.makeSigList(l.MultiSigTxOwner, false, false, common.ToHex(t.tx.Hash()))
	t2 := &SigListRecord{
		IKey: newListKey(sigList.ID()),
		Op:   db.NewOp(Op),
		m:    sigList,
	}

	return []db.Record{t2}, nil
}

// DailyLimit更新，DailyLimit在Submit和Confirm阶段都可能有变化
func (t *msConvert) TyLogDailyLimitUpdate(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptAccDailyLimitUpdate
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogDailyLimitUpdate log")
	}

	limit := makeLimit(l.CurDailyLimit, l.MultiSigAddr)
	if op == db.SeqTypeDel {
		limit = makeLimit(l.PrevDailyLimit, l.MultiSigAddr)
	}

	t2 := &MSUpdateLimitRecord{
		IKey: newKey(limit.ID()),
		Op:   db.NewOp(db.OpAdd),
		u:    limit,
	}

	return []db.Record{t2}, nil
}

func (t *msConvert) makeSigList(l *pty.MultiSigTxOwner, isCreator bool, executed bool, txHash string) *SigList {
	sigList := SigList{
		Address: l.MultiSigAddr,
		TxID:    l.Txid,
		Owner:   l.ConfirmedOwner.OwnerAddr,
		Weight:  l.ConfirmedOwner.Weight,
		// 角色和状态: 是否为创建者，是否是促成交易执行的
		Creator:  isCreator,
		Executed: executed,
		TxHash:   txHash,
	}
	return &sigList
}

//在Submit提交交易阶段才会有更新
func (t *msConvert) TyLogMultiSigTx(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptMultiSigTx
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogMultiSigTx log")
	}

	var records []db.Record
	creator := false
	if len(l.TxHash) > 0 {
		tx := &SigTx{
			TxHash: l.TxHash,
			// 不同类型的多重签名
			Type:    toStringSigTxType(l.TxType),
			Address: l.MultiSigTxOwner.MultiSigAddr,
			TxID:    l.MultiSigTxOwner.Txid,
			Detail:  t.detail,
		}
		creator = true
		record := &SigTxRecord{
			IKey: newTxKey(tx.ID()),
			Op:   db.NewOp(op),
			m:    tx,
		}
		records = append(records, record)
	}

	executed := l.CurExecuted
	if op == db.SeqTypeDel {
		executed = l.PrevExecuted
	}

	txSig := t.makeSigList(l.MultiSigTxOwner, creator, executed, common.ToHex(t.tx.Hash()))
	record := &SigListRecord{
		IKey: newListKey(txSig.ID()),
		Op:   db.NewOp(op),
		m:    txSig,
	}
	records = append(records, record)

	return records, nil
}

//txcount只在在Submit阶段提交新的交易是才会增加计数
func (t *msConvert) TyLogTxCountUpdate(v []byte, op int) ([]db.Record, error) {
	var l pty.ReceiptTxCountUpdate
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "Decode TyLogTxCountUpdate log")
	}

	count := MSUpdateTxCount{
		Address: l.MultiSigAddr,
		TxCount: l.CurTxCount,
	}
	if op == db.SeqTypeDel {
		count.TxCount = count.TxCount - 1
	}

	t2 := &MSUpdateTxCountRecord{
		IKey: newKey(msID(l.MultiSigAddr)),
		Op:   db.NewOp(db.OpUpdate),
		u:    &count,
	}

	return []db.Record{t2}, nil
}

func (t *msConvert) setupAsset(exec, symbol string) {
	if symbol == types.BTY {
		symbol = "bty"
	}
	t.accountIDAsset = account.Account{
		AssetSymbol: symbol,
		AssetExec:   exec,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
	}
}
