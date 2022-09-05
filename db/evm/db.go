package evm

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli/querypara"
	lru "github.com/hashicorp/golang-lru"
)

type TokenDB interface {
	GetToken(contractAddr, ownerAddr string, goodsID string) (*Token, error)
	GetTokens(contractAddr string, goodsID string) ([]*Token, error)
	GetNft(contractAddr string, goodsID string) (*Token, error)
	UpdateCache(token *Token)
	GetTokenList(contractAddr, ownerAddr string, ids []string) ([]*Token, error)
	GetTokenMap(contractAddr, ownerAddr string, ids []string) (map[string]*Token, error)
}

// TokenDB TokenDB
type tokenDB struct {
	db    db.WrapDB
	cache *lru.Cache
}

// NewTokenDB NewTokenDB
func NewTokenDB(db db.WrapDB) TokenDB {
	d := &tokenDB{
		db: db,
	}
	d.cache, _ = lru.New(2048)
	return d
}

func (d *tokenDB) GetToken(contractAddr, ownerAddr, goodsID string) (*Token, error) {
	id := TokenID(contractAddr, ownerAddr, goodsID)
	// 缓存：由于es更新发生在整个区块处理完成之后，在一个区块多条交易时，没有缓存可能会导致数量统计错误
	// 注意更新缓存内容
	if c, ok := d.cache.Get(id); ok {
		if v, ok := c.(*Token); ok {
			return v, nil
		}
		d.cache.Remove(id)
	}
	token, err := d.db.Get(EVMTokenX, EVMTokenX, id)
	if err != nil {
		return nil, err
	}
	var t Token
	err = json.Unmarshal(*token, &t)
	if err != nil {
		return nil, err
	}
	d.cache.Add(id, t)
	return &t, nil
}

func (d *tokenDB) UpdateCache(token *Token) {
	if token == nil {
		return
	}
	d.cache.Add(token.Key(), token)
}

// GetTokens 获取单个id的token所有信息，FT可能有多份，在不同用户手上
func (d *tokenDB) GetTokens(contractAddr, goodsID string) ([]*Token, error) {
	records, err := d.db.List(EVMTokenX, EVMTokenX, []*db.ListKV{{Key: "contract_addr", Value: contractAddr}, {Key: "goods_id", Value: goodsID}})
	if err != nil {
		return nil, err
	}

	tokens := make([]*Token, 0)
	for _, r := range records {
		var t Token
		err = json.Unmarshal(*r, &t)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &t)
	}
	return tokens, nil
}

func (d *tokenDB) GetNft(contractAddr, goodsID string) (*Token, error) {
	tokens, err := d.GetTokens(contractAddr, goodsID)
	if err != nil {
		return nil, err
	}
	if len(tokens) <= 0 {
		return nil, db.ErrDBNotFound
	}
	return tokens[0], nil
}

func TokenID(contractAddr, ownerAddr, goodsID string) string {
	return fmt.Sprintf("%s-%s-%s-%s", "token", contractAddr, ownerAddr, goodsID)
}

// GetTokenList 获取某用户的多个id通证
func (d *tokenDB) GetTokenList(contractAddr, ownerAddr string, ids []string) ([]*Token, error) {
	query := querypara.Query{
		MatchOne: make([]*querypara.QMatch, 0, len(ids)),
		Match:    []*querypara.QMatch{{Key: "owner", Value: ownerAddr}, {Key: "contract_addr", Value: contractAddr}},
	}
	var resp []interface{}
	for i, id := range ids {
		query.MatchOne = append(query.MatchOne, &querypara.QMatch{Key: "token_id", Value: id})
		if i%1000 == 0 { // 太长ES查询会报错
			rTmp, err := d.db.Search(EVMTokenX, EVMTokenX, &query, decodeToken)
			if err != nil {
				return nil, err
			}
			resp = append(resp, rTmp...)
			query.MatchOne = nil
		}
	}
	rTmp, err := d.db.Search(EVMTokenX, EVMTokenX, &query, decodeToken)
	if err != nil {
		return nil, err
	}
	resp = append(resp, rTmp...)

	tokens := make([]*Token, 0, len(resp))
	for _, v := range resp {
		tokens = append(tokens, v.(*Token))
	}
	return tokens, nil
}

// GetTokenMap GetTokenList and turn to map
func (d *tokenDB) GetTokenMap(contractAddr, ownerAddr string, ids []string) (map[string]*Token, error) {
	tokens, err := d.GetTokenList(contractAddr, ownerAddr, ids)
	if err != nil {
		return nil, err
	}

	ans := make(map[string]*Token)
	for _, t := range tokens {
		ans[t.TokenID] = t
	}
	return ans, nil
}

func decodeToken(x *json.RawMessage) (interface{}, error) {
	r := Token{}
	err := json.Unmarshal(*x, &r)
	return &r, err
}
