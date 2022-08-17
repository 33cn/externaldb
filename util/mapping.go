package util

import (
	"encoding/json"

	"github.com/33cn/externaldb/proto"
)

var mappings = make(map[string]interface{})

func InitMapSet(esVersion int32, esIndex *proto.ESIndex) {
	if esVersion == 7 {
		//设置索引的分片和副本数
		settingMap(esIndex.NumberOfShards, esIndex.NumberOfReplicas)
	} else if esVersion != 6 && esVersion != 7 {
		panic("es version err about InitMapSet")
	}
}

//es7时设置分片和副本数
func settingMap(numberOfShards, numberOfReplicas int32) {
	mappings["settings"] = map[string]interface{}{
		"number_of_shards":   numberOfShards,   //指索引要做多少个分片，只能在创建索引时指定，后期无法修改
		"number_of_replicas": numberOfReplicas, //指每个分片有多少个副本，后期可以动态修改
	}
}

//组合mapping结构
func CombineMap(mapping, typ string, version int32) (string, error) {
	para := make(map[string]interface{})
	if err := json.Unmarshal([]byte(mapping), &para); err != nil {
		log.Error("json mapping Unmarshal", "err", err, "mapping", mapping)
		return "", err
	}

	switch version {
	case 6: //es6，将“mappings”改为对应type名
		for k := range mappings {
			delete(mappings, k)
		}
		mappings[typ] = para["mappings"]
	case 7:
		mappings["mappings"] = para["mappings"]
	}

	res, err := json.Marshal(mappings)
	if err != nil {
		log.Error("json mapping failed", "err", err)
		return "", err
	}
	return string(res), nil
}
