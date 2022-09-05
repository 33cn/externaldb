package block

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/block"
	"github.com/stretchr/testify/assert"
)

func TestBlockStat_Stat(t *testing.T) {
	assert := assert.New(t)
	blockStat := NewStat("bityuan", "", 100000, 1)

	//cli, err := escli.NewESLongConnect("http://172.16.100.249:9200", "seq_")
	// := fmt.Sprintf("%d", 2)
	//result, err := cli.Get("seq", "seq", id)
	result := `{
	"sync_seq":1,
		"from":"http://localhost:8801", 
		"number":1,
		"hash":"0x93596bc60d2427fb8ac1c4e86d3c7a5d1c9c6c1fe3b6fbe92837587a14fb2f55",
		"type":1,
		"block_detall":"CtQCEiDIeZvvrNlwml9N+mjW/VOjfMzoM1KZehMq3+1/RnR3VxogtvV8aAr9ldfPo+6xRt8pbsjzMJIQVF0EgVNzuQiJjsEiIJsRoaS8aD6tcaTkcQZtc4Sksoqqzqf8CKIX5hVdnNhiKAEwzqep+QU63QEKBWNvaW5zEi0YAQopEIDC1y8iIjFDc2loM3U2TkhBcmpKZWVoTjkyVDFad3ljWUtpdDJSN0YabQgBEiECia8vnVMQlxJCQCkyq1CphaIQNnFqNWwkGSU8JH0ogy4aRjBEAiAcaLnsMJirxmI31rs9LK7A0dDTX0Fcts31Od9Kp+DGcgIgFHo5IKRMCSB2uSlAYpTW99/ZW5jaX85XYz92+0hR+HggoI0GKMaoqfkFMOul0euRnsvZRToiMUNzaWgzdTZOSEFyakplZWhOOTJUMVp3eWNZS2l0MlI3Rlj//4P4ARKhAggCGmIIAhJeCi0QgICE/qbe4REiIjFDYkVWVDlSbk01b1poV01qNGZ4VXJKWDk0VnRSb3R6dnMSLRDg8v39pt7hESIiMUNiRVZUOVJuTTVvWmhXTWo0ZnhVckpYOTRWdFJvdHp2cxpiCAMSXgotEODy/f2m3uERIiIxQ2JFVlQ5Um5NNW9aaFdNajRmeFVySlg5NFZ0Um90enZzEi0Q4LCmzqbe4REiIjFDYkVWVDlSbk01b1poV01qNGZ4VXJKWDk0VnRSb3R6dnMaVQgDElEKJCIiMUNzaWgzdTZOSEFyakplZWhOOTJUMVp3eWNZS2l0MlI3RhIpEIDC1y8iIjFDc2loM3U2TkhBcmpKZWVoTjkyVDFad3ljWUtpdDJSN0Y="
}`

	json.Marshal(result)
	var seq block.Seq
	err := json.Unmarshal([]byte(result), &seq)
	assert.Nil(err)

	var detail types.BlockDetail
	err = types.Decode(seq.BlockDetail, &detail)
	assert.Nil(err)

	record, err := blockStat.Stat(&detail, seq.Type)
	t.Log("block height", detail.Block.Height)
	t.Log("record is", record)

	assert.Nil(err)
}

func TestBlockStat_Stat1(t *testing.T) {
	assert := assert.New(t)
	blockStat := NewStat("bityuan", "", 100000, 1)

	detail1 := &types.BlockDetail{
		Block: &types.Block{
			Version:   1,
			Height:    0,
			BlockTime: time.Now().Unix(),
			Txs: []*types.Transaction{
				{Fee: 2},
			},
		},
	}
	record1, err := blockStat.Stat(detail1, 1)
	assert.Nil(err)
	for i, v := range record1 {
		t.Log("block height", detail1.Block.Height)
		t.Log("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}

	result := `{
	"sync_seq":1,
		"from":"http://localhost:8801", 
		"number":1,
		"hash":"0x93596bc60d2427fb8ac1c4e86d3c7a5d1c9c6c1fe3b6fbe92837587a14fb2f55",
		"type":1,
		"block_detall":"CtQCEiDIeZvvrNlwml9N+mjW/VOjfMzoM1KZehMq3+1/RnR3VxogtvV8aAr9ldfPo+6xRt8pbsjzMJIQVF0EgVNzuQiJjsEiIJsRoaS8aD6tcaTkcQZtc4Sksoqqzqf8CKIX5hVdnNhiKAEwzqep+QU63QEKBWNvaW5zEi0YAQopEIDC1y8iIjFDc2loM3U2TkhBcmpKZWVoTjkyVDFad3ljWUtpdDJSN0YabQgBEiECia8vnVMQlxJCQCkyq1CphaIQNnFqNWwkGSU8JH0ogy4aRjBEAiAcaLnsMJirxmI31rs9LK7A0dDTX0Fcts31Od9Kp+DGcgIgFHo5IKRMCSB2uSlAYpTW99/ZW5jaX85XYz92+0hR+HggoI0GKMaoqfkFMOul0euRnsvZRToiMUNzaWgzdTZOSEFyakplZWhOOTJUMVp3eWNZS2l0MlI3Rlj//4P4ARKhAggCGmIIAhJeCi0QgICE/qbe4REiIjFDYkVWVDlSbk01b1poV01qNGZ4VXJKWDk0VnRSb3R6dnMSLRDg8v39pt7hESIiMUNiRVZUOVJuTTVvWmhXTWo0ZnhVckpYOTRWdFJvdHp2cxpiCAMSXgotEODy/f2m3uERIiIxQ2JFVlQ5Um5NNW9aaFdNajRmeFVySlg5NFZ0Um90enZzEi0Q4LCmzqbe4REiIjFDYkVWVDlSbk01b1poV01qNGZ4VXJKWDk0VnRSb3R6dnMaVQgDElEKJCIiMUNzaWgzdTZOSEFyakplZWhOOTJUMVp3eWNZS2l0MlI3RhIpEIDC1y8iIjFDc2loM3U2TkhBcmpKZWVoTjkyVDFad3ljWUtpdDJSN0Y="
}`
	json.Marshal(result)
	var seq block.Seq
	err = json.Unmarshal([]byte(result), &seq)
	assert.Nil(err)

	var detail2 types.BlockDetail
	err = types.Decode(seq.BlockDetail, &detail2)
	assert.Nil(err)

	record2, err := blockStat.Stat(&detail2, seq.Type)
	assert.Nil(err)
	for i, v := range record2 {
		t.Log("block height", detail2.Block.Height)
		t.Log("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}

}

func TestBlockStat_Stat2(t *testing.T) {
	assert := assert.New(t)
	blockStat := NewStat("bit", "", 100000000, 10)

	detail1 := &types.BlockDetail{
		Block: &types.Block{
			Version:   1,
			Height:    0,
			BlockTime: time.Now().Unix(),
			Txs: []*types.Transaction{
				{Fee: 2},
			},
		},
	}
	detail2 := &types.BlockDetail{
		Block: &types.Block{
			Version:   1,
			Height:    1,
			BlockTime: time.Now().Unix(),
			Txs: []*types.Transaction{
				{Fee: 2},
			},
		},
	}
	detail3 := &types.BlockDetail{
		Block: &types.Block{
			Version:   1,
			Height:    2,
			BlockTime: time.Now().Unix(),
			Txs: []*types.Transaction{
				{Fee: 2},
			},
		},
	}

	record1, err := blockStat.Stat(detail1, 1)
	for i, v := range record1 {
		t.Log("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}
	record2, err := blockStat.Stat(detail2, 1)
	for i, v := range record2 {
		t.Log("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}
	record3, err := blockStat.Stat(detail3, 1)
	for i, v := range record3 {
		t.Log("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
	}
	assert.Nil(err)
}
