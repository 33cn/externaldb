// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"

	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/ticket"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	//"encoding/json"
	//"io/ioutil"
)

var defaultAddresses = []string{
	"1Q5QcUaDXET3RJ3UBurMZzF3gGHyjnFQEa",
	"1M8E2TnHFgxRisCuMXuYHMfELbVk4AkrCh",
	"1PXhFDUTiR69vn9EdbJ9AyviR6wzjsVJuL",
	"19zQrJrtZTkLHDMogRo4vaw3QzNoqTn8aX",
	"1Gv6QVMz24yGMg6UYQ2TXSr64sng48Wj8W",
	"1GR1Ubkt4u2owmSHEWarsYkA1HgCvJ3rFq",
	"1EqgQMSsHxkDHPVZvr8jMQkBAaeC1Ujkn5",
	"1Ff2bGhKS5YFEbNBWrHwzbChizrMZhyi8e",
	"19MaTsSzsZBTpFcubANJ1DMVXKLfaA92Zv",
	"13SnepxL5RELnLwxGTWFkMJMgMJqJhP4FV",
	"1K5aR21sU7xDEwcQ7G1fPXbCJW6R5kLxjZ",
	"1Laq81McJfNwFB2uqdJTosogoaBVm8x6Z2",
	"121qqR7Kb7CcwuHfx9GYkSZhx6L6w8GmzH",
	"1EVGErxMR45eWu28moAQank4A8dQf5b698",
	"19QPCDcrPPPnYDAKeMPCaBFnra5vdjcUoR",
	"1Hy8oXArAPSWFu1P3hHvUaxCNjacnwLKeb",
	"16vAAd4WeqWMCJjeTciUTvJHVj9GuPAyWQ",
	"18AiuYrj2UKUBeFKrCw5mLCC9xyyWxiu4z",
	"137esbqhb3fv8gQMWZEZAqxmSktzXA9jSr",
	"1M4ns1eGHdHak3SNc2UTQB75vnXyJQd91s",
	"1C9M1RCv2e9b4GThN9ddBgyxAphqMgh5zq",
	"1HFUhgxarjC7JLru1FLEY6aJbQvCSL58CB",
	"1EwkKd9iU1pL2ZwmRAC5RrBoqFD1aMrQ2",
	"19ozyoUGPAQ9spsFiz9CJfnUCFeszpaFuF",
	"1NbLdeZYqdaYHFyjBu9evsAmBzfRpiQAZP",
	"1MoEnCDhXZ6Qv5fNDGYoW6MVEBTBK62HP2",
	"12T8QfKbCRBhQdRfnAfFbUwdnH7TDTm4vx",
	"1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J",
	"12pLjJkEAzJCnX9o1enC1yF5MpAffexaKi",
	"1E4UF3HW8LEwdbJjF9FzKhNFeqhp5zNm8",
}

//MinerAccount 挖矿账户
type MinerAccount struct {
	*DBRead
}

//Echo 打印
func (*MinerAccount) Echo(in *string, out *interface{}) error {
	if in == nil {
		return types.ErrInvalidParam
	}
	*out = *in
	return nil
}

type miner struct {
	Address            string `json:"address"`
	LastMineTs         int64  `json:"last_mine_ts"`
	LastMineHeight     int64  `json:"last_mine_height"`
	LastHourMinerCount int64  `json:"mine_count_last_hour"`
	LastDayMinerCount  int64  `json:"mine_count_last_day"`
}

// ShowAccounts ShowAccounts
func (m *MinerAccount) ShowAccounts(in *Accounts, out *interface{}) error {
	acc1 := account.Account{
		AssetSymbol: m.Symbol,
		AssetExec:   account.ExecCoinsX,
	}
	acc1.Detall = &account.Detall{}
	acc2 := account.Account{
		AssetSymbol: m.Symbol,
		AssetExec:   account.ExecCoinsX,
	}
	acc2.Detall = &account.Detall{
		Exec: "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
	}

	cli, err := escli.NewESShortConnect(m.Host, m.Prefix, m.Version, m.Username, m.Password)
	if err != nil {
		return err
	}

	addrs := defaultAddresses
	if in != nil && len(in.Addresses) > 0 {
		addrs = in.Addresses
	}
	ids := make([]string, 0)
	for _, addr := range addrs {
		acc1.Address = addr
		acc2.Address = addr
		ids = append(ids, acc1.Key())
		ids = append(ids, acc2.Key())
	}
	resp, err := cli.MGet(account.DBX, account.TableX, ids, decodeAccount)
	if err != nil {
		return err
	}
	*out = resp

	return nil
}

func decodeTicket(x *json.RawMessage) (interface{}, error) {
	t := ticket.Ticket{}
	err := json.Unmarshal([]byte(*x), &t)
	return &t, err
}

// ShowMinerStatus ShowMinerStatus
func (m *MinerAccount) ShowMinerStatus(in *Accounts, out *interface{}) error {
	cli, err := escli.NewESShortConnect(m.Host, m.Prefix, m.Version, m.Username, m.Password)
	if err != nil {
		return err
	}

	addrs := defaultAddresses
	if in != nil && len(in.Addresses) > 0 {
		addrs = in.Addresses
	}

	now := types.Now().Unix()
	log.Debug("ts", "now", now)
	resp := make([]miner, 0)
	for _, addr := range addrs {
		m1 := miner{
			Address: addr,
		}
		t, h, err := lastMiner(cli, addr)
		if err == nil {
			m1.LastMineTs = t
			m1.LastMineHeight = h
		}
		c1, err := lastMinerCount(cli, addr, now-3600)
		if err == nil {
			m1.LastHourMinerCount = c1
		}
		c2, err := lastMinerCount(cli, addr, now-24*3600)
		if err == nil {
			m1.LastDayMinerCount = c2
		}
		resp = append(resp, m1)
	}

	*out = resp

	return nil
}

func lastMiner(cli escli.ESClient, addr string) (ts, height int64, err error) {
	q1 := &querypara.Query{
		Page: &querypara.QPage{
			Size:   1,
			Number: 1,
		},
		Sort: []*querypara.QSort{
			{
				Key:       "miner_at.ts",
				Ascending: false,
			},
		},
		Match: []*querypara.QMatch{
			{
				Key:   "owner",
				Value: addr,
			},
		},
	}
	//修改
	r, err := cli.Search(ticket.TicketDBX, ticket.TicketTableX, q1, decodeTicket)
	if err != nil || r == nil || len(r) == 0 {
		return
	}
	if t, ok := r[0].(*ticket.Ticket); ok {
		ts = t.MinerAt.Ts
		height = t.MinerAt.Height
	}
	return
}

func lastMinerCount(cli escli.ESClient, addr string, after int64) (count int64, err error) {
	q1 := &querypara.Query{
		Range: []*querypara.QRange{
			{
				Key:    "miner_at.ts",
				RStart: after,
			},
		},
		Match: []*querypara.QMatch{
			{
				Key:   "owner",
				Value: addr,
			},
		},
	}
	count, err = cli.Count(ticket.TicketDBX, ticket.TicketTableX, q1)
	return
}
