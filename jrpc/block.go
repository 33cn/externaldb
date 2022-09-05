package main

import (
	"encoding/json"

	"github.com/33cn/externaldb/db/blockinfo"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/pkg/errors"
)

type Block struct {
	*DBRead
}

func decodeBlock(x *json.RawMessage) (interface{}, error) {
	b := blockinfo.BlockInfo{}
	err := json.Unmarshal([]byte(*x), &b)
	return &b, err
}

//BlockList block list
func (t *Block) BlockList(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.New(ErrBadParm)
	}

	r, err := t.search(blockinfo.TableName, blockinfo.TableName, q, decodeBlock)
	if err != nil {
		return err
	}
	*out = r
	return nil
}

// Count block count
func (t *Block) Count(q *querypara.Query, out *interface{}) error {
	var err error
	*out, err = t.count(blockinfo.TableName, blockinfo.TableName, q)
	return err
}
