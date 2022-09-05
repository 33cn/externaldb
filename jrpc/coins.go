package main

import (
	"encoding/json"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/stat/block"
	"github.com/pkg/errors"
)

// BlockStat BlockStat
type BlockStat struct {
	*DBRead
}

// Heights Heights
type Heights struct {
	Height []int64 `json:"height"`
}

// Gets gets
func (b *BlockStat) Gets(req *Heights, out *interface{}) error {
	if req == nil || len(req.Height) == 0 {
		return nil
	}
	ids := make([]string, 0)
	for _, h := range req.Height {
		s := block.BStat{Height: h}
		ids = append(ids, s.ID())
		if h > 0 {
			s2 := block.BStat{Height: h - 1}
			ids = append(ids, s2.ID())
		}
	}

	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	resp, err := cli.MGet(block.DBStatX, block.DBStatX, ids, decodeBlockStat)
	if err != nil {
		return err
	}

	output, err := withoutStat(resp, true)
	*out = output
	return err
}

func toBlockStat(in []interface{}) ([]*block.BStat, error) {
	blocks := make([]*block.BStat, 0)
	for _, one := range in {
		bs, ok := one.(*block.BStat)
		if !ok {
			return nil, errors.Wrap(errors.New("internal error"), "get stat failed")
		}
		blocks = append(blocks, bs)
	}
	return blocks, nil
}

func withoutStat(in []interface{}, skip bool) ([]*block.BStat, error) {
	blocks, err := toBlockStat(in)
	if err != nil {
		return nil, err
	}

	output := make([]*block.BStat, 0)
	if skip {
		count := len(blocks)
		for i := 0; i < count; {
			if blocks[i].Height > 0 && i+1 < count {
				first := blocks[i]
				second := blocks[i+1]
				bs := block.BStat{
					Height:  first.Height,
					Time:    first.Time,
					TxCount: first.TxCount - second.TxCount,
					Fee:     first.Fee - second.Fee,
					Mine:    first.Mine - second.Mine,
					Coins:   first.Coins,
				}
				output = append(output, &bs)
				i = i + 2
			} else {
				output = append(output, blocks[i])
				i++
			}
		}
	} else {
		count := len(blocks)
		for i := 1; i < count; i++ {
			first := blocks[i-1]
			second := blocks[i]
			if first.Height < second.Height {
				first, second = second, first
			}
			bs := block.BStat{
				Height:  first.Height,
				Time:    first.Time,
				TxCount: first.TxCount - second.TxCount,
				Fee:     first.Fee - second.Fee,
				Mine:    first.Mine - second.Mine,
				Coins:   first.Coins,
			}
			output = append(output, &bs)

		}
	}
	return output, nil
}

// Search search
func (b *BlockStat) Search(req *querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	if req.Page != nil {
		req.Page.Size = req.Page.Size + 1
	}
	// 修改Type
	resp, err := cli.Search(block.DBStatX, block.DBStatX, req, decodeBlockStat)
	if err != nil {
		return err
	}
	output, err := withoutStat(resp, false)
	*out = output
	return err
}

// Searches search
func (b *BlockStat) Searches(reqs []*querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	all := make([][]*block.BStat, 0)
	for _, req := range reqs {
		if req.Page != nil {
			req.Page.Size = req.Page.Size + 1
		}
		resp, err := cli.Search(block.DBStatX, block.DBStatX, req, decodeBlockStat)
		if err != nil {
			return err
		}
		output, err := withoutStat(resp, false)
		if err != nil {
			return err
		}
		all = append(all, output)
	}
	*out = all
	return nil
}

// StatGets gets
func (b *BlockStat) StatGets(req *Heights, out *interface{}) error {
	if req == nil || len(req.Height) == 0 {
		log.Debug("1")
		return nil
	}
	ids := make([]string, 0)
	for _, h := range req.Height {
		s := block.BStat{Height: h}
		ids = append(ids, s.ID())
	}

	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	resp, err := cli.MGet(block.DBStatX, block.DBStatX, ids, decodeBlockStat)
	if err != nil {
		return err
	}
	*out = resp

	return nil
}

func decodeBlockStat(x *json.RawMessage) (interface{}, error) {
	r := block.BStat{}
	err := json.Unmarshal([]byte(*x), &r)
	return &r, err
}

// StatSearch search
func (b *BlockStat) StatSearch(req *querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	resp, err := cli.Search(block.DBStatX, block.DBStatX, req, decodeBlockStat)
	if err != nil {
		return err
	}
	*out = resp
	return nil
}

// StatSearches search
func (b *BlockStat) StatSearches(reqs []*querypara.Query, out *interface{}) error {
	cli, err := escli.NewESShortConnect(b.Host, b.Prefix, b.Version, b.Username, b.Password)
	if err != nil {
		return err
	}
	all := make([][]*block.BStat, 0)
	for _, req := range reqs {
		resp, err := cli.Search(block.DBStatX, block.DBStatX, req, decodeBlockStat)
		if err != nil {
			return err
		}
		blocks := make([]*block.BStat, 0)
		for _, one := range resp {
			if bs, ok := one.(*block.BStat); ok {
				blocks = append(blocks, bs)
			}
		}
		all = append(all, blocks)
	}
	*out = all
	return nil
}
