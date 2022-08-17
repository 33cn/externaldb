package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//修改Mapping去掉了Type
const SeqMapping = `{
    "mappings":{
        "properties":{
            "sync_seq":{
                "type":"long"
            }
        }
    }
}`

const txMapping = `{
    "mappings":{
        "properties":{
            "multi_signature_address":{
                "type":"keyword"
            },
            "tx_hash":{
                "type":"keyword"
            },
            "tx_id":{
                "type":"long"
            },
            "type":{
                "type":"keyword"
            }
        }
    }
}`

func TestCombineMap1(t *testing.T) {
	assert := assert.New(t)
	settingMap(3, 1)
	res, err := CombineMap(SeqMapping, "", 7)
	t.Log(res)
	assert.Nil(err)
}

func TestCombineMap2(t *testing.T) {
	assert := assert.New(t)
	settingMap(3, 1)
	res, err := CombineMap(txMapping, "", 7)
	t.Log(res)
	assert.Nil(err)
}

func TestCombineMap3(t *testing.T) {
	assert := assert.New(t)
	settingMap(3, 1)
	res1, err := CombineMap(txMapping, "account1", 6)
	res2, err := CombineMap(txMapping, "account2", 6)
	t.Log(res1)
	t.Log(res2)
	assert.Nil(err)
}
