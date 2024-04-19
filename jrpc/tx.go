package main

import (
	"encoding/json"
	"strings"

	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/pkg/errors"
)

var blockSupportKey = map[string]bool{
	"height":     true,
	"block_time": true,
	"block_hash": true,
}

// Tx Tx
type Tx struct {
	*DBRead
	ChainGrpc string
	Symbol    string
}

// BlockInfo BlockInfo
type BlockInfo struct {
	*transaction.Block
}

// ExecerResponse 统计交易数
type ExecerResponse struct {
	Intro   string `json:"introduction"`
	Name    string `json:"name"`
	TxCount int64  `json:"tx_count"`
}

// TxCount 交易数
func (t *Tx) TxCount(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.New(ErrBadParm)
	}

	r, err := t.count(transaction.TransactionX, transaction.TransactionX, q)
	if err != nil {
		return err
	}
	*out = r
	return nil
}

// TxCounts 获取多组查询条件的交易数
func (t *Tx) TxCounts(reqs []*querypara.Query, out *interface{}) error {
	assets := make([]int64, 0)
	for _, req := range reqs {
		r, err := t.count(transaction.TransactionX, transaction.TransactionX, req)
		if err != nil {
			return err
		}
		assets = append(assets, r)
	}

	*out = assets
	return nil
}

// TxList 全部交易列表接口
func (t *Tx) TxList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.New(ErrBadParm)
	}

	r, err := t.search(transaction.TransactionX, transaction.TransactionX, q, decodeTransaction)
	if err != nil {
		return err
	}

	// 找出amount
	evm := Evm{DBRead: t.DBRead, ChainGrpc: t.ChainGrpc, Symbol: t.Symbol}
	for _, tx1 := range r {
		tx2, ok := tx1.(*transaction.Transaction)
		if !ok {
			continue
		}
		if tx2.Amount != 0 {
			continue
		}
		if !isEvmTx(string(tx2.Execer)) {
			continue
		}
		host := t.ChainGrpc
		detail, err := getTxDetailFromChain33(host, tx2.Hash)
		if err != nil {
			log.Error("TxList", "getTxDetailFromChain33", err.Error())
			*out = err.Error()
			continue
		}

		parsed := parseEvmTx(detail, evm.get, t.Symbol)
		// 初始asset 可能有默认的空值
		log.Info("TxList", "getTxDetailFromChain33", parsed)
		tx2.Amount = int64(parsed.Amount)
		log.Info("TxList", "asset 2", tx2.Assets)
		if len(tx2.Assets) == 1 && tx2.Assets[1].Amount == 0 {
			tx2.Assets[1].Amount = tx2.Amount
			tx2.Assets[1].Exec = tx2.Execer
			if strings.HasSuffix(string(tx2.Assets[1].Exec), ".evm") {
				tx2.Assets[1].Symbol = "Para"
			} else {
				tx2.Assets[1].Symbol = "BTY"
			}
			log.Info("TxList", "asset 1", tx2.Assets)
		}

		if tx2.Assets == nil || len(tx2.Assets) == 0 {
			var asset transaction.Asset
			asset.Amount = tx2.Amount
			asset.Exec = tx2.Execer
			if strings.HasSuffix(string(asset.Exec), ".evm") {
				asset.Symbol = "Para"
			} else {
				asset.Symbol = "BTY"
			}
			tx2.Assets = append(tx2.Assets, asset)
		}
		log.Info("TxList", "asset 0", tx2.Assets)
	}
	*out = r
	return nil
}

// ExecerInfo 统计交易数
func (t *Tx) ExecerInfo(name string, out *interface{}) error {
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	q := &querypara.Query{}
	q.Match = append(q.Match, &querypara.QMatch{Key: "execer", Value: name})
	r, err := cli.Count(transaction.TransactionX, transaction.TransactionX, q)
	if err != nil {
		return err
	}

	*out = &ExecerResponse{Intro: name, Name: name, TxCount: r}
	return nil
}

// BlockList 全部区块列表接口
func (t *Tx) BlockList(q *querypara.Query, out *interface{}) error {
	err := checkQuery(q, blockSupportKey)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	initQuery(q)
	q.Match = append(q.Match, &querypara.QMatch{Key: "index", Value: 0})
	rs, err := cli.Search(transaction.TransactionX, transaction.TransactionX, q, decodeTransaction)
	if err != nil {
		return err
	}
	var blocks []*BlockInfo
	for _, r := range rs {
		tx := r.(*transaction.Transaction)
		blocks = append(blocks, &BlockInfo{Block: tx.Block})
	}
	*out = blocks
	return nil
}

func decodeTransaction(x *json.RawMessage) (interface{}, error) {
	t := transaction.Transaction{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}
