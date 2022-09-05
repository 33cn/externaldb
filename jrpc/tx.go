package main

import (
	"encoding/json"

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

//TxCount 交易数
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

//TxCounts 获取多组查询条件的交易数
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

//TxList 全部交易列表接口
func (t *Tx) TxList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.New(ErrBadParm)
	}

	r, err := t.search(transaction.TransactionX, transaction.TransactionX, q, decodeTransaction)
	if err != nil {
		return err
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

//BlockList 全部区块列表接口
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
