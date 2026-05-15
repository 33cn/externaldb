package main

import (
	"fmt"
	"strings"

	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/erc20Scaner/blockalign"
)

// buildSlotEthExpectsFromBlockDetail 为每个 chain33 交易下标构造对齐期望。
// EVM 槽位：Nonce、From 直接取自 chain33 types.Transaction（GetNonce / From），不解析 payload.Note。
// From 为 tx.From() 原样字符串，对齐时由 eth 侧 Sender.Hex() 转成字符串再比较，不把 chain33 地址转成 eth 地址格式。
func buildSlotEthExpectsFromBlockDetail(detail *chain33types.BlockDetail) ([]blockalign.SlotEthExpect, error) {
	if detail.Block == nil {
		return nil, fmt.Errorf("nil block in detail")
	}
	txs := detail.Block.Txs
	out := make([]blockalign.SlotEthExpect, len(txs))
	for i, tx := range txs {
		if tx == nil || !isEvmExecer(string(tx.Execer)) {
			out[i] = blockalign.SlotEthExpect{NeedEth: false}
			continue
		}
		n := tx.GetNonce()
		if n < 0 {
			return nil, fmt.Errorf("chain33 tx[%d]: negative nonce %d", i, n)
		}
		out[i] = blockalign.SlotEthExpect{
			NeedEth: true,
			Nonce:   uint64(n),
			From:    strings.TrimSpace(tx.From()),
		}
	}
	return out, nil
}
