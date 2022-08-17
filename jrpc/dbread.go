package main

import (
	"encoding/json"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
)

// DBRead DBRead
type DBRead struct {
	Title    string
	Prefix   string
	Host     string
	Symbol   string
	Version  int32
	Username string
	Password string
}

// Errors
const (
	ErrBadParm   = "Bad Parm"
	ErrTypeAsset = "Type Asset failed"
	// ErrSearchSize = "Search Return Size not match"
)

type decodeDBRecord func(x *json.RawMessage) (interface{}, error)

// get get
func (b *DBRead) get(db, table string, id string) (*json.RawMessage, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	return cli.Get(db, table, id)
}

// gets gets
func (b *DBRead) gets(db, table string, ids []string, decode decodeDBRecord) ([]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	return cli.MGet(db, table, ids, decode)
}

// searches searches
func (b *DBRead) searches(db, table string, reqs []*querypara.Query, decode decodeDBRecord) ([][]interface{}, error) {
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

// search search
func (b *DBRead) search(db, table string, req *querypara.Query, decode decodeDBRecord) ([]interface{}, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return nil, err
	}
	return cli.Search(db, table, req, decode)
}

// count 交易数
func (b *DBRead) count(db, table string, q *querypara.Query) (int64, error) {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return 0, err
	}

	initQuery(q)
	return cli.Count(db, table, q)
}
