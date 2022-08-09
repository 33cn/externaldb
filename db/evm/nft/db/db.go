package db

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

type TokenDB interface {
	GetToken(contractAddr, ownerAddr string, goodsID int64) (*Token, error)
	GetTokens(contractAddr string, goodsID int64) ([]*Token, error)
	GetNft(contractAddr string, goodsID int64) (*Token, error)
}

// TokenDB TokenDB
type tokenDB struct {
	db db.WrapDB
}

// NewTokenDB NewTokenDB
func NewTokenDB(db db.WrapDB) TokenDB {
	d := &tokenDB{
		db: db,
	}
	return d
}

func (d *tokenDB) GetToken(contractAddr, ownerAddr string, goodsID int64) (*Token, error) {
	token, err := d.db.Get(TokenX, TokenX, TokenID(contractAddr, ownerAddr, goodsID))
	if err != nil {
		return nil, err
	}
	var t Token
	err = json.Unmarshal(*token, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *tokenDB) GetTokens(contractAddr string, goodsID int64) ([]*Token, error) {
	records, err := d.db.List(TokenX, TokenX, []*db.ListKV{{Key: "contract_addr", Value: contractAddr}, {Key: "goods_id", Value: goodsID}})
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

func (d *tokenDB) GetNft(contractAddr string, goodsID int64) (*Token, error) {
	tokens, err := d.GetTokens(contractAddr, goodsID)
	if err != nil {
		return nil, err
	}
	if len(tokens) <= 0 {
		return nil, db.ErrDBNotFound
	}
	return tokens[0], nil
}
