// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"

	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/pkg/errors"
)

// Accounts s
type Accounts struct {
	Addresses []string `json:"addresses"`
}

func decodeAccount(x *json.RawMessage) (interface{}, error) {
	acc := account.Account{}
	err := json.Unmarshal([]byte(*x), &acc)
	return &acc, err
}

// Account Account
type Account struct {
	*DBRead
}

// ListAsset ListAsset
func (t *Account) ListAsset(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(errors.New(ErrBadParm), "empty queryPara input")
	}
	// 现在币种不是很多， 不分页就一次返回
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10000,
		}
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	//修改type:account.DBX
	r, err := cli.Search(account.DBX, account.DBX, q, decodeAccount)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// Count count account
func (t *Account) Count(q *querypara.Query, out *interface{}) error {
	var err error
	*out, err = t.count(account.DBX, account.DBX, q)
	return err
}

// Searches search account
func (t *Account) Searches(reqs []*querypara.Query, out *interface{}) error {
	if len(reqs) == 0 {
		return nil
	}

	resp, err := t.DBRead.searches(account.DBX, account.DBX, reqs, decodeAccount)
	if err != nil {
		return err
	}

	*out = resp
	return err
}

// Search search
func (t *Account) Search(req *querypara.Query, out *interface{}) error {
	resp, err := t.DBRead.search(account.DBX, account.DBX, req, decodeAccount)
	if err != nil {
		return err
	}

	*out = resp
	return err
}

// AccountRecords 获取账户历史余额变化
func (t *Account) AccountRecords(reqs []*querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	all := make([][]*account.Account, 0)
	for _, req := range reqs {
		resp, err := cli.Search(account.AccountRecordDBX, account.AccountRecordTableX, req, decodeAccount)
		if err != nil {
			return err
		}
		accounts := make([]*account.Account, 0)
		for _, one := range resp {
			if bs, ok := one.(*account.Account); ok {
				accounts = append(accounts, bs)
			}
		}
		all = append(all, accounts)
	}
	*out = all
	return nil
}
