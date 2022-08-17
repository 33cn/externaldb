package db

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

const (
	TableName = "file_summary"
	KeyPrefix = TableName + "-"
)

// AddKeyPrefix 添加索引前缀
func AddKeyPrefix(hash string) string {
	return KeyPrefix + hash
}

// InitESDB 创建db (插件满足接口 db.ExecConvert)
func InitESDB(cli db.DBCreator) error {
	return util.InitIndex(cli, TableName, TableName, FileSummaryMapping)
}

// EsDB elasticsearch
type esDB struct {
	client db.WrapDB
}

// NewEsDB new es DB
func NewEsDB(client db.WrapDB) DB {
	return &esDB{client: client}
}

// Get FileSummary
func (d *esDB) Get(hash string) (*FileSummary, error) {
	buf, err := d.client.Get(TableName, TableName, AddKeyPrefix(hash))
	if err != nil {
		return nil, err
	}
	var ans FileSummary
	err = json.Unmarshal(*buf, &ans)
	return &ans, err
}
