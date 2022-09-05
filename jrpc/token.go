package main

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"

	"fmt"

	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/token"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/pkg/errors"
)

// TokenCache cache
var TokenCache tokenCache
var txSupportKey = map[string]bool{
	"height":               true,
	"block_time":           true,
	"block_hash":           true,
	"success":              true,
	"hash":                 true,
	"from":                 true,
	"to":                   true,
	"execer":               true,
	"amount":               true,
	"fee":                  true,
	"action_name":          true,
	"group_count":          true,
	"is_withdraw":          true,
	"options.symbol":       true,
	"options.to":           true,
	"options.exec_name":    true,
	"options.amount":       true,
	"options.name":         true,
	"options.introduction": true,
	"options.total":        true,
	"options.price":        true,
	"options.owner":        true,
	"options.category":     true,
	"options.note":         true,
}

// TokenAddrsInfoQuery TokenAddrsInfoQuery
type TokenAddrsInfoQuery struct {
	*querypara.Query
	Tokens []string `json:"tokens"`
}

// Token Token
type Token struct {
	*DBRead
}

//TokenAddrsInfo 代币持有人列表
type TokenAddrsInfo struct {
	Symbol string         `json:"symbol"`
	Total  int64          `json:"total"`
	Addrs  []*AddrBalance `json:"addrs"`
}

// AddrBalance AddrBalance
type AddrBalance struct {
	Addr    string `json:"addr"`
	Balance int64  `json:"balance"`
}

//缓存数据
type tokenCache struct {
	//各代币总发行量
	TokenList map[string]int64
	*sync.RWMutex
}

func (t *tokenCache) Get(key string) int64 {
	t.RLock()
	defer t.RUnlock()
	return t.TokenList[key]
}

func (t *tokenCache) Set(key string, value int64) {
	t.Lock()
	defer t.Unlock()
	t.TokenList[key] = value
}

// InitCache 初始化缓存
func InitCache(t *Token) {
	TokenCache.TokenList = make(map[string]int64)
	TokenCache.RWMutex = new(sync.RWMutex)

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return
	}

	err = initTokenList(cli)
	if err != nil {
		return
	}
}

//更新token列表
func initTokenList(cli escli.ESClient) error {
	q := querypara.Query{
		Page: &querypara.QPage{Size: 10000, Number: 1},
		Match: []*querypara.QMatch{
			{Key: "status", Value: 1},
		},
	}
	r, err := cli.Search(token.TokenInfoDB, token.TokenInfoDB, &q, decodeTokenInfo)
	if err != nil || r == nil {
		log.Error("init symbol", "queryPara token info err", err)
		return err
	}

	for _, v := range r {
		info := v.(*token.Token)
		TokenCache.Set(info.Symbol, info.Amount)
	}

	return nil
}

func isPara(title string) bool {
	return strings.Count(title, ".") == 3 && strings.HasPrefix(title, types.ParaKeyX)
}

func argsExecer(title, e string) string {
	if isPara(title) {
		e = title + e
	}
	return e
}

// TxList TxList
func (t *Token) TxList(q *querypara.Query, out *interface{}) error {
	err := checkQuery(q, txSupportKey)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	initQuery(q)
	q.Match = append(q.Match, &querypara.QMatch{Key: "execer", Value: argsExecer(t.Title, "token")})

	r, err := cli.Search(transaction.TransactionX, transaction.TransactionX, q, decodeTransaction)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

//TxCount 代币交易数
func (t *Token) TxCount(q *querypara.Query, out *interface{}) error {
	err := checkQuery(q, txSupportKey)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	initQuery(q)
	q.Match = append(q.Match, &querypara.QMatch{Key: "execer", Value: argsExecer(t.Title, "token")})

	r, err := cli.Count(transaction.TransactionX, transaction.TransactionX, q)
	if err != nil {
		return err
	}
	*out = r
	return nil
}

//AddrsInfo 对应代币持有人列表，按持币量降序
func (t *Token) AddrsInfo(q *TokenAddrsInfoQuery, out *interface{}) error {
	if len(q.Tokens) == 0 {
		return errors.New("null token list")
	}

	var infos []*TokenAddrsInfo
	err := checkQuery(q.Query, nil)
	if err != nil {
		return err
	}
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	//按持币量降序
	if q.Query == nil {
		q.Query = &querypara.Query{}
	}
	q.Sort = append(q.Sort, &querypara.QSort{Key: "balance", Ascending: false})
	for _, symbol := range q.Tokens {
		q.Match = []*querypara.QMatch{
			{Key: "asset_symbol", Value: symbol},
			{Key: "type", Value: account.AccountPersonage},
		}

		r, err := cli.Search(account.DBX, account.TableX, q.Query, decodeAccount)
		if err != nil || r == nil {
			return err
		}
		var addrs []*AddrBalance
		for _, v := range r {
			acc := v.(*account.Account)
			addrs = append(addrs, &AddrBalance{acc.Address, acc.Balance})
		}

		if TokenCache.Get(symbol) == 0 {
			//缓存中不存在则更新缓存
			err = initTokenList(cli)
			if err != nil {
				return err
			}
		}
		total := TokenCache.Get(symbol)
		if total == 0 {
			return errors.New("symbol [" + symbol + "] not exists")
		}

		infos = append(infos, &TokenAddrsInfo{symbol, total, addrs})
	}

	*out = infos
	return nil
}

//ListToken 代币列表及其发行量
func (t *Token) ListToken(_ interface{}, out *interface{}) error {
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	err = initTokenList(cli)
	if err != nil {
		return err
	}
	*out = TokenCache.TokenList

	return nil
}

//TokenCount 代币种类总数量
func (t *Token) TokenCount(q *querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	initQuery(q)
	r, err := cli.Count(token.TokenInfoDB, token.TokenInfoDB, q)
	if err != nil {
		return err
	}
	*out = r
	return nil
}

//TokenAddrCount 代币持有人地址数量
func (t *Token) TokenAddrCount(qs []*querypara.Query, out *interface{}) error {
	if len(qs) == 0 {
		return errors.New("null token list")
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	var counts []int64
	for _, q := range qs {
		if q == nil {
			counts = append(counts, 0)
			continue
		}

		//q.Match = append(q.Match, &escli.QMatch{Key: "type", Value: account.AccountPersonage})
		r, err := cli.Count(account.DBX, account.DBX, q)
		if err != nil {
			return errors.Wrap(err, "not exist")
		}
		counts = append(counts, r)
	}

	if len(qs) != len(counts) {
		return errors.New("length of tokens and counts not equal")
	}
	*out = counts
	return nil
}

func decodeTokenInfo(x *json.RawMessage) (interface{}, error) {
	t := token.Token{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}

func initQuery(q *querypara.Query) {
	if q == nil {
		q = new(querypara.Query)
		fmt.Println("debug =========", q)
	}
}
