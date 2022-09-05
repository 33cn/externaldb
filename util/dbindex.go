package util

import (
	"github.com/33cn/externaldb/db"
	"github.com/pkg/errors"
)

// InitIndex create index/type mapping if not exist
func InitIndex(cli db.DBCreator, index, typ, mapping string) error {
	exist, err := cli.Exists(index)
	if err != nil {
		log.Error("IndexExists failed", "err", err, "index", index)
		return errors.Wrapf(err, "check IndexExists:%s failed", index)
	}
	if !exist {
		log.Info("CreateIndex", "index", index)

		//根据es版本的不同，组合mapping
		version := cli.GetVersion()
		res, err := CombineMap(mapping, typ, version)
		if err != nil {
			return err
		}
		if res == "" {
			log.Error("combineMap fail")
			return errors.New("combineMap res result is nil")
		}

		//es7：无需typ参数,默认"_doc"
		//es6：正常获取typ参数，创建index
		_, err = cli.Create(index, typ, res)
		if err != nil {
			return errors.Wrapf(err, "CreateIndex:%s failed", index)
		}
	}
	return nil
}
