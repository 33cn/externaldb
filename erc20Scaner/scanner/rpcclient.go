package main

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"math/big"
	"time"
)

type Client struct {
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
		cli, err := ethclient.DialContext(ctx, nodeURL)
		if err != nil {
			panic(err)
		}
		c.cli = cli
		return cli
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

// BlockByNumber 根据blockNum获取区块信息
func (c *Client) BlockByNumber(number uint64) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return c.cli.BlockByNumber(ctx, big.NewInt(int64(number)))
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
