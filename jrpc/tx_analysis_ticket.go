package main

import (
	coinsTypes "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"

	types2 "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

const EvmCallGoAddr = "0x0000000000000000000000000000000000200005"
const TicketExecName = "ticket"
const CoinsExecName = "coins"

func parseEvmCallGoTx(info *EvmTxInfo, para []byte) *EvmTxInfo {
	var tx types.Transaction
	err := types.Decode(para, &tx)
	if err != nil {
		log.Error("parseEvmCallGoTx", "err", err.Error())
		return info
	}

	// evm para action
	// para 有可能是各种内置chain合约的action
	log.Info("parseEvmCallGoTx", "exec", string(tx.Execer))
	if string(tx.Execer) == TicketExecName {
		// 1、 para => ticket action
		var action types2.TicketAction
		err = types.Decode(tx.Payload, &action)
		if err != nil {
			log.Error("parseEvmCallGoTx ticket decode", "err", err.Error())
			return info
		}
		switch action.Ty {
		case types2.TicketActionBind:
			var act string
			if action.GetTbind().ReturnAddress == action.GetTbind().MinerAddress {
				act = "ticketBind"
			} else {
				act = "ticketUnbind"
			}
			log.Info("parseEvmCallGoTx ticket decode", "act", act)

		}
		return info
	} else if string(tx.Execer) == CoinsExecName {
		// 2 、 coins action
		var cAction coinsTypes.CoinsAction
		err := types.Decode(tx.Payload, &cAction)
		if err != nil {
			log.Error("parseEvmCallGoTx coins decode", "err", err.Error())
			return info
		}
		switch cAction.Ty {
		case coinsTypes.CoinsActionWithdraw:
			info.Amount = uint64(cAction.GetWithdraw().Amount)
			info.Asset.Amount = cAction.GetWithdraw().Amount
		case coinsTypes.CoinsActionTransferToExec:
			info.Amount = uint64(cAction.GetTransferToExec().Amount)
			info.Asset.Amount = cAction.GetWithdraw().Amount
		default:
			log.Info("switch ty", "in", cAction.Ty, "1th", coinsTypes.CoinsActionWithdraw, "2nd", coinsTypes.CoinsActionTransferToExec)
		}
		log.Info("parseEvmCallGoTx coins", "Amount", info.Amount, "asset", info.Asset)
		return info
	}

	return info
}
