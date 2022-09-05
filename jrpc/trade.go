package main

import (
	"encoding/json"

	"github.com/33cn/externaldb/db/trade"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/pkg/errors"
)

// TradeQuery TradeQuery
type TradeQuery struct {
	*querypara.Query
}

// Trade Trade
type Trade struct {
	*DBRead
}

const (
	errBadParm = "Bad Parm"
)

var tradeSupportKey = map[string]bool{
	"asset_exec":      true,
	"is_sell":         true,
	"is_finished":     true,
	"asset_symbol":    true,
	"boardlot_price":  true,
	"status":          true,
	"traded_boardlot": true,
	"owner":           true,
	"send":            true,
	"height":          true,
	"block_hash":      true,
	"total_boardlot":  true,
	"success":         true,
	"height_index":    true,
	"price_symbol":    true,
	"price_exec":      true,
}

func checkQuery(q *querypara.Query, keys map[string]bool) error {
	if q == nil {
		return nil
	}

	if q.Page != nil {
		if q.Page.Size <= 0 || q.Page.Number <= 0 {
			return errors.Wrapf(errors.New(errBadParm), "page: n%d s%d", q.Page.Number, q.Page.Size)
		}
	}
	if q.Match != nil {
		for _, m := range q.Match {
			if _, ok := keys[m.Key]; !ok {
				return errors.Wrapf(errors.New(errBadParm), "match key: %s", m.Key)
			}
		}
	}
	if q.Range != nil {
		for _, m := range q.Range {
			if _, ok := keys[m.Key]; !ok {
				return errors.Wrapf(errors.New(errBadParm), "range key: %s", m.Key)
			}
			// TODO check range
		}
	}
	if q.Sort != nil {
		for _, m := range q.Sort {
			if _, ok := keys[m.Key]; !ok {
				return errors.Wrapf(errors.New(errBadParm), "sort key: %s", m.Key)
			}
		}
	}
	return nil
}

// ListOrder ListOrder
func (t *Trade) ListOrder(q *TradeQuery, out *interface{}) error {
	err := checkQuery(q.Query, tradeSupportKey)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	r, err := cli.Search(trade.TradeDBX, trade.TradeDBX, q.Query, decodeTradeOrder)
	if err != nil || r == nil {
		return err
	}
	*out = r //&TradeOrder{Order: r}
	return nil
}

// ListTx ListTx
func (t *Trade) ListTx(q *TradeQuery, out *interface{}) error {
	err := checkQuery(q.Query, tradeSupportKey)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	r, err := cli.Search(trade.TradeTxDBX, trade.TradeTxDBX, q.Query, decodeTradeTx)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// ListAsset ListAsset
func (t *Trade) ListAsset(q interface{}, out *interface{}) error {
	q1 := &querypara.Query{
		Page: &querypara.QPage{
			Number: 1,
			Size:   10000,
		},
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	r, err := cli.Search(trade.TradeAssetDBX, trade.TradeAssetDBX, q1, decodeTradeAsset)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// LastPrice LastPrice
type LastPrice struct {
	Height            int64  `json:"height"`
	Index             int64  `json:"index"`
	AssetExec         string `json:"asset_exec"`
	AssetSymbil       string `json:"asset_symbol"`
	PricePerBoardlot  int64  `json:"boardlot_price"`
	AmountPerBoardlot int64  `json:"boardlot_amount"`
	PriceExec         string `json:"price_exec"`
	PriceSymbol       string `json:"price_symbol"`
}

// ListLastPrice ListLastPrice
func (t *Trade) ListLastPrice(q []*trade.Asset, out *interface{}) error {
	checkedAssets := make([]*trade.Asset, 0)
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	if len(q) == 0 {
		q1 := &querypara.Query{
			Page: &querypara.QPage{
				Number: 1,
				Size:   10000,
			},
		}

		assets, err := cli.Search(trade.TradeAssetDBX, trade.TradeAssetDBX, q1, decodeTradeAsset)
		if err != nil || assets == nil {
			return err
		}
		for _, asset := range assets {
			a, ok := asset.(*trade.Asset)
			if !ok {
				continue
			}
			checkedAssets = append(checkedAssets, a)
		}

	} else {
		checkedAssets = q
	}

	prices := make([]*LastPrice, 0)
	for _, asset := range checkedAssets {
		if asset.PriceExec == "" {
			asset.PriceSymbol = t.Symbol
			asset.PriceExec = "coins"
		}
		q2 := &querypara.Query{
			Page: &querypara.QPage{
				Number: 1,
				Size:   1,
			},
			Sort: []*querypara.QSort{
				{
					Key:       "height_index",
					Ascending: false,
				},
			},
			Match: []*querypara.QMatch{
				{
					Key:   "asset_exec",
					Value: asset.AssetExec,
				},
				{
					Key:   "asset_symbol",
					Value: asset.AssetSymbol,
				},
				{
					Key:   "status",
					Value: "done",
				},
				{
					Key:   "price_exec",
					Value: asset.PriceExec,
				},
				{
					Key:   "price_symbol",
					Value: asset.PriceSymbol,
				},
			},
		}
		last, err := cli.Search(trade.TradeDBX, trade.TradeDBX, q2, decodeTradeOrder)
		if err != nil || last == nil {
			return err
		}
		if len(last) == 0 {
			continue
		}
		lastPrice, ok := last[0].(*trade.Order)
		if !ok {
			continue
		}
		price := &LastPrice{
			Height:            lastPrice.Height,
			Index:             lastPrice.Index,
			AssetExec:         asset.AssetExec,
			AssetSymbil:       asset.AssetSymbol,
			PricePerBoardlot:  lastPrice.PricePerBoardlot,
			AmountPerBoardlot: lastPrice.AmountPerBoardlot,
			PriceExec:         asset.AssetExec,
			PriceSymbol:       asset.PriceSymbol,
		}
		prices = append(prices, price)
	}
	*out = prices

	return nil
}

func decodeTradeOrder(x *json.RawMessage) (interface{}, error) {
	t := trade.Order{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}

func decodeTradeTx(x *json.RawMessage) (interface{}, error) {
	t := trade.Tx{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}

func decodeTradeAsset(x *json.RawMessage) (interface{}, error) {
	t := trade.Asset{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}
