package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	rpcCli  *rpc.Client
	cli     *ethclient.Client
	nodeURL string
}

// ConnectEth 连接节点
func (c *Client) ConnectEth(nodeURL string) *ethclient.Client {
	if c.cli == nil {
		if nodeURL == "" {
			panic("node URL is required")
		}
		c.nodeURL = nodeURL
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		rpcCli, err := rpc.DialContext(ctx, nodeURL)
		if err != nil {
			panic(err)
		}
		c.rpcCli = rpcCli
		c.cli = ethclient.NewClient(rpcCli)
		return c.cli
	}
	return c.cli
}

func (c *Client) CloseConnect() {
	if c.cli == nil {
		return
	}

	c.cli.Close()
}

// BlockNum 获取区块高度
func (c *Client) BlockNum() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.BlockNumber(ctx)
}

// rawBlock 用于接收 eth_getBlockByNumber 的原始 JSON，交易先按 RawMessage 拿入再按需补全 EIP-1559 字段后解码
type rawBlock struct {
	Hash         common.Hash        `json:"hash"`
	Transactions []json.RawMessage `json:"transactions"`
	UncleHashes  []common.Hash     `json:"uncles"`
}

// fixTxJSON 对 type=2 的交易若缺少 maxPriorityFeePerGas/maxFeePerGas 则补 0x0，避免 go-ethereum 解码报错
func fixTxJSON(data json.RawMessage) (json.RawMessage, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	t, _ := m["type"]
	if t == nil {
		return data, nil
	}
	// type 可能是 "0x2" 或 2
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

// BlockByNumber 根据 blockNum 获取区块信息。对返回 type=2 但缺少 maxPriorityFeePerGas/maxFeePerGas 的节点做兼容。
func (c *Client) BlockByNumber(number uint64) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if c.rpcCli == nil {
		return c.cli.BlockByNumber(ctx, big.NewInt(int64(number)))
	}

	var raw json.RawMessage
	err := c.rpcCli.CallContext(ctx, &raw, "eth_getBlockByNumber", toBlockNumArg(big.NewInt(int64(number))), true)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
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
		reqs := make([]rpc.BatchElem, len(body.UncleHashes))
		for i := range reqs {
			reqs[i] = rpc.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))},
				Result: &uncles[i],
			}
		}
		if err := c.rpcCli.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
			if uncles[i] == nil {
				return nil, errors.New("got null header for uncle")
			}
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
	return hexutil.EncodeBig(n)
}

// TxByHash 根据哈希获取交易
func (c *Client) TxByHash(hash common.Hash) (*types.Transaction, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.TransactionByHash(ctx, hash)

}

// TxReceipt 根据哈希获取交易日志
func (c *Client) TxReceipt(hash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.TransactionReceipt(ctx, hash)
}

// CallContract 合约调用
func (c *Client) CallContract(call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.CallContract(ctx, call, blockNumber)
}

// CodeAt 获取合约代码
func (c *Client) CodeAt(address common.Address, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.CodeAt(ctx, address, blockNumber)
}
