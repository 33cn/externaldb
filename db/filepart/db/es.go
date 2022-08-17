package db

import (
	"encoding/json"
	"time"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/util"
)

const (
	TableName         = "file_part"
	KeyPrefix         = TableName + "-"
	FilePartCacheTime = 5 // 5 day
)

// AddKeyPrefix 添加索引前缀
func AddKeyPrefix(hash string) string {
	return KeyPrefix + hash
}

// InitESDB 创建db (插件满足接口 db.ExecConvert)
func InitESDB(cli db.DBCreator) error {
	go func() {
		d := NewEsDB(cli.(escli.ESClient))
		t := time.NewTicker(FilePartCacheTime * time.Hour * 24)
		for range t.C {
			if err := d.Clean(); err != nil {
				log.Error("file part cache clean", "err", err)
			}
		}
	}()
	return util.InitIndex(cli, TableName, TableName, FilePartMapping)
}

// NewEsDB new es DB
func NewEsDB(client db.WrapDB) DB {
	return &esDB{client: client}
}

// EsDB elasticsearch
type esDB struct {
	client db.WrapDB
}

// Get FilePart
func (d *esDB) Get(hash string) (*FilePart, error) {
	buf, err := d.client.Get(TableName, TableName, AddKeyPrefix(hash))
	if err != nil {
		return nil, err
	}
	var ans FilePart
	err = json.Unmarshal(*buf, &ans)
	return &ans, err
}

// Set a file record
func (d *esDB) Set(r *RecordFilePart) error {
	return d.client.Set(TableName, TableName, r.value.Key(), r)
}

// Clean file part cache in es
func (d *esDB) Clean() error {
	q := querypara.Query{
		Range: []*querypara.QRange{
			{
				Key: "insert_time",
				LT:  time.Now().Unix() - FilePartCacheTime*3600*24,
			},
		},
	}
	return d.client.(escli.ESClient).DeleteByQuery(TableName, TableName, &q)
}
