// blockparse prints block-level inspection using the same RPC + parsing logic as
// erc20Scaner/parse_tools (per-tx: receipt, ERC20 path, contract creation probe).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/33cn/externaldb/erc20Scaner/txinspect"
)

// blockparseOut wraps the original ETH-RPC inspection plus an optional chain33 Note→eth second pass.
type blockparseOut struct {
	EthPass      *txinspect.BlockReport `json:"ethPass"`
	Chain33Notes *chain33NoteReport     `json:"chain33NotePass,omitempty"`
}

func main() {
	rpcURL := flag.String("rpc", "", "Ethereum JSON-RPC URL (required)")
	height := flag.Uint64("height", 0, "block number to parse (decimal)")
	seqTxCount := flag.Int("seq-tx-count", 0, "chain33 seq tx count for alignment (same as len(detail.Block.Txs) in parseBlockFromES); 0 = eth block tx count only")
	chain33GRPC := flag.String("chain33-grpc", "", "optional chain33 gRPC host:port for second pass: fetch block by same height, filter evm execer, decode payload.Note → eth tx (jrpc/tx_analysis style)")
	asJSON := flag.Bool("json", false, "print single JSON object")
	flag.Parse()

	args := flag.Args()
	if *rpcURL == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -rpc <url> [-height N] [-seq-tx-count M] [flags] [N]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Block height: -height N or positional decimal N.\n")
		fmt.Fprintf(os.Stderr, "  When parseBlockFromES fails on a block, pass -seq-tx-count with len(chain33.Txs) from ES.\n")
		fmt.Fprintf(os.Stderr, "  Second pass: -chain33-grpc host:port decodes EVM txs from chain33 block (same height) via payload.Note.\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	h := *height
	if h == 0 && len(args) >= 1 {
		raw := strings.TrimSpace(args[0])
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid block height %q: %v\n", raw, err)
			os.Exit(2)
		}
		h = v
	}
	if h == 0 {
		fmt.Fprintf(os.Stderr, "block height is required (-height or positional)\n")
		os.Exit(2)
	}

	c := new(txinspect.Client)
	c.ConnectEth(*rpcURL)
	defer c.CloseConnect()

	rep, err := txinspect.NewAnalyzer(c).InspectBlock(h, *seqTxCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inspect block: %s\n", formatInspectErr(err))
		//os.Exit(1)
	}

	var noteRep *chain33NoteReport
	if strings.TrimSpace(*chain33GRPC) != "" {
		noteRep = runChain33NotePass(strings.TrimSpace(*chain33GRPC), int64(h))
	}

	out := &blockparseOut{EthPass: rep, Chain33Notes: noteRep}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "encode: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printHuman(rep)
	if noteRep != nil {
		printChain33NotePass(noteRep)
		printNonceMatch(rep, noteRep)
	}
}

// printNonceMatch compares chain33 EVM slot nonces against ETH transaction nonces
// and prints a matching summary table.
func printNonceMatch(eth *txinspect.BlockReport, c33 *chain33NoteReport) {
	fmt.Printf("\n=== Nonce Matching (chain33 EVM ↔ ETH) ===\n")

	// Collect ETH nonces with their indices
	type ethNonce struct {
		idx   int
		nonce uint64
		hash  string
	}
	var ethNonces []ethNonce
	for _, slot := range eth.Transactions {
		if slot.Diag != nil {
			ethNonces = append(ethNonces, ethNonce{idx: slot.Index, nonce: slot.Diag.Nonce, hash: slot.TxHash})
		} else if slot.Report != nil && slot.Report.Transaction != nil {
			ethNonces = append(ethNonces, ethNonce{idx: slot.Index, nonce: slot.Report.Transaction.Nonce, hash: slot.TxHash})
		}
	}

	// For each chain33 EVM slot, find matching ETH nonce
	for _, it := range c33.Items {
		if it.Skipped != "" {
			continue
		}
		matched := -1
		for _, e := range ethNonces {
			if e.nonce == it.Nonce {
				matched = e.idx
				break
			}
		}
		if matched >= 0 {
			fmt.Printf("  c33:%d nonce=%d → ETH[%d] %s ✓\n", it.Index, it.Nonce, matched, ethNonces[matched].hash)
		} else if it.Error != "" {
			fmt.Printf("  c33:%d nonce=%d → N/A (decode err: %s)\n", it.Index, it.Nonce, it.Error)
		} else {
			fmt.Printf("  c33:%d nonce=%d → NOT FOUND in %d ETH txs\n", it.Index, it.Nonce, len(ethNonces))
		}
	}
}

func printHuman(b *txinspect.BlockReport) {
	fmt.Printf("=== Block (eth RPC pass) ===\n")
	fmt.Printf("Number:     %d\n", b.Number)
	fmt.Printf("Hash:       %s\n", b.Hash)
	fmt.Printf("Parent:     %s\n", b.ParentHash)
	fmt.Printf("Time (CST): %s\n", b.TimeCST)
	fmt.Printf("Tx count:   %d\n", b.TxCount)
	if b.GasUsed > 0 {
		fmt.Printf("Gas used:   %d\n", b.GasUsed)
	}
	fmt.Printf("\n=== Transactions (scanner-style per tx) ===\n")
	for _, slot := range b.Transactions {
		fmt.Printf("\n--- [%d] %s ---\n", slot.Index, slot.TxHash)
		if slot.Error != "" {
			fmt.Printf("Error: %s\n", slot.Error)
			if slot.Diag != nil {
				printTxSlotDiag(slot.Diag)
			}
			continue
		}
		r := slot.Report
		if r == nil {
			fmt.Printf("(nil report)\n")
			continue
		}
		fmt.Printf("Scanner processes: %v\n", r.ScannerWouldProcess)
		if r.ScannerSkipReason != "" {
			fmt.Printf("Scanner note:      %s\n", r.ScannerSkipReason)
		}
		if r.Transaction != nil {
			t := r.Transaction
			logs := 0
			if r.Receipt != nil {
				logs = r.Receipt.LogsCount
			}
			fmt.Printf("From: %s  To: %s  nonce: %d  gas: %d  logs: %d\n", t.From, nullStr(t.To), t.Nonce, t.Gas, logs)
		}
		if r.ContractCreation != nil {
			cc := r.ContractCreation
			fmt.Printf("Contract create -> %s  ERC20: %v\n", cc.DeployedAddress, cc.ERC20.IsERC20)
		}
		if r.ERC20Path != nil {
			fmt.Printf("ERC20 transfers in receipt: %d\n", len(r.ERC20Path.Transfers))
			for i, tr := range r.ERC20Path.Transfers {
				if i >= 8 {
					fmt.Printf("  ... and %d more\n", len(r.ERC20Path.Transfers)-8)
					break
				}
				fmt.Printf("  [%d] %s %s -> %s  %s %s\n", tr.LogIndex, tr.TokenInfo.Symbol, tr.From, tr.To, tr.ValueFormatted, tr.TokenAddress)
			}
		}
	}
}

func printTxSlotDiag(d *txinspect.TxSlotDiag) {
	if d == nil {
		return
	}
	fmt.Printf("  diag.nonce: %d  gas: %d\n", d.Nonce, d.Gas)
}

func printChain33NotePass(n *chain33NoteReport) {
	fmt.Printf("\n=== Second pass: chain33 gRPC → EVM payload.Note → eth tx (tx_analysis style) ===\n")
	fmt.Printf("gRPC:   %s\n", n.GRPC)
	fmt.Printf("Height: %d\n", n.Height)
	if n.Error != "" {
		fmt.Printf("Error:  %s\n", n.Error)
		return
	}
	for _, it := range n.Items {
		fmt.Printf("\n--- [c33:%d] chain33_tx=%s execer=%s ---\n", it.Index, it.Chain33TxHash, it.Execer)
		if it.Skipped != "" {
			fmt.Printf("Skipped: %s\n", it.Skipped)
			continue
		}
		if it.Error != "" {
			fmt.Printf("Error: %s\n", it.Error)
			continue
		}
		fmt.Printf("evmID (ntx.Hash, same as tx_analysis evmIdStr): %s\n", it.EvmID)
		fmt.Printf("evm_eth_tx_hash: %s\n", it.EvmEthTxHash)
		fmt.Printf("gas=%d nonce=%d type=%d to=%s valueWei=%s note_hex_len=%d\n",
			it.Gas, it.Nonce, it.Type, nullStr(it.To), it.ValueWei, it.NoteLenHexChars)
	}
}

func nullStr(s string) string {
	if s == "" {
		return "(nil)"
	}
	return s
}

// formatInspectErr expands separators so multi-line RPC detail is readable on stderr.
func formatInspectErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	s = strings.ReplaceAll(s, " | ", "\n    ")
	return s
}
