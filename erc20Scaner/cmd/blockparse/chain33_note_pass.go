package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// chain33NoteReport is the second pass: EVM txs from chain33 block, eth tx decoded from payload.Note (same as jrpc/tx_analysis.go).
type chain33NoteReport struct {
	GRPC   string            `json:"grpc,omitempty"`
	Height int64             `json:"height"`
	Items  []chain33NoteItem `json:"items"`
	Error  string            `json:"error,omitempty"`
}

type chain33NoteItem struct {
	Index         int    `json:"index"`
	Chain33TxHash string `json:"chain33TxHash"`
	Execer        string `json:"execer"`
	Skipped       string `json:"skipped,omitempty"`
	Error         string `json:"error,omitempty"`
	// EvmID is common.ToHex(ntx.Hash().Bytes()) in tx_analysis.go (evmIdStr).
	EvmID           string `json:"evmID,omitempty"`
	EvmEthTxHash    string `json:"evmEthTxHash,omitempty"`
	NoteLenHexChars int    `json:"noteLenHexChars,omitempty"`
	Gas             uint64 `json:"gas,omitempty"`
	Nonce           uint64 `json:"nonce,omitempty"`
	To              string `json:"to,omitempty"`
	ValueWei        string `json:"valueWei,omitempty"`
	Type            uint8  `json:"type,omitempty"`
}

func isEvmExecer(execer string) bool {
	e := strings.TrimSpace(execer)
	return e == "evm" || strings.HasSuffix(e, ".evm")
}

// runChain33NotePass fetches one block from chain33 gRPC and decodes EVM note → go-ethereum tx (tx_analysis path).
func runChain33NotePass(grpcAddr string, height int64) *chain33NoteReport {
	out := &chain33NoteReport{GRPC: grpcAddr, Height: height}
	if grpcAddr == "" {
		return out
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024)),
	)
	if err != nil {
		out.Error = fmt.Sprintf("grpc dial: %v", err)
		return out
	}
	defer conn.Close()

	cli := chain33types.NewChain33Client(conn)
	reply, err := cli.GetBlocks(ctx, &chain33types.ReqBlocks{
		Start:    height,
		End:      height,
		IsDetail: true,
	})
	if err != nil {
		out.Error = fmt.Sprintf("GetBlocks: %v", err)
		return out
	}
	if reply == nil || !reply.IsOk {
		out.Error = "GetBlocks: empty or not ok"
		if reply != nil && len(reply.Msg) > 0 {
			out.Error += ": " + string(reply.Msg)
		}
		return out
	}

	var details chain33types.BlockDetails
	if err := chain33types.Decode(reply.Msg, &details); err != nil {
		out.Error = fmt.Sprintf("decode BlockDetails: %v", err)
		return out
	}
	if len(details.Items) == 0 {
		out.Error = "GetBlocks: no block items"
		return out
	}
	detail := details.Items[0]
	if detail.Block == nil || len(detail.Block.Txs) == 0 {
		out.Error = "block has no txs"
		return out
	}

	for i, tx := range detail.Block.Txs {
		item := chain33NoteItem{Index: i, Chain33TxHash: "0x" + hex.EncodeToString(tx.Hash()), Execer: string(tx.Execer)}
		if tx == nil {
			item.Skipped = "nil tx"
			out.Items = append(out.Items, item)
			continue
		}
		if !isEvmExecer(string(tx.Execer)) {
			item.Skipped = "not evm execer"
			out.Items = append(out.Items, item)
			continue
		}
		var payload proto.EVMContractAction
		if err := chain33types.Decode(tx.Payload, &payload); err != nil {
			item.Error = "decode EVMContractAction: " + err.Error()
			out.Items = append(out.Items, item)
			continue
		}
		note := strings.TrimSpace(payload.GetNote())
		item.NoteLenHexChars = len(note)
		if note == "" {
			item.Error = "empty payload.Note"
			out.Items = append(out.Items, item)
			continue
		}
		raw, err := hexutil.Decode(strings.TrimSpace(note))
		if err != nil {
			item.Error = "decode note hex: " + err.Error()
			out.Items = append(out.Items, item)
			continue
		}
		ntx := new(etypes.Transaction)
		if err := ntx.UnmarshalBinary(raw); err != nil {
			item.Error = "eth UnmarshalBinary(note): " + err.Error()
			out.Items = append(out.Items, item)
			continue
		}
		item.EvmID = ntx.Hash().Hex()
		item.EvmEthTxHash = item.EvmID
		item.Gas = ntx.Gas()
		item.Nonce = ntx.Nonce()
		item.Type = ntx.Type()
		if ntx.Value() != nil {
			item.ValueWei = ntx.Value().String()
		}
		if ntx.To() != nil {
			item.To = ntx.To().Hex()
		}
		out.Items = append(out.Items, item)
	}
	return out
}
