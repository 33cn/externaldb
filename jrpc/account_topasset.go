package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/aggdecode"
	"github.com/33cn/externaldb/escli/querypara"
)

var symbol string

type totalAsset struct {
	Address     string `json:"address"`
	AssetExec   string `json:"asset_exec"`
	AssetSymbol string `json:"asset_symbol"`
	Total       int64  `json:"total"`
	Balance     int64  `json:"balance"`
	Frozen      int64  `json:"frozen"`
}

type cacheAccount struct {
	*sync.RWMutex
	ts      *time.Time
	account []*totalAsset
}

func (c *cacheAccount) needUpdate(now time.Time, secords int64) bool {
	c.RLock()
	defer c.RUnlock()
	if c.ts == nil {
		return true
	}
	return c.ts.Add(time.Duration(secords) * time.Second).Before(now)
}

func (c *cacheAccount) update(ts time.Time, data []*totalAsset) {
	c.Lock()
	defer c.Unlock()
	c.ts = &ts
	c.account = data
}

func (c *cacheAccount) getAccount(number int, size int) []*totalAsset {
	from := (number - 1) * size
	to := from + size
	c.RLock()
	defer c.RUnlock()
	total := len(c.account)
	if from >= total {
		return []*totalAsset{}
	}
	if to > total {
		to = total
	}
	return c.account[from:to]
}

func (c *cacheAccount) getAccountCount() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.account)
}

var btyCache = cacheAccount{ts: nil, RWMutex: new(sync.RWMutex)}

// TopAsset 定制需求 返回资产topN
//  主链浏览器：目前去bityuan链的 coins/ticket 合约中
// 固定参数：指定资产，coins/bty
//         指定在合约，coins+ticket
//         排序 total/frozen/balance
// 实现：由于数据聚合没有分页功能， 有两个选择： 每次重新取; 用缓存。
//      由于目前需求：资产/合约都是特定的，所有缓冲更有效
// 目前接受参数：1. 分页, 默认 1, 10
func (t *Account) TopAsset(req *querypara.Query, out *interface{}) error {
	number, size := 1, 10
	if req != nil && req.Page != nil {
		number, size = req.Page.Number, req.Page.Size
		if number <= 0 || size <= 0 {
			log.Error("TopAsset Bad input", "number", number, "size", size)
			return errors.New(errBadParm)
		}
	}

	if err := t.maybeUpdateCache(); err != nil {
		return err
	}

	*out = btyCache.getAccount(number, size)
	return nil

}

// TopAssetCount 返回对应上面接口的总数
func (t *Account) TopAssetCount(req *querypara.Query, out *interface{}) error {
	if err := t.maybeUpdateCache(); err != nil {
		return err
	}
	*out = btyCache.getAccountCount()
	return nil

}

func (t *Account) maybeUpdateCache() error {
	now := time.Now().UTC()
	// get from cache
	if btyCache.needUpdate(now, 3600) {
		// load data
		data, err := t.loadAsset()
		if err != nil {
			log.Error("TopAsset load", "err", err)
			return err
		}
		btyCache.update(now, data)
	}
	return nil
}

func (t *Account) loadAsset() ([]*totalAsset, error) {
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return nil, err
	}

	q := &querypara.Query{
		Range:    []*querypara.QRange{{Key: "total", GT: 0}},
		MatchOne: []*querypara.QMatch{{Key: "exec", Value: ""}, {Key: "exec", Value: "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"}},
		Not:      []*querypara.QMatch{{Key: "type", Value: "contract"}},
		Filter:   []*querypara.QMatch{{Key: "asset_symbol", Value: t.Symbol}, {Key: "asset_exec", Value: "coins"}},
	}

	a := &querypara.Agg{
		Name:  "total",
		Size:  &querypara.ASize{Size: 100000},
		Order: &querypara.AOrder{Key: "total", Asc: false},
		Term:  &querypara.AAgg{Key: "address"},
		Subs: &querypara.ASub{
			Sum: []*querypara.AAgg{{Name: "frozen", Key: "frozen"}, {Name: "balance", Key: "balance"}, {Name: "total", Key: "total"}},
		},
	}
	s := &querypara.Search{
		Query: q,
		Agg:   a,
	}

	symbol = t.Symbol
	result, err := cli.Agg(account.DBX, account.DBX, s)
	if err != nil {
		log.Error("loadAsset", "err", err.Error())
		return nil, err
	}

	totalAgg, found := result.Term("total")
	if !found {
		log.Error("loadAsset", "load total failed", "not found")
		return nil, fmt.Errorf("not found")
	}
	log.Info("loadAsset", "Buckets.count", len(totalAgg))

	assets := make([]*totalAsset, 0)
	for i, b := range totalAgg {
		if i < 2 {
			//修改string(*b.Aggregations["balance"]
			log.Debug("loadAsset", "address", b.Key.(string), "balance", string(b.Aggregations["balance"]),
				"frozen", string(b.Aggregations["frozen"]), "total", string(b.Aggregations["total"]))
		}
		asset, err := getAccountAsset(b)
		if err != nil {
			log.Error("loadAsset getAccountAsset", "err", err, "index", i)
			return nil, err
		}
		asset.AssetExec, asset.AssetSymbol = "coins", t.Symbol
		assets = append(assets, asset)
	}

	return assets, nil
}

type sumValue struct {
	Value float64 `json:"value"`
}

// ErrDecode ErrDecode
var ErrDecode = fmt.Errorf("Decode failed")

func getAccountAsset(item *aggdecode.AggregationBucketKeyItem) (*totalAsset, error) {
	var asset totalAsset
	address, ok := item.Key.(string)
	if !ok {
		return nil, ErrDecode
	}
	asset.Address = address

	var total, frozen, balance sumValue

	err := json.Unmarshal([]byte(item.Aggregations["total"]), &total)
	if err != nil {
		return nil, err
	}
	asset.Total = int64(total.Value)
	err = json.Unmarshal([]byte(item.Aggregations["frozen"]), &frozen)
	if err != nil {
		return nil, err
	}
	asset.Frozen = int64(frozen.Value)
	err = json.Unmarshal([]byte(item.Aggregations["balance"]), &balance)
	if err != nil {
		return nil, err
	}
	asset.Balance = int64(balance.Value)
	return &asset, nil
}
