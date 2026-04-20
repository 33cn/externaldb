package txinspect

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// Client uses HTTP JSON-RPC with lenient error parsing (see jsonrpc.go). This avoids
// go-ethereum rpc.Client which rejects servers that encode "error" as a string.
type Client struct {
	http     *http.Client
	nodeURL  string
	rpcReqID int64
}

// ConnectEth connects to nodeURL (idempotent).
func (c *Client) ConnectEth(nodeURL string) {
	if c.http == nil {
		c.http = &http.Client{}
	}
	c.nodeURL = nodeURL
}

// CloseConnect is a no-op for HTTP; kept for API parity with scanner.
func (c *Client) CloseConnect() {}

type rawBlock struct {
	Hash         common.Hash        `json:"hash"`
	Transactions []json.RawMessage `json:"transactions"`
	UncleHashes  []common.Hash     `json:"uncles"`
}

func fixTxJSON(data json.RawMessage) (json.RawMessage, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	t, _ := m["type"]
	if t == nil {
		return data, nil
	}
	var isType2 bool
	switch v := t.(type) {
	case string:
		isType2 = (v == "0x2" || v == "2")
	case float64:
		isType2 = (v == 2)
	}
	if !isType2 {
		return data, nil
	}
	if m["maxPriorityFeePerGas"] == nil {
		m["maxPriorityFeePerGas"] = "0x0"
	}
	if m["maxFeePerGas"] == nil {
		m["maxFeePerGas"] = "0x0"
	}
	return json.Marshal(m)
}

// BlockByNumber fetches a full block (same decoding logic as scanner).
func (c *Client) BlockByNumber(number uint64) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var raw json.RawMessage
	err := c.callRPC(ctx, "eth_getBlockByNumber", []interface{}{toBlockNumArg(big.NewInt(int64(number))), true}, &raw)
	if err != nil {
		return nil, err
	}
	if len(bytesTrimSpace(raw)) == 0 || string(bytesTrimSpace(raw)) == "null" {
		return nil, ethereum.NotFound
	}

	var head types.Header
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	var body rawBlock
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}

	if head.UncleHash == types.EmptyUncleHash && len(body.UncleHashes) > 0 {
		return nil, errors.New("server returned non-empty uncle list but block header indicates no uncles")
	}
	if head.TxHash == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, errors.New("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.TxHash != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, errors.New("server returned empty transaction list but block header indicates transactions")
	}

	var uncles []*types.Header
	if len(body.UncleHashes) > 0 {
		uncles = make([]*types.Header, len(body.UncleHashes))
		for i := range uncles {
			var uh *types.Header
			err := c.callRPC(ctx, "eth_getUncleByBlockHashAndIndex", []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))}, &uh)
			if err != nil {
				return nil, err
			}
			if uh == nil {
				return nil, errors.New("got null header for uncle")
			}
			uncles[i] = uh
		}
	}

	txs := make([]*types.Transaction, len(body.Transactions))
	for i, rawTx := range body.Transactions {
		fixed, err := fixTxJSON(rawTx)
		if err != nil {
			return nil, err
		}
		var tx types.Transaction
		if err := tx.UnmarshalJSON(fixed); err != nil {
			return nil, err
		}
		txs[i] = &tx
	}

	return types.NewBlockWithHeader(&head).WithBody(txs, uncles), nil
}

func toBlockNumArg(n *big.Int) string {
	if n == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if n.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(n)
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

// TxByHash returns the transaction and whether it is pending.
func (c *Client) TxByHash(hash common.Hash) (*types.Transaction, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var raw json.RawMessage
	if err := c.callRPC(ctx, "eth_getTransactionByHash", []interface{}{hash}, &raw); err != nil {
		return nil, false, err
	}
	tx, pending, err := decodeTransactionByHashResult(raw)
	if err != nil {
		return nil, false, err
	}
	if tx == nil {
		return nil, false, ethereum.NotFound
	}
	return tx, pending, nil
}

// TxReceipt returns the receipt for hash.
func (c *Client) TxReceipt(hash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var r *types.Receipt
	if err := c.callRPC(ctx, "eth_getTransactionReceipt", []interface{}{hash}, &r); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ethereum.NotFound
	}
	return r, nil
}

// CallContract performs eth_call at blockNumber (nil = latest).
func (c *Client) CallContract(call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var out hexutil.Bytes
	if err := c.callRPC(ctx, "eth_call", []interface{}{toCallArg(call), toBlockNumArg(blockNumber)}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CodeAt returns contract bytecode at address.
func (c *Client) CodeAt(address common.Address, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var out hexutil.Bytes
	if err := c.callRPC(ctx, "eth_getCode", []interface{}{address, toBlockNumArg(blockNumber)}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
