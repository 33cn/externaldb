package escli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli/aggdecode"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/escli/status"
	v6 "github.com/33cn/externaldb/escli/v6"
	v7 "github.com/33cn/externaldb/escli/v7"
)

type ESClient interface {
	Update(index, typ, id string, value string) error
	Set(index, typ, id string, value db.Record) error
	Get(index, typ, id string) (*json.RawMessage, error)
	List(index, typ string, kv []*db.ListKV) ([]*json.RawMessage, error)
	BulkUpdate(rs []db.Record) error

	IndexExists(index string) (bool, error)
	CreateIndex(index string, typ, mapping string) (bool, error)
	DeleteIndex(index string) (bool, error)

	MGet(idx, typ string, ids []string, decode func(x *json.RawMessage) (interface{}, error)) ([]interface{}, error)
	Search(idx, typ string, query *querypara.Query, decode func(x *json.RawMessage) (interface{}, error)) ([]interface{}, error)
	Count(idx, typ string, query *querypara.Query) (int64, error)
	Agg(idx, typ string, query *querypara.Search) (*aggdecode.AggregationBucketKeyItem, error)

	Exists(name string) (bool, error)
	Create(name string, table, definition string) (bool, error)
	Delete(name string) (bool, error)

	Status() *status.Status
	GetVersion() int32
	PerformRequest(method, path string, params url.Values, body interface{}, headers http.Header) (interface{}, error)
	DeleteByQuery(idx, typ string, query *querypara.Query) error
}

// NewESLongConnect create db handler
func NewESLongConnect(host string, prefix string, version int32, username, password string) (ESClient, error) {
	switch version {
	case 6:
		return v6.NewESLongConnect(host, prefix, version, username, password)
	case 7:
		return v7.NewESLongConnect(host, prefix, version, username, password)
	default:
		panic("not support es version" + fmt.Sprint(version) + "about NewESLongConnect")
	}
}

// NewESShortConnect create db handler
func NewESShortConnect(host string, prefix string, version int32, username, password string) (ESClient, error) {
	switch version {
	case 6:
		return v6.NewESShortConnect(host, prefix, version, username, password)
	case 7:
		return v7.NewESShortConnect(host, prefix, version, username, password)
	default:
		panic("not support es version" + fmt.Sprint(version) + "about NewESShortConnect")
	}

}
