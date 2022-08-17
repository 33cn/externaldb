package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/address"
	"github.com/33cn/externaldb/db/contract"
	"github.com/33cn/externaldb/db/evm"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
)

// EVM EVM
type EVM struct {
	*DBRead
}

func decodeContract(x *json.RawMessage) (interface{}, error) {
	c := contract.Contract{}
	err := json.Unmarshal([]byte(*x), &c)
	return &c, err
}

func decodeToken(x *json.RawMessage) (interface{}, error) {
	token := evm.Token{}
	err := json.Unmarshal([]byte(*x), &token)
	return &token, err
}

func decodeTransfer(x *json.RawMessage) (interface{}, error) {
	transfer := evm.Transfer{}
	err := json.Unmarshal([]byte(*x), &transfer)
	return &transfer, err
}

// ContractList 查询合约列表
func (t *EVM) ContractList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(errors.New(ErrBadParm), "empty queryPara input")
	}

	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	r, err := t.search(contract.TableName, contract.TableName, q, decodeContract)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// ContractCount 查询合约数量
func (t *EVM) ContractCount(q *querypara.Query, out *interface{}) error {
	var err error
	*out, err = t.count(contract.TableName, contract.TableName, q)
	return err
}

// GetContract 查询合约详情
func (t *EVM) GetContract(address string, out *interface{}) error {
	if address == "" {
		return errors.Wrapf(errors.New(ErrBadParm), "empty input")
	}

	id := "contract-" + address
	r, err := t.get(contract.TableName, contract.TableName, id)
	if err != nil {
		return err
	}
	res, err := decodeContract(r)
	*out = res
	return nil
}

//// AccCount 查询NFT账户数量
//func (t *EVM) AccCount(q *queryPara.Query, out *interface{}) error {
//	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
//	if err != nil {
//		return err
//	}
//
//	a := &querypara.Agg{
//		Name:        "count",
//		Cardinality: "owner_addr",
//	}
//	s := &querypara.Search{
//		Agg: a,
//	}
//
//	symbol = t.Symbol
//	result, err := cli.Agg(db.AccountX, db.AccountX, s)
//	if err != nil {
//		log.Error("Count", "err", err.Error())
//		return err
//	}
//	*out = result.Aggregations["count"]
//
//	return err
//}

// TokenList 查询地址对应的代币列表
func (t *EVM) TokenList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(errors.New(ErrBadParm), "empty queryPara input")
	}
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	r, err := t.search(evm.EVMTokenX, evm.EVMTokenX, q, decodeToken)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// TokenListAgg 查询地址对应的代币列表,并统计
func (t *EVM) TokenListAgg(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(errors.New(ErrBadParm), "empty queryPara input")
	}
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		log.Error("TokenListAgg NewESShortConnect", "err", err.Error())
		return err
	}
	q.Range = append(q.Range, &querypara.QRange{
		Key: "amount",
		GT:  0,
	})

	a := &querypara.Agg{
		Name: "count",
		Term: &querypara.AAgg{Key: "contract_addr"},
		Subs: &querypara.ASub{
			Sum: []*querypara.AAgg{
				{Name: "amount", Key: "amount"},
			},
		},
		Size: &querypara.ASize{
			Size: q.Page.Size,
		},
	}
	s := &querypara.Search{
		Query: q,
		Agg:   a,
	}

	resp, err := cli.Agg(evm.EVMTokenX, evm.EVMTokenX, s)
	if err != nil {
		log.Error("TokenListAgg cli.Agg", "err", err.Error())
		return err
	}

	var aggRes AggResult
	err = json.Unmarshal(resp.Aggregations["count"], &aggRes)
	if err != nil {
		log.Error("TokenListAgg json.Unmarshal", "err", err.Error())
		return err
	}
	*out = &aggRes

	// add contract info
	ctQy := &querypara.Query{Fetch: &querypara.QFetch{
		FetchSource: true,
		Keys:        []string{"contract_address", "creator", "contract_bin_hash", "contract_type", "tx_count", "publish_count", "name", "symbol", "uri"},
	}, Page: &querypara.QPage{
		Number: 1,
		Size:   q.Page.Size,
	}}

	for _, bucket := range aggRes.Buckets {
		ctQy.MatchOne = append(ctQy.MatchOne, &querypara.QMatch{
			Key:   "contract_address",
			Value: bucket.Key,
		})
	}
	var resp2 interface{}
	var ctMap = make(map[string]*contract.Contract)
	err = t.ContractList(ctQy, &resp2)
	if err != nil {
		log.Error("TokenListAgg ContractList", "err", err.Error())
	}
	ctList, ok := resp2.([]interface{})
	if !ok {
		return err
	}
	for _, ct := range ctList {
		ct, ok := ct.(*contract.Contract)
		if !ok {
			log.Error("TokenListAgg ct.(*contract.Contract) failed")
			continue
		}
		ctMap[ct.Address] = ct
	}
	for i := range aggRes.Buckets {
		aggRes.Buckets[i].Contract = ctMap[aggRes.Buckets[i].Key]
	}
	return err
}

type AggResult struct {
	DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int `json:"sum_other_doc_count"`
	Buckets                 []struct {
		Key      string `json:"key"`
		DocCount int    `json:"doc_count"`
		Amount   struct {
			Value float64 `json:"value"`
		} `json:"amount"`
		*contract.Contract
	} `json:"buckets"`
}

// AddrCount count address
func (t *EVM) AddrCount(q *querypara.Query, out *interface{}) error {
	var err error
	*out, err = t.count(address.TableName, address.TableName, q)
	return err
}

// IsContract address
func (t *EVM) IsContract(address string, out *interface{}) error {
	if address == "" {
		return errors.Wrapf(errors.New(ErrBadParm), "empty input")
	}
	id := "contract-" + address
	_, err := t.get(contract.TableName, contract.TableName, id)
	if err != nil {
		if err == db.ErrDBNotFound {
			*out = false
		} else {
			return err
		}
	} else {
		*out = true
	}
	return nil
}

// TransferList 转账列表
func (t *EVM) TransferList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(errors.New(ErrBadParm), "empty queryPara input")
	}

	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	r, err := t.search(evm.EVMTransferX, evm.EVMTransferX, q, decodeTransfer)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// TransferCount count transfer
func (t *EVM) TransferCount(q *querypara.Query, out *interface{}) error {
	var err error
	*out, err = t.count(evm.EVMTransferX, evm.EVMTransferX, q)
	return err
}
