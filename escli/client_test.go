package escli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIndex TestIndex
func TestIndex(t *testing.T) {
	mapping := `{
    "settings":{
        "number_of_shards":5,
        "number_of_replicas":2
    },
    "mappings":{
        "properties":{
            "sync_seq":{
                "type":"long"
            }
        }
    }
}`
	cli, err := NewESLongConnect("http://172.16.100.238:9200", "", 7, "elastic", "elastic")
	if err != nil {
		return
	}
	assert.Nil(t, err)

	exists, err := cli.IndexExists("test_setting")
	assert.Nil(t, err)
	assert.False(t, exists)

	create, err := cli.CreateIndex("test_setting", "_doc", mapping)
	assert.Nil(t, err)
	assert.True(t, create)

	//create, err = cli.CreateIndex("test", "seq", mapping)
	//assert.NotNil(t, err)
	//assert.False(t, create)

	//v := `{
	//	"sync_seq" : 666
	//}`
	//err = cli.Update("test", "seq", "6", v)
	//assert.Nil(t, err)
	//
	//v2 := `{
	//	"sync_seq" : "v666"
	//}`
	//err = cli.Update("test", "seq", "7", v2)
	//assert.NotNil(t, err)
	//
	//del, err := cli.DeleteIndex("test")
	//assert.Nil(t, err)
	//assert.True(t, del)
}

/*
func TestIndexMapping(t *testing.T) {
	mapping := unfreeze.TxRecordMapping
	cli, err := NewESClient("http://localhost:9200", "")
	assert.Nil(t, err)

	exists, err := cli.IndexExists("test")
	assert.Nil(t, err)
	assert.False(t, exists)

	create, err := cli.CreateIndex("test", "unfreeze", mapping)
	assert.Nil(t, err)
	assert.True(t, create)

	create, err = cli.CreateIndex("test", "unfreeze", mapping)
	assert.NotNil(t, err)
	assert.False(t, create)

	v := `{
		"beneficiary" : "you",
		"terminate" : {
		   "amount_left" : 10000000,
		   "amount_back" : 570000000
		},
		"success" : true,
		"block" : {
		   "hash" : "hash1",
		   "index" : 2,
		   "height" : 6,
		   "ts" : 15000000
		},
		"unfreeze_id" : "the-id",
		"action_type" : 3,
		"creator" : "me"
	 }`
	err = cli.Update("test", "unfreeze", "6", v)
	assert.Nil(t, err)

	v2 := `{
		"sync_seq" : "v666"
	}`
	err = cli.Update("test", "xx", "7", v2)
	assert.NotNil(t, err)

	del, err := cli.DeleteIndex("test")
	assert.Nil(t, err)
	assert.True(t, del)
}
*/
