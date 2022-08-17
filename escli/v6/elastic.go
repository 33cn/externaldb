package v6

import (
	"context"
	"encoding/json"
	logx "log"
	"net/http"
	"net/url"
	"os"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	elasticV6 "github.com/olivere/elastic"
	"github.com/pkg/errors"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/escli/status"
)

var (
	log = log15.New("module", "escliV6")
)

type ESClientV6 struct {
	Host     string
	Prefix   string
	Version  int32
	Username string
	Password string
	Client   *elasticV6.Client
}

func NewESLongConnect(host string, prefix string, version int32, username, password string) (*ESClientV6, error) {
	cli := &ESClientV6{Host: host, Prefix: prefix, Version: version, Username: username, Password: password}
	err := cli.Connect()
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func NewESShortConnect(host string, prefix string, version int32, username, password string) (*ESClientV6, error) {
	var err error
	cli := &ESClientV6{Host: host, Prefix: prefix, Version: version}
	errorLog := logx.New(os.Stdout, "APP", logx.LstdFlags)

	cli.Client, err = elasticV6.NewSimpleClient(elasticV6.SetErrorLog(errorLog), elasticV6.SetURL(cli.Host), elasticV6.SetSniff(false), elasticV6.SetBasicAuth(username, password))
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (cli *ESClientV6) index(idx string) string {
	return cli.Prefix + idx
}

func (cli *ESClientV6) Connect() error {
	errorLog := logx.New(os.Stdout, "APP", logx.LstdFlags)
	var err error
	cli.Client, err = elasticV6.NewClient(elasticV6.SetErrorLog(errorLog), elasticV6.SetURL(cli.Host), elasticV6.SetSniff(false), elasticV6.SetBasicAuth(cli.Username, cli.Password))
	if err != nil {
		return err
	}
	info, code, err := cli.Client.Ping(cli.Host).Do(context.Background())
	if err != nil {
		return err
	}
	log.Info("Elasticsearch connect", "code", code, "version", info.Version.Number)

	version, err := cli.Client.ElasticsearchVersion(cli.Host)
	if err != nil {
		return err
	}
	log.Info("Elasticsearch get version", "version", version)
	return nil
}

// Update db operator
func (cli *ESClientV6) Update(index, typ, id string, value string) error {
	_, err := cli.Client.Index().Index(cli.index(index)).Type(typ).Id(id).BodyJson(value).Do(context.Background())
	if err != nil {
		log.Error("escli update failed", "index", index, "type", typ, "id", id, "value", value)
		return err
	}
	return nil
}

// Set db operator:
func (cli *ESClientV6) Set(index, typ, id string, value db.Record) error {
	_, err := cli.Client.Index().Index(cli.index(index)).Type(typ).Id(id).BodyJson(string(value.Value())).Do(context.Background())
	if err != nil {
		log.Error("escli set failed", "index", index, "type", typ, "id", id, "value", value)
		return err
	}
	return nil
}

// Get RawMessage
func (cli *ESClientV6) Get(index, typ, id string) (*json.RawMessage, error) {
	res, err := cli.Client.Get().Index(cli.index(index)).Type(typ).Id(id).Do(context.Background())
	if err != nil {
		if elasticV6.IsNotFound(err) {
			return nil, db.ErrDBNotFound
		}
		log.Error("escli get failed", "index", index, "type", typ, "id", id, "err", err)
		return nil, err
	}
	return res.Source, nil
}

// List RawMessage
func (cli *ESClientV6) List(index, typ string, kv []*db.ListKV) ([]*json.RawMessage, error) {
	if len(kv) == 0 {
		return nil, db.ErrDBBadParam
	}
	query := querypara.Query{
		Match: make([]*querypara.QMatch, 0),
	}
	cnt := len(kv)
	for i := 0; i < cnt; i++ {
		query.Match = append(query.Match, &querypara.QMatch{
			Key:   kv[i].Key,
			Value: kv[i].Value,
		})
	}
	q := elasticV6.NewBoolQuery()
	q = addMatch(&query, q)
	search := cli.Client.Search(cli.index(index)).Type(typ)
	search.Query(q)
	responses, err := search.Do(context.Background())
	if err != nil {
		if elasticV6.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "Search failed")
	}
	rs := make([]*json.RawMessage, 0)
	if responses != nil && responses.Hits != nil && responses.Hits.Hits != nil {
		for _, hit := range responses.Hits.Hits {
			log.Debug("Search", "r", string(*hit.Source))
			rs = append(rs, hit.Source)
		}
	}

	return rs, err
}

// BulkUpdate 一个区块对应的记录一起更新
func (cli *ESClientV6) BulkUpdate(rs []db.Record) error {
	beg := types.Now()
	defer func() {
		log15.Info("BulkUpdate", "cost", types.Since(beg))
	}()
	b := cli.Client.Bulk()
	for _, r := range rs {
		if r.OpType() == db.OpAdd {
			req := elasticV6.NewBulkIndexRequest().Index(cli.index(r.Index())).Type(r.Type()).Id(r.ID()).Doc(json.RawMessage(r.Value()))
			b.Add(req)
		} else if r.OpType() == db.OpDel {
			req := elasticV6.NewBulkDeleteRequest().Index(cli.index(r.Index())).Type(r.Type()).Id(r.ID())
			b.Add(req)
		} else if r.OpType() == db.OpUpdate {
			req := elasticV6.NewBulkUpdateRequest().Index(cli.index(r.Index())).Type(r.Type()).Id(r.ID()).Doc(json.RawMessage(r.Value()))
			b.Add(req)
		}
	}
	responses, err := b.Do(context.Background())
	if err != nil {
		log.Error("Bulk Save failed", "err", err)
	}

	showBulkResult := true
	if showBulkResult && responses != nil {
		for key, response := range responses.Items {
			for k, v := range response {
				log.Debug("Bulk Save Part", "key", key, "op", k, "value", v)
				if v.Status >= 400 {
					source := ""
					if v.GetResult != nil && v.GetResult.Source != nil {
						source = string(*v.GetResult.Source)
					}
					log.Error("Bulk Save Part failed", "key", key, "op", k, "Status", v.Status,
						"err", v.Error, "value", v, "source", source)
				}
			}
		}
	}

	return err
}

// IndexExists check index exists
func (cli *ESClientV6) IndexExists(index string) (bool, error) {
	return cli.Client.IndexExists(cli.index(index)).Do(context.Background())
}

// CreateIndex create index
func (cli *ESClientV6) CreateIndex(index string, typ, mapping string) (bool, error) {
	ret, err := cli.Client.CreateIndex(cli.index(index)).Do(context.Background())
	if err != nil {
		log15.Error("CreateIndex", "err", err, "ret", ret)
		return false, err
	}
	ret2, err := cli.Client.PutMapping().Index(cli.index(index)).Type(typ).BodyString(mapping).Do(context.Background())
	if err != nil {
		cli.Client.DeleteIndex(cli.index(index)).Do(context.Background())
		log15.Error("CreateIndex PutMapping", "err", err, "ret", ret2)
		return false, err
	}
	return true, nil
}

// DeleteIndex  DeleteIndex
func (cli *ESClientV6) DeleteIndex(index string) (bool, error) {
	ret, err := cli.Client.DeleteIndex(cli.index(index)).Do(context.Background())
	if err != nil {
		log15.Info("IndexDelete", "err", err, "ret", ret)
		return false, err
	}
	return true, nil
}

// Exists adapter DBCreator
func (cli *ESClientV6) Exists(name string) (bool, error) {
	return cli.IndexExists(name)
}

// Create adapter DBCreator
func (cli *ESClientV6) Create(name string, table, definitioin string) (bool, error) {
	return cli.CreateIndex(name, table, definitioin)
}

// Delete adapter DBCreator
func (cli *ESClientV6) Delete(name string) (bool, error) {
	return cli.DeleteIndex(name)
}

// DeleteByQuery deletes documents that match a query.
func (cli *ESClientV6) DeleteByQuery(idx, typ string, query *querypara.Query) error {
	service := elasticV6.NewDeleteByQueryService(cli.Client).Index(cli.index(idx)).Type(typ)
	resp, err := service.Query(getQuery(query)).Do(context.Background())
	if err != nil {
		log.Error("DeleteByQuery", "err", err, "resp", resp)
		return err
	}
	return err
}

// Status 获取ES状态信息
func (cli *ESClientV6) Status() (res *status.Status) {
	res = &status.Status{Status: "UP"}
	resp, err := cli.Client.PerformRequest(context.Background(), elasticV6.PerformRequestOptions{
		Method: "GET",
		Path:   "/_nodes/stats",
	})
	if err != nil {
		log15.Info("Status, PerformRequest _nodes/stats", "err", err)
		res.Status = err.Error()
		return
	}
	if err := json.Unmarshal(resp.Body, &res); err != nil {
		log15.Info("Status, json.Unmarshal", "err", err)
		res.Status = err.Error()
		return
	}
	return
}

func (cli *ESClientV6) GetVersion() int32 {
	return cli.Version
}

func (cli *ESClientV6) PerformRequest(method, path string, params url.Values, body interface{}, headers http.Header) (interface{}, error) {
	return cli.Client.PerformRequest(context.Background(), elasticV6.PerformRequestOptions{
		Method:  method,
		Path:    path,
		Params:  params,
		Body:    body,
		Headers: headers,
	})
}
