// txinfo prints detailed transaction information using the same RPC + parsing
// logic as erc20Scaner/scanner (receipt, block, ERC20 Transfer logs, contract
// creation ERC20 probe, DB-shaped previews, balanceOf snapshots, Approval events).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/33cn/externaldb/erc20Scaner/txinspect"
)

func main() {
	rpcURL := flag.String("rpc", "", "Ethereum JSON-RPC URL (required)")
	asJSON := flag.Bool("json", false, "print single JSON object")
	flag.Parse()
	args := flag.Args()
	if *rpcURL == "" || len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s -rpc <url> [flags] <txHash>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	raw := args[0]
	raw = strings.TrimSpace(raw)
	if raw == "" {
		fmt.Fprintf(os.Stderr, "tx hash is required\n")
		os.Exit(2)
	}
	h := common.HexToHash(raw)

	c := new(txinspect.Client)
	c.ConnectEth(*rpcURL)
	defer c.CloseConnect()

	rep, err := txinspect.NewAnalyzer(c).Inspect(h)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inspect: %v\n", err)
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

func printHuman(r *txinspect.Report) {
	fmt.Printf("=== Transaction ===\n")
	fmt.Printf("Hash:              %s\n", r.TxHash)
	if r.ScannerSkipReason != "" {
		fmt.Printf("Scanner note:      %s\n", r.ScannerSkipReason)
	}
	fmt.Printf("Scanner processes: %v\n", r.ScannerWouldProcess)

	if r.Block != nil {
		fmt.Printf("\n=== Block ===\n")
		fmt.Printf("Number:     %d\n", r.Block.Number)
		fmt.Printf("Hash:       %s\n", r.Block.Hash)
		fmt.Printf("Time (CST): %s\n", r.Block.TimeCST)
		fmt.Printf("Tx index:   %d\n", r.Block.TxIndex)
	}

	if r.Transaction != nil {
		t := r.Transaction
		fmt.Printf("\n=== Tx fields (scanner / DB related) ===\n")
		fmt.Printf("From:           %s\n", t.From)
		fmt.Printf("To:             %s\n", nullStr(t.To))
		fmt.Printf("Nonce:          %d\n", t.Nonce)
		fmt.Printf("Type:           %d\n", t.Type)
		fmt.Printf("ChainID:        %s\n", nullStr(t.ChainID))
		fmt.Printf("Gas limit:      %d\n", t.Gas)
		fmt.Printf("Gas price:      %s\n", nullStr(t.GasPrice))
		if t.GasTipCap != "" {
			fmt.Printf("Max priority:   %s\n", t.GasTipCap)
		}
		if t.GasFeeCap != "" {
			fmt.Printf("Max fee:        %s\n", t.GasFeeCap)
		}
		fmt.Printf("Value (wei):    %s\n", t.ValueWei)
		fmt.Printf("Func selector:  %s\n", nullStr(t.FuncSelector))
		fmt.Printf("Data (hex):     %s\n", t.DataHex)
	}

	if r.Receipt != nil {
		rc := r.Receipt
		fmt.Printf("\n=== Receipt ===\n")
		fmt.Printf("Status:            %d\n", rc.Status)
		fmt.Printf("Gas used:          %d\n", rc.GasUsed)
		fmt.Printf("Cumulative gas:    %d\n", rc.CumulativeGasUsed)
		fmt.Printf("Tx index:          %d\n", rc.TransactionIndex)
		fmt.Printf("Block number:      %d\n", rc.BlockNumber)
		fmt.Printf("Block hash:        %s\n", rc.BlockHash)
		fmt.Printf("Contract address:  %s\n", nullStr(rc.ContractAddress))
		fmt.Printf("Logs count:        %d\n", rc.LogsCount)
	}

	fmt.Printf("\n=== All logs ===\n")
	for _, lg := range r.AllLogs {
		fmt.Printf("- #%d addr=%s removed=%v transferShape=%v approvalShape=%v\n",
			lg.Index, lg.Address, lg.Removed, lg.IsTransferEvent, lg.IsApprovalEvent)
		for j, tp := range lg.Topics {
			fmt.Printf("    topic[%d]: %s\n", j, tp)
		}
		fmt.Printf("    data: %s\n", lg.Data)
	}

	if r.ContractCreation != nil {
		cc := r.ContractCreation
		fmt.Printf("\n=== Contract creation (scanner path) ===\n")
		fmt.Printf("Deployed: %s\n", cc.DeployedAddress)
		fmt.Printf("ERC20 check: isERC20=%v core=%s optional=%s ext=%s err=%q\n",
			cc.ERC20.IsERC20, cc.ERC20.CoreFound, cc.ERC20.OptionalFound, cc.ERC20.ExtensionFound, cc.ERC20.Error)
		if cc.Name != "" || cc.Symbol != "" {
			fmt.Printf("Token name:     %s\n", cc.Name)
			fmt.Printf("Token symbol:   %s\n", cc.Symbol)
			fmt.Printf("Decimals:       %s\n", cc.Decimals)
			fmt.Printf("Total supply:   %s\n", cc.TotalSupply)
		}
		if cc.DBContractPreview != nil {
			p := cc.DBContractPreview
			fmt.Printf("\n--- DB contracts row (preview) ---\n")
			fmt.Printf("%+v\n", p)
		}
	}

	if r.ERC20Path != nil {
		e := r.ERC20Path
		fmt.Printf("\n=== ERC20 transfer path (scanner parseERC20Transfer) ===\n")
		fmt.Printf("Calldata selector: %s (direct transfer/from: %v)\n", e.CalldataFuncSelector, e.IsDirectTransferCall)
		if e.ScannerNote != "" {
			fmt.Printf("Note: %s\n", e.ScannerNote)
		}
		for _, pc := range e.PerTokenContract {
			fmt.Printf("- token %s direct=%v func=%s selector=%s txToEqualsToken=%v\n",
				pc.TokenAddress, pc.IsDirectCall, pc.FuncName, pc.FinalFuncSelector, pc.TxToEqualsToken)
		}
		if e.DBTransactionPreview != nil {
			fmt.Printf("\n--- DB transactions row (preview, see scannerNote) ---\n")
			fmt.Printf("%+v\n", e.DBTransactionPreview)
		}
		for _, tr := range e.Transfers {
			fmt.Printf("\n--- Transfer log #%d ---\n", tr.LogIndex)
			fmt.Printf("Token: %s (%s / %s, decimals=%d)\n", tr.TokenAddress, tr.TokenInfo.Name, tr.TokenInfo.Symbol, tr.TokenInfo.Decimals)
			fmt.Printf("From:  %s\n", tr.From)
			fmt.Printf("To:    %s\n", tr.To)
			fmt.Printf("Value raw: %s\n", tr.ValueRaw)
			fmt.Printf("Formatted: %s\n", tr.ValueFormatted)
			if tr.FromBalance != nil {
				fmt.Printf("balanceOf(from) now: %s\n", *tr.FromBalance)
			}
			if tr.ToBalance != nil {
				fmt.Printf("balanceOf(to) now:   %s\n", *tr.ToBalance)
			}
			if tr.BalanceErr != "" {
				fmt.Printf("balance query: %s\n", tr.BalanceErr)
			}
			if tr.DBEventPreview != nil {
				fmt.Printf("DB events row preview: %+v\n", tr.DBEventPreview)
			}
		}
	}
	if r.ApprovalPath != nil {
		a := r.ApprovalPath
		fmt.Printf("\n=== Approval path (scanner updateAllowanceInDB) ===\n")
		fmt.Printf("DB action: %s\n", a.DBAllowanceAction)
		for _, ap := range a.Approvals {
			fmt.Printf("\n--- Approval log #%d ---\n", ap.LogIndex)
			fmt.Printf("Token: %s (%s / %s, decimals=%d)\n", ap.TokenAddress, ap.TokenInfo.Name, ap.TokenInfo.Symbol, ap.TokenInfo.Decimals)
			fmt.Printf("Owner:   %s\n", ap.Owner)
			fmt.Printf("Spender: %s\n", ap.Spender)
			fmt.Printf("Amount raw: %s\n", ap.AmountRaw)
			fmt.Printf("Formatted:  %s\n", ap.AmountFormatted)
			if ap.DBEventPreview != nil {
				fmt.Printf("DB events row preview: %+v\n", ap.DBEventPreview)
			}
		}
	}
}

func nullStr(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
