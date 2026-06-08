package txinspect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/externaldb/erc20Scaner/blockalign"
	"github.com/33cn/externaldb/erc20Scaner/txparser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BlockReport aggregates per-transaction inspection for one block height (same logic paths as Inspect).
type BlockReport struct {
	Number       uint64          `json:"number"`
	Hash         string          `json:"hash"`
	ParentHash   string          `json:"parentHash"`
	TimeUnix     uint64          `json:"timeUnix"`
	TimeCST      string          `json:"timeAsiaShanghai"`
	TxCount      int             `json:"txCount"`
	GasUsed      uint64          `json:"gasUsed,omitempty"`
	Transactions []*TxSlotReport `json:"transactions"`
}

// TxSlotReport is one position in the block: full Report or RPC/analysis error.
type TxSlotReport struct {
	Index  int         `json:"index"`
	TxHash string      `json:"txHash"`
	Error  string      `json:"error,omitempty"`
	Diag   *TxSlotDiag `json:"diag,omitempty"`
	Report *Report     `json:"report,omitempty"`
}

// TxSlotDiag is filled when receipt lookup fails; nonce and gas limit for debugging.
type TxSlotDiag struct {
	Nonce uint64 `json:"nonce"`
	Gas   uint64 `json:"gas"`
}

// InspectBlock loads the block at height and runs the same analysis as Inspect for each tx slot.
// seqTxCount: use the same value as parseBlockFromES len(detail.Block.Txs) when debugging
// eth-returned tx count < chain33 tx count; 0 means use eth block tx count only (no expansion).
func (a *Analyzer) InspectBlock(height uint64, seqTxCount int) (*BlockReport, error) {
	block, txHashes, err := a.c.BlockByNumber(height)
	if err != nil {
		return nil, fmt.Errorf("BlockByNumber %d: %w", height, err)
	}
	if block == nil {
		return nil, fmt.Errorf("nil block at height %d", height)
	}

	n := seqTxCount
	if n <= 0 {
		n = block.Transactions().Len()
	}
	aligned, err := blockalign.AlignEthTxsWithSeqCount(block, n, a.c)
	if err != nil {
		return nil, fmt.Errorf("AlignEthTxsWithSeqCount seqTxCount=%d (parseBlockFromES uses len(chain33.Txs)): %w", n, err)
	}

	bt := time.Unix(int64(block.Time()), 0)
	cst, _ := time.LoadLocation("Asia/Shanghai")

	out := &BlockReport{
		Number:       block.NumberU64(),
		Hash:         block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		TimeUnix:     block.Time(),
		TimeCST:      bt.In(cst).Format("2006-01-02 15:04:05"),
		GasUsed:      block.GasUsed(),
		Transactions: nil,
	}

	out.TxCount = len(aligned)
	for i := 0; i < len(aligned); i++ {
		tx := aligned[i]
		var knownHash string
		if tx == nil {
			// txHashes is parallel to the node's tx slice, not necessarily to aligned[i]
			// when len(eth txs) < seqTxCount; resolve by same index fetch as AlignEthTxsWithSeqCount.
			if block.Transactions().Len() == len(aligned) && txHashes != nil && i < len(txHashes) && txHashes[i] != (common.Hash{}) {
				knownHash = txHashes[i].Hex()
			} else {
				knownHash = hashAtSeqIndexFromEthBlock(block, i, a.c)
			}
		}
		if tx == nil {
			msg := "skipped: no transaction body from node (non-standard RPC or fetch failed)"
			if knownHash != "" {
				msg = fmt.Sprintf("%s; list_hash=%s", msg, knownHash)
			} else {
				msg = fmt.Sprintf("%s; index=%d (hash not parsed)", msg, i)
			}
			out.Transactions = append(out.Transactions, &TxSlotReport{
				Index:  i,
				TxHash: knownHash,
				Error:  msg,
			})
			continue
		}
		slot := &TxSlotReport{Index: i, TxHash: tx.Hash().Hex()}
		receipt, err := a.c.TxReceipt(tx.Hash())
		if err != nil {
			slot.Error = fmt.Sprintf("TransactionReceipt: %v", err)
			slot.Diag = &TxSlotDiag{Nonce: tx.Nonce(), Gas: tx.Gas()}
			out.Transactions = append(out.Transactions, slot)
			continue
		}
		if receipt == nil {
			slot.Error = "nil receipt"
			slot.Diag = &TxSlotDiag{Nonce: tx.Nonce(), Gas: tx.Gas()}
			out.Transactions = append(out.Transactions, slot)
			continue
		}
		rep, err := a.buildReport(tx, receipt, block, i)
		if err != nil {
			slot.Error = err.Error()
		} else {
			slot.Report = rep
		}
		out.Transactions = append(out.Transactions, slot)
	}
	return out, nil
}

// hashAtSeqIndexFromEthBlock returns the canonical tx hash at seqIndex (same RPC path as AlignEthTxsWithSeqCount).
func hashAtSeqIndexFromEthBlock(block *types.Block, seqIndex int, f blockalign.BlockIndexTxFetcher) string {
	tx, err := f.TransactionAtCanonicalIndex(block, uint64(seqIndex))
	if err != nil || tx == nil {
		return ""
	}
	return tx.Hash().Hex()
}

// buildReport fills Report from tx + receipt + block already loaded (same semantics as Inspect).
func (a *Analyzer) buildReport(tx *types.Transaction, receipt *types.Receipt, block *types.Block, txIdx int) (*Report, error) {
	rep := &Report{
		TxHash: tx.Hash().Hex(),
		Block:  blockView(block, txIdx),
		Receipt: &ReceiptView{
			Status:            receipt.Status,
			CumulativeGasUsed: receipt.CumulativeGasUsed,
			GasUsed:           receipt.GasUsed,
			TransactionIndex:  uint64(receipt.TransactionIndex),
			BlockNumber:       receipt.BlockNumber.Uint64(),
			BlockHash:         receipt.BlockHash.Hex(),
			TxHash:            receipt.TxHash.Hex(),
			LogsCount:         len(receipt.Logs),
		},
		AllLogs: make([]LogView, 0, len(receipt.Logs)),
	}

	if receipt.ContractAddress != (common.Address{}) {
		rep.Receipt.ContractAddress = receipt.ContractAddress.Hex()
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		rep.ScannerWouldProcess = false
		rep.ScannerSkipReason = fmt.Sprintf("receipt status %d (scanner only processes successful txs)", receipt.Status)
	} else {
		rep.ScannerWouldProcess = true
	}

	var fromAddr common.Address
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if s, err := types.Sender(signer, tx); err == nil {
			fromAddr = s
		}
	}

	ts := &TxSummary{
		From:           fromAddr.Hex(),
		Nonce:          tx.Nonce(),
		Gas:            tx.Gas(),
		ValueWei:       tx.Value().String(),
		Type:           tx.Type(),
		DataHex:        hex.EncodeToString(tx.Data()),
		IsContractCall: len(tx.Data()) >= 4,
	}
	if tx.ChainId() != nil {
		ts.ChainID = tx.ChainId().String()
	}
	if tx.To() != nil {
		ts.To = tx.To().Hex()
	}
	if gp := tx.GasPrice(); gp != nil {
		ts.GasPrice = gp.String()
	}
	if tx.Type() == types.DynamicFeeTxType {
		if tx.GasTipCap() != nil {
			ts.GasTipCap = tx.GasTipCap().String()
		}
		if tx.GasFeeCap() != nil {
			ts.GasFeeCap = tx.GasFeeCap().String()
		}
	}
	if len(tx.Data()) >= 4 {
		ts.FuncSelector = hex.EncodeToString(tx.Data()[:4])
	}
	rep.Transaction = ts

	transferEventID := txparser.TransferEventID()
	approvalEventID := txparser.ApprovalEventID()

	for i, lg := range receipt.Logs {
		lv := LogView{
			Index:       uint(i),
			Address:     lg.Address.Hex(),
			Topics:      topicsHex(lg.Topics),
			Data:        hex.EncodeToString(lg.Data),
			Removed:     lg.Removed,
			BlockNumber: lg.BlockNumber,
			TxHash:      lg.TxHash.Hex(),
			TxIndex:     lg.TxIndex,
			BlockHash:   lg.BlockHash.Hex(),
		}
		if len(lg.Topics) >= 3 && lg.Topics[0] == transferEventID {
			lv.IsTransferEvent = true
		}
		if len(lg.Topics) >= 3 && approvalEventID != (common.Hash{}) && lg.Topics[0] == approvalEventID {
			lv.IsApprovalEvent = true
		}
		rep.AllLogs = append(rep.AllLogs, lv)
	}

	if !rep.ScannerWouldProcess {
		return rep, nil
	}

	blockTime := time.Unix(int64(block.Time()), 0)
	cst, _ := time.LoadLocation("Asia/Shanghai")

	if tx.To() == nil {
		cc := a.buildContractCreationView(receipt, block, tx, blockTime, cst)
		rep.ContractCreation = cc
		return rep, nil
	}

	ep := a.buildERC20Path(tx, receipt, block, fromAddr, blockTime, cst)
	if ep != nil {
		rep.ERC20Path = ep
	}
	ap := a.buildApprovalPath(tx, receipt, block, blockTime, cst)
	if ap != nil {
		rep.ApprovalPath = ap
	}
	if ep == nil && ap == nil {
		rep.ScannerSkipReason = strings.TrimSpace(rep.ScannerSkipReason + "; no ERC20 Transfer or Approval logs in receipt (scanner returns without DB writes for this tx)")
	}

	return rep, nil
}
