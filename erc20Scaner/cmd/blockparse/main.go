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

func main() {
	rpcURL := flag.String("rpc", "", "Ethereum JSON-RPC URL (required)")
	height := flag.Uint64("height", 0, "block number to parse (decimal)")
	asJSON := flag.Bool("json", false, "print single JSON object")
	flag.Parse()

	args := flag.Args()
	if *rpcURL == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -rpc <url> [-height N] [flags] [N]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Block height: -height N or positional decimal N.\n")
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

	rep, err := txinspect.NewAnalyzer(c).InspectBlock(h)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inspect block: %s\n", formatInspectErr(err))
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			fmt.Fprintf(os.Stderr, "encode: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printHuman(rep)
}

func printHuman(b *txinspect.BlockReport) {
	fmt.Printf("=== Block ===\n")
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
			fmt.Printf("From: %s  To: %s  logs: %d\n", t.From, nullStr(t.To), logs)
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
