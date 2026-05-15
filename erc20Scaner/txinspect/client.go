package txinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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
	// After eth_getTransactionByBlockNumberAndIndex returns -32601 once, skip it for later indices on this client.
	txByNumberIdxUnsupported atomic.Bool
	// After eth_getTransactionByBlockHashAndIndex returns -32601 once, skip it for later indices on this client.
	txByHashIdxUnsupported atomic.Bool
	// After eth_getBlockReceipts returns -32601 once, skip it for later indices on this client.
	blockReceiptsUnsupported atomic.Bool
	receiptsMu               sync.Mutex
	receiptsBlockHash        common.Hash
	receiptsCache            []*types.Receipt
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
	Hash         common.Hash       `json:"hash"`
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

// isLikelyFullTxBlockFailure detects RPC nodes that panic on eth_getBlockByNumber(..., true)
// but still serve fullTx=false (tx hash list) + eth_getTransactionByHash.
func isLikelyFullTxBlockFailure(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "-32603") ||
		strings.Contains(s, "method handler crashed") ||
		strings.Contains(s, "could not coalesce error")
}

func txRawLooksLikeHashString(raw json.RawMessage) bool {
	b := bytesTrimSpace(raw)
	return len(b) >= 2 && b[0] == '"'
}

// parseTxHashFromBlockJSON decodes a tx hash from eth_getBlockByNumber(..., false) list entries.
// Some chains omit the "0x" prefix on hex strings; common.Hash JSON does not accept that.
func parseTxHashFromBlockJSON(raw json.RawMessage) (common.Hash, error) {
	var s string
	if err := json.Unmarshal(bytesTrimSpace(raw), &s); err != nil {
		return common.Hash{}, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return common.Hash{}, fmt.Errorf("empty hash string")
	}
	low := strings.ToLower(s)
	if !strings.HasPrefix(low, "0x") {
		// bare 64-hex is common on partial-EVM JSON-RPC
		if len(s) == 64 {
			b, err := hexutil.Decode("0x" + s)
			if err != nil {
				return common.Hash{}, err
			}
			if len(b) != 32 {
				return common.Hash{}, fmt.Errorf("hash length %d", len(b))
			}
			return common.BytesToHash(b), nil
		}
		s = "0x" + s
	}
	b, err := hexutil.Decode(s)
	if err != nil {
		return common.Hash{}, err
	}
	if len(b) != 32 {
		return common.Hash{}, fmt.Errorf("hash length %d", len(b))
	}
	return common.BytesToHash(b), nil
}

// BlockByNumber fetches a block body. The second return lists known tx hashes per index
// (useful when the slot has no *Transaction). Non-ETH-compatible nodes may omit "0x" on
// hash strings or omit full tx bodies; those slots are nil in the transaction slice.
func (c *Client) BlockByNumber(number uint64) (*types.Block, []common.Hash, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	numArg := toBlockNumArg(big.NewInt(int64(number)))
	var raw json.RawMessage
	err := c.callRPC(ctx, "eth_getBlockByNumber", []interface{}{numArg, true}, &raw)
	if err != nil {
		if !isLikelyFullTxBlockFailure(err) {
			return nil, nil, fmt.Errorf("eth_getBlockByNumber height=%d: %w", number, err)
		}
		if err2 := c.callRPC(ctx, "eth_getBlockByNumber", []interface{}{numArg, false}, &raw); err2 != nil {
			return nil, nil, fmt.Errorf("eth_getBlockByNumber height=%d fullTx=true failed (%v); fullTx=false: %w", number, err, err2)
		}
	}
	if len(bytesTrimSpace(raw)) == 0 || string(bytesTrimSpace(raw)) == "null" {
		return nil, nil, ethereum.NotFound
	}
	return c.decodeBlockResult(ctx, number, raw)
}

func (c *Client) decodeBlockResult(ctx context.Context, number uint64, raw json.RawMessage) (*types.Block, []common.Hash, error) {
	var head types.Header
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, nil, fmt.Errorf("decode block header height=%d: %w | result_prefix=%s", number, err, truncateBytes([]byte(raw), 800))
	}
	var body rawBlock
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, nil, fmt.Errorf("decode block body height=%d: %w | result_prefix=%s", number, err, truncateBytes([]byte(raw), 800))
	}

	if head.UncleHash == types.EmptyUncleHash && len(body.UncleHashes) > 0 {
		return nil, nil, errors.New("server returned non-empty uncle list but block header indicates no uncles")
	}
	if head.TxHash == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, nil, errors.New("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.TxHash != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, nil, errors.New("server returned empty transaction list but block header indicates transactions")
	}

	var uncles []*types.Header
	if len(body.UncleHashes) > 0 {
		uncles = make([]*types.Header, len(body.UncleHashes))
		for i := range uncles {
			var uh *types.Header
			err := c.callRPC(ctx, "eth_getUncleByBlockHashAndIndex", []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))}, &uh)
			if err != nil {
				return nil, nil, fmt.Errorf("eth_getUncleByBlockHashAndIndex height=%d uncle_index=%d block_hash=%s: %w", number, i, body.Hash.Hex(), err)
			}
			if uh == nil {
				return nil, nil, errors.New("got null header for uncle")
			}
			uncles[i] = uh
		}
	}

	txs := make([]*types.Transaction, len(body.Transactions))
	parallel := make([]common.Hash, len(body.Transactions))

	for i, rawTx := range body.Transactions {
		if txRawLooksLikeHashString(rawTx) {
			th, err := parseTxHashFromBlockJSON(rawTx)
			if err != nil {
				txs[i] = nil
				continue
			}
			parallel[i] = th
			tx, _, err := c.TxByHash(th)
			if err != nil || tx == nil {
				txs[i] = nil
				continue
			}
			txs[i] = tx
			parallel[i] = tx.Hash()
			continue
		}
		fixed, err := fixTxJSON(rawTx)
		if err != nil {
			txs[i] = nil
			continue
		}
		var tx types.Transaction
		if err := tx.UnmarshalJSON(fixed); err != nil {
			txs[i] = nil
			continue
		}
		txs[i] = &tx
		parallel[i] = tx.Hash()
	}

	return types.NewBlockWithHeader(&head).WithBody(txs, uncles), parallel, nil
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

func isRPCMethodUnavailable(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "-32601") ||
		strings.Contains(s, "method not found") ||
		strings.Contains(s, "does not exist") ||
		strings.Contains(s, "not available")
}

func findReceiptByTxIndex(recs []*types.Receipt, txIndex uint64) *types.Receipt {
	for _, r := range recs {
		if r == nil {
			continue
		}
		if uint64(r.TransactionIndex) == txIndex {
			return r
		}
	}
	return nil
}

// TransactionAtCanonicalIndex returns the tx at txIndex inside block.
// Order: non-nil body from getBlock; eth_getTransactionByBlockNumberAndIndex;
// eth_getTransactionByBlockHashAndIndex; eth_getBlockReceipts + eth_getTransactionByHash (receipt.transactionIndex).
func (c *Client) TransactionAtCanonicalIndex(block *types.Block, txIndex uint64) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ti := int(txIndex)
	eth := block.Transactions()
	if ti >= 0 && ti < eth.Len() {
		if tx := eth[ti]; tx != nil {
			return tx, nil
		}
	}

	num := hexutil.EncodeUint64(block.NumberU64())
	idx := hexutil.EncodeUint64(txIndex)

	tryByHash := func() (*types.Transaction, error) {
		bh := block.Hash()
		if bh == (common.Hash{}) {
			return nil, fmt.Errorf("block hash is zero, cannot call eth_getTransactionByBlockHashAndIndex")
		}
		var raw json.RawMessage
		if err := c.callRPC(ctx, "eth_getTransactionByBlockHashAndIndex", []interface{}{bh, idx}, &raw); err != nil {
			return nil, err
		}
		tx, _, err := decodeTransactionByHashResult(raw)
		if err != nil {
			return nil, err
		}
		if tx == nil {
			return nil, ethereum.NotFound
		}
		return tx, nil
	}

	if !c.txByNumberIdxUnsupported.Load() {
		var raw json.RawMessage
		if err := c.callRPC(ctx, "eth_getTransactionByBlockNumberAndIndex", []interface{}{num, idx}, &raw); err != nil {
			if isRPCMethodUnavailable(err) {
				c.txByNumberIdxUnsupported.Store(true)
			} else {
				return nil, err
			}
		} else {
			tx, _, err := decodeTransactionByHashResult(raw)
			if err != nil {
				return nil, err
			}
			if tx == nil {
				return nil, ethereum.NotFound
			}
			return tx, nil
		}
	}

	if !c.txByHashIdxUnsupported.Load() {
		tx, err := tryByHash()
		if err == nil {
			return tx, nil
		}
		if isRPCMethodUnavailable(err) {
			c.txByHashIdxUnsupported.Store(true)
		} else {
			return nil, err
		}
	}

	return c.txFromBlockReceipts(ctx, block, txIndex)
}

func (c *Client) txFromBlockReceipts(ctx context.Context, block *types.Block, txIndex uint64) (*types.Transaction, error) {
	if c.blockReceiptsUnsupported.Load() {
		return nil, fmt.Errorf("eth_getBlockReceipts not implemented on this RPC (also need getBlock tx count = chain33 tx count, or use a fuller JSON-RPC endpoint)")
	}
	bh := block.Hash()
	if bh == (common.Hash{}) {
		return nil, fmt.Errorf("block hash is zero, cannot call eth_getBlockReceipts")
	}
	c.receiptsMu.Lock()
	if c.receiptsCache != nil && c.receiptsBlockHash == bh {
		recs := c.receiptsCache
		c.receiptsMu.Unlock()
		return c.txFromReceiptList(ctx, recs, txIndex)
	}
	c.receiptsMu.Unlock()

	var raw json.RawMessage
	if err := c.callRPC(ctx, "eth_getBlockReceipts", []interface{}{bh}, &raw); err != nil {
		if isRPCMethodUnavailable(err) {
			c.blockReceiptsUnsupported.Store(true)
			return nil, fmt.Errorf("eth_getBlockReceipts not implemented on this RPC (-32601 is method missing, not \"no EVM txs\"; empty block would return []): %w", err)
		}
		return nil, err
	}
	var recs []*types.Receipt
	if err := json.Unmarshal(raw, &recs); err != nil {
		return nil, fmt.Errorf("decode eth_getBlockReceipts: %w", err)
	}
	c.receiptsMu.Lock()
	c.receiptsBlockHash = bh
	c.receiptsCache = recs
	c.receiptsMu.Unlock()
	return c.txFromReceiptList(ctx, recs, txIndex)
}

func (c *Client) txFromReceiptList(ctx context.Context, recs []*types.Receipt, txIndex uint64) (*types.Transaction, error) {
	r := findReceiptByTxIndex(recs, txIndex)
	if r == nil {
		return nil, ethereum.NotFound
	}
	_ = ctx
	tx, _, err := c.TxByHash(r.TxHash)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, ethereum.NotFound
	}
	return tx, nil
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
