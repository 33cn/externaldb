package account

import (
	"fmt"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
	"github.com/33cn/externaldb/db"
)

func accountConvert(acc *types.Account, t string) *Detall {
	return &Detall{
		Frozen:  acc.Frozen,
		Balance: acc.Balance,
		Total:   acc.Frozen + acc.Balance,
		Type:    t,
		Address: acc.Addr,
	}
}

func fromAccount(l *types.ReceiptAccountTransfer, op int) *Detall {
	accType := AccountPersonage
	if _, ok := execAddress[l.Current.Addr]; ok {
		accType = AccountContract
	}
	if op == db.SeqTypeDel {
		return accountConvert(l.Prev, accType)
	}
	return accountConvert(l.Current, accType)
}

// coins/ticket-addr
func fromExecAccount(l *types.ReceiptExecAccountTransfer, op int) *Detall {
	acc := l.Current
	if op == db.SeqTypeDel {
		acc = l.Prev
	}
	detail := accountConvert(acc, AccountContractInternal)
	detail.Exec = l.ExecAddr
	return detail
}

// LogFeeConvert LogFeeConvert
func LogFeeConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogTransferConvert LogTransferConvert
func LogTransferConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogDepositConvert LogDepositConvert
func LogDepositConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogExecTransferConvert LogExecTransferConvert
func LogExecTransferConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptExecAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromExecAccount(&l, op), nil
	}
	return nil, err
}

// LogExecWithdrawConvert LogExecWithdrawConvert
func LogExecWithdrawConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptExecAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromExecAccount(&l, op), nil
	}
	return nil, err
}

// LogExecDepositConvert LogExecDepositConvert
func LogExecDepositConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptExecAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromExecAccount(&l, op), nil
	}
	return nil, err
}

// LogExecFrozenConvert LogExecFrozenConvert
func LogExecFrozenConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptExecAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromExecAccount(&l, op), nil
	}
	return nil, err
}

// LogExecActiveConvert LogExecActiveConvert
func LogExecActiveConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptExecAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromExecAccount(&l, op), nil
	}
	return nil, err
}

// LogGenesisTransferConvert LogGenesisTransferConvert
func LogGenesisTransferConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogGenesisDepositConvert LogGenesisDepositConvert
func LogGenesisDepositConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogMintConvert LogMintConvert
func LogMintConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// LogBurnConvert LogBurnConvert
func LogBurnConvert(v []byte, op int) (*Detall, error) {
	var l types.ReceiptAccountTransfer
	err := types.Decode(v, &l)
	if err == nil {
		return fromAccount(&l, op), nil
	}
	return nil, err
}

// AssetLogConvert  convert asset log
func AssetLogConvert(ty int32, v []byte, op int) (*Detall, error) {
	detail, err := assetLogConvert(ty, v, op)
	if err != nil {
		return nil, errors.Wrapf(err, "convert asset log: %d", ty)
	}
	return detail, nil
}

func assetLogConvert(ty int32, v []byte, op int) (*Detall, error) {
	if ty == types.TyLogFee {
		return LogFeeConvert(v, op)
	} else if ty == types.TyLogTransfer {
		return LogTransferConvert(v, op)
	} else if ty == types.TyLogDeposit {
		return LogDepositConvert(v, op)
	} else if ty == types.TyLogExecTransfer {
		return LogExecTransferConvert(v, op)
	} else if ty == types.TyLogExecWithdraw {
		return LogExecWithdrawConvert(v, op)
	} else if ty == types.TyLogExecDeposit {
		return LogExecDepositConvert(v, op)
	} else if ty == types.TyLogExecFrozen {
		return LogExecFrozenConvert(v, op)
	} else if ty == types.TyLogExecActive {
		return LogExecActiveConvert(v, op)
	} else if ty == types.TyLogGenesisTransfer {
		return LogGenesisTransferConvert(v, op)
	} else if ty == types.TyLogGenesisDeposit {
		return LogGenesisDepositConvert(v, op)
	} else if ty == types.TyLogMint {
		return LogMintConvert(v, op)
	} else if ty == types.TyLogBurn {
		return LogBurnConvert(v, op)
	}
	return nil, notSupport(ty, v)
}

func notSupport(logType int32, json []byte) (err error) {
	return errors.New("notSupport" + fmt.Sprintf("-log:%d", logType))
}
