package v7

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/33cn/externaldb/escli/querypara"
)

type Detall struct {
	Address  string `json:"address"`
	Exec     string `json:"exec"` // 执行器名 和 执行器地址 等效， 可能不知道执行器名， 这时填地址
	Frozen   int64  `json:"frozen"`
	Balance  int64  `json:"balance"`
	Total    int64  `json:"total"`
	Type     string `json:"type"` // contract, personage
	AddrType string `json:"addr_type"`
}

// Account 记录帐号信息， Detail 解析自执行日志， 其他项在交易解析时获得
type Account struct {
	*Detall
	HeightIndex int64  `json:"height_index"`
	AssetSymbol string `json:"asset_symbol"`
	AssetExec   string `json:"asset_exec"`
	Height      int64  `json:"height"`
	BlockTime   int64  `json:"block_time"`
}

// escli 测试依赖外部数据库，出错是没有外部数据，在有外部数据库情况下，测试接口是否有效，对数据不做要求

func decodeAccount(x *json.RawMessage) (interface{}, error) {
	acc := Account{}

	err := json.Unmarshal([]byte(*x), &acc)
	return &acc, err
}

// Test_MGet Test_MGet
func Test_MGet(t *testing.T) {
	cli, err := NewESLongConnect("http://localhost:9200", "", 7, "elastic", "elastic")
	if err != nil {
		return
	}
	assert.Nil(t, err)

	ids := []string{
		"1FVWscp7ZmXWjkY4oKr8JvkrQZrxUHAnKW-coins-coins:bty",
		"1EezS8sB7me5gcjxP8udGAWuBp3gtsJKvD-coins-coins:bty",
		"not-exist",
		"1G9LN3j7E3L7ZuKtpceptnBt3GikmGPLuW-coins-coins:bty",
	}

	var rs []interface{}
	rs, err = cli.MGet("account", "account", ids, decodeAccount)
	assert.Nil(t, err)
	for _, r := range rs {
		acc := r.(*Account)
		t.Log(*acc, *acc.Detall)
	}
}

func Test_Search(t *testing.T) {
	cli, err := NewESLongConnect("http://localhost:9200", "", 7, "elastic", "elastic")
	if err != nil {
		return
	}
	assert.Nil(t, err)

	query := &querypara.Query{
		Page: &querypara.QPage{
			Size:   20,
			Number: 1,
		},
		Sort: []*querypara.QSort{
			{
				Key:       "balance",
				Ascending: true,
			},
		},
		Range: []*querypara.QRange{
			{
				Key:    "balance",
				RStart: 1,
				// REnd:   33300000,
			},
		},
		Match: []*querypara.QMatch{
			{
				Key:   "asset_symbol",
				Value: "bty",
			},
		},
	}
	v, _ := json.Marshal(query)
	t.Log("queryPara", "input", string(v))
	var rs []interface{}
	rs, err = cli.Search("account", "account", query, decodeAccount)
	if err != nil {
		return
	}
	assert.Nil(t, err)
	for _, r := range rs {
		acc := r.(*Account)
		t.Log(*acc, *acc.Detall)
	}
}
