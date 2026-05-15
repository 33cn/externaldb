package blockalign

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// sameFromForTiebreak 比较 chain33 的 From 字符串与 eth Sender 的 Hex 字符串；不做地址格式互转。
func sameFromForTiebreak(chain33From, ethFromHex string) bool {
	chain33From = strings.TrimSpace(chain33From)
	ethFromHex = strings.TrimSpace(ethFromHex)
	if chain33From == "" || ethFromHex == "" {
		return false
	}
	return strings.EqualFold(chain33From, ethFromHex)
}

// SlotEthExpect 表示 chain33 某一交易下标在对齐时的期望信息（通常由 chain33 types.Transaction 的 Nonce / From 得到）。
// NeedEth 为 false 表示该槽位不是 EVM 执行器，不需要对应 eth 交易体（对齐结果为 nil）。
// Nonce 为主要对齐键，与 eth 区块体中交易的 eth nonce 按顺序匹配。
// From 为 chain33 tx.From() 原样字符串；仅在连续同 nonce 时与 eth Sender 的 Hex() 字符串比较，不将 chain33 地址转成 eth 地址格式。
type SlotEthExpect struct {
	NeedEth bool
	From    string
	Nonce   uint64
}

// AlignEthTxsByNonce 将 eth_getBlock 返回的交易体，按顺序与 chain33 各槽位的期望 Nonce 对齐。
//
// 背景：chain33 区块交易数可能多于 eth 区块体中的笔数，但相对顺序一致；主键为 Nonce 顺序消费，不以 (chain33From,ethFrom) 做主键映射。
// 做法：按区块体中交易的出现顺序依次消费；对每个需要 eth 的槽位，在剩余候选里从前往后找 nonce 与期望一致的一笔；
// 若当前位置与下一笔候选 nonce 相同（连续同 nonce），再用 chain33 From 字符串与两笔候选 eth Sender 的 Hex() 比较；仅第二笔匹配且第一笔不匹配时选第二笔，否则取第一笔。
// 非 EVM 槽位输出 nil；缺少对应 nonce 的 eth 体则该槽位为 nil。
func AlignEthTxsByNonce(block *types.Block, expects []SlotEthExpect) ([]*types.Transaction, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block")
	}
	if len(expects) == 0 {
		return nil, nil
	}
	eth := block.Transactions()
	type candView struct {
		tx       *types.Transaction
		nonce    uint64
		fromHex  string // eth Sender 转为 Hex 字符串，与 chain33 From 字符串比较
	}

	var cands []candView
	for i := 0; i < eth.Len(); i++ {
		tx := eth[i]
		if tx == nil {
			continue
		}
		signer := types.LatestSignerForChainID(tx.ChainId())
		from, err := types.Sender(signer, tx)
		if err != nil {
			return nil, fmt.Errorf("eth tx[%d] hash=%s: sender: %w", i, tx.Hash().Hex(), err)
		}
		cands = append(cands, candView{tx: tx, nonce: tx.Nonce(), fromHex: from.Hex()})
	}

	out := make([]*types.Transaction, len(expects))
	scan := 0
	for i := range expects {
		if !expects[i].NeedEth {
			out[i] = nil
			continue
		}
		wantN := expects[i].Nonce
		wantF := expects[i].From

		if scan >= len(cands) {
			out[i] = nil
			continue
		}
		// 跳过 nonce 仍小于期望的候选（多余或已过期的 eth 笔，顺序消费）
		for scan < len(cands) && cands[scan].nonce < wantN {
			scan++
		}
		if scan >= len(cands) || cands[scan].nonce > wantN {
			out[i] = nil
			continue
		}
		// cands[scan].nonce == wantN
		pick := scan
		if scan+1 < len(cands) && cands[scan+1].nonce == wantN {
			if sameFromForTiebreak(wantF, cands[scan+1].fromHex) && !sameFromForTiebreak(wantF, cands[scan].fromHex) {
				pick = scan + 1
			}
		}
		out[i] = cands[pick].tx
		scan = pick + 1
	}
	return out, nil
}

// BlockIndexTxFetcher loads a transaction by its canonical index inside a block.
// Implemented by scanner.Client and txinspect.Client (block body, then block+index RPCs,
// then eth_getBlockReceipts + receipt.transactionIndex + eth_getTransactionByHash when needed).
type BlockIndexTxFetcher interface {
	TransactionAtCanonicalIndex(block *types.Block, txIndex uint64) (*types.Transaction, error)
}

// AlignEthTxsWithSeqCount maps Ethereum *types.Transaction into chain33/seq tx indices (len = seqTxCount).
// Rules (same intent as parseBlockFromES):
//   - When len(eth txs) > seqTxCount: error.
//   - When len(eth txs) == seqTxCount: copy by index from block only — no extra RPC (block already has one slot per index).
//   - When len(eth txs) < seqTxCount: prefer AlignEthTxsByNonce with chain33-derived SlotEthExpect when the
//     node omits txs from getBlock; otherwise TransactionAtCanonicalIndex fills each aligned[i] (may mismatch order).
func AlignEthTxsWithSeqCount(block *types.Block, seqTxCount int, f BlockIndexTxFetcher) ([]*types.Transaction, error) {
	if seqTxCount < 0 {
		return nil, fmt.Errorf("invalid seqTxCount %d", seqTxCount)
	}
	eth := block.Transactions()
	aligned := make([]*types.Transaction, seqTxCount)

	if eth.Len() > seqTxCount {
		return nil, fmt.Errorf("eth block has more txs than seq: eth=%d seq=%d", eth.Len(), seqTxCount)
	}

	if eth.Len() == seqTxCount {
		for i := 0; i < seqTxCount; i++ {
			aligned[i] = eth[i]
		}
		return aligned, nil
	}

	bn := block.NumberU64()
	for i := 0; i < seqTxCount; i++ {
		tx, err := f.TransactionAtCanonicalIndex(block, uint64(i))
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				aligned[i] = nil
				continue
			}
			return nil, fmt.Errorf("AlignEthTxs: TransactionAtCanonicalIndex block=%d index=%d: %w", bn, i, err)
		}
		if tx == nil {
			aligned[i] = nil
			continue
		}
		aligned[i] = tx
	}
	return aligned, nil
}
