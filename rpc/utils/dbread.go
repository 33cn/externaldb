package rpcutils

import (
	"encoding/json"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/util"
)

// DBRead DBRead
type DBRead struct {
	Title    string
	Prefix   string
	Host     string
	Symbol   string
	Version  int32
	ID       string
	Username string
	Password string
}

type decodeDBRecord func(x *json.RawMessage) (interface{}, error)

// NewESShortConnect new a es short connect
func (b *DBRead) NewESShortConnect() (escli.ESClient, error) {
	return escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
}

// Gets gets
func (b *DBRead) Gets(db, table string, ids []string, decode decodeDBRecord) ([]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	return cli.MGet(db, table, ids, decode)
}

// Searches searches
func (b *DBRead) Searches(db, table string, reqs []*querypara.Query, decode decodeDBRecord) ([][]interface{}, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	all := make([][]interface{}, 0)
	for _, req := range reqs {
		resp, err := cli.Search(db, table, req, decode)
		if err != nil {
			return nil, err
		}
		all = append(all, resp)
	}
	return all, nil
}

// Search search
func (b *DBRead) Search(db, table string, req *querypara.Query, decode decodeDBRecord) ([]interface{}, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	return cli.Search(db, table, req, decode)
}

// Count 交易数
func (b *DBRead) Count(db, table string, q *querypara.Query) (int64, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return 0, err
	}

	q = initQuery(q)
	return cli.Count(db, table, q)
}

func initQuery(q *querypara.Query) *querypara.Query {
	if q == nil {
		q = new(querypara.Query)
	}
	return q
}

// List 列表
func (b *DBRead) List(db, table string, q *querypara.Query, decoder func(x *json.RawMessage) (interface{}, error)) ([]interface{}, error) {
	q = initQuery(q)
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}
	cli, err := b.NewESShortConnect()
	if err != nil {
		return nil, err
	}
	r, err := cli.Search(db, table, q, decoder)
	if err != nil || r == nil {
		//特殊处理一下没有数据时sort功能报错的场景,取消sort再尝试调用一次
		if q.Sort != nil {
			q.Sort = nil
			r2, err2 := cli.Search(db, table, q, decoder)
			if err2 == nil && r2 != nil {
				return r2, nil
			}
		}
		return nil, err
	}
	return r, nil
}

//LastConvertSeq 获取已经解析的最新seq值
func (b *DBRead) LastConvertSeq() int64 {

	client, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return -1
	}

	num, err := util.LastSyncSeq(client, b.ID)
	if err != nil {
		return -1
	}
	return num

}

// DecodeJSONToMap Decode json data to map[string]interface{}
func DecodeJSONToMap(x *json.RawMessage) (interface{}, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal(*x, &m)
	return m, err
}
