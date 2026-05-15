//go:build integration

// 链上连通性 / 扫块冒烟测试，依赖外网节点；默认不参与单元测试。
// 运行：go test -tags=integration -v -timeout=120s ./scanner/

package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func testNodeURL() string {
	return "https://mainnet.bityuan.com/eth"
}

func TestClient_BlockNum(t *testing.T) {
	cli := new(Client)
	cli.ConnectEth(testNodeURL())
	t.Cleanup(func() { cli.CloseConnect() })

	num, err := cli.BlockNum()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("current block number:", num)
}

func TestClient_BlockByNumber(t *testing.T) {
	cli := new(Client)
	cli.ConnectEth(testNodeURL())
	t.Cleanup(func() { cli.CloseConnect() })

	const blockNum = uint64(34562329)
	block, err := cli.BlockByNumber(blockNum)
	if err != nil {
		t.Fatal(err)
	}
	txs := block.Transactions()
	t.Log("tx.num:", txs.Len())
	for _, tx := range txs {
		if tx.To() == nil {
			t.Log("tx.Hash:", tx.Hash(), "tx.to:", tx.To(), "tx.value:", tx.Value(), "tx.data", common.Bytes2Hex(tx.Data()))
		}
	}
}

func TestParseBlock_Integration(t *testing.T) {
	p := new(Process)
	p.nodeURL = testNodeURL()
	p.Init()
	t.Cleanup(func() { _ = p.Close() })

	const blockNum = uint64(34562329)
	block, err := p.cli.BlockByNumber(blockNum)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.ParaseBlock(block); err != nil {
		t.Fatal(err)
	}
}
