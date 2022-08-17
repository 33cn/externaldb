package block

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/stat"
	"github.com/33cn/externaldb/util"
)

// db
const (
	DBStatX     = "block_stat"
	NameX       = "block"
	DefaultType = "_doc"
)

const (
	coin = 100000000
)

// BStat 统计从0高度的累计值
type BStat struct {
	Height  int64 `json:"height"`
	Time    int64 `json:"time"`
	TxCount int   `json:"tx_count"`
	Fee     int64 `json:"fee"`
	Mine    int64 `json:"mine"`
	Coins   int64 `json:"coins"`
}

// 记录以往的累计
var (
	prev *BStat
)

func init() {
	converts.RegisterStat(NameX, NewStat)
}

// ID id
func (s *BStat) ID() string {
	return fmt.Sprintf("stat-%d", s.Height)
}

// NewStat NewStat
func NewStat(title, symbol string, genesis, coin int64) stat.Stat {
	if title != "bityuan" {
		return &blockStat{
			title:   title,
			genesis: genesis,
			coin:    coin,
		}
	}
	return &blockStat{title: title}
}

type blockStat struct {
	title   string
	genesis int64
	coin    int64
}

// InitDB init db
func (s *blockStat) InitDB(cli db.DBCreator) error {
	err := util.InitIndex(cli, DBStatX, DBStatX, StatMapping)
	return err
}

// Recover Recover
func (s *blockStat) Recover(client escli.ESClient, lastSeq int64) error {
	if prev != nil {
		return nil
	}
	q := querypara.Query{
		Page: &querypara.QPage{Size: 1, Number: 1},
		Sort: []*querypara.QSort{{Key: "height", Ascending: false}},
	}
	rs, err := client.Search(DBStatX, DBStatX, &q, decode)
	if err != nil {
		return errors.Wrapf(err, "block stat recover search")
	}
	if len(rs) == 0 {
		return nil
	}
	st, ok := rs[0].(*BStat)
	if !ok {
		return errors.Wrapf(err, "block stat recover get value")
	}
	prev = st
	return nil
}

func decode(x *json.RawMessage) (interface{}, error) {
	s := BStat{}
	err := json.Unmarshal([]byte(*x), &s)
	return &s, err
}

// Stat impl
func (s *blockStat) Stat(detail *types.BlockDetail, op int) ([]db.Record, error) {
	cur := statCur(s.title, detail, s.genesis, s.coin)

	if cur.Height == 0 {
		prev = cur
		return []db.Record{makeRecord(cur, op)}, nil
	}

	if cur.Height > 0 && prev != nil {
		factor := 1
		if op == db.SeqTypeDel {
			factor = -1
		}
		ret := BStat{
			Height:  detail.Block.Height,
			Time:    detail.Block.BlockTime,
			TxCount: prev.TxCount + cur.TxCount*factor,
			Fee:     prev.Fee + cur.Fee*int64(factor),
			Mine:    prev.Mine + cur.Mine*int64(factor),
			Coins:   prev.Coins + (cur.Mine-cur.Fee)*int64(factor),
		}
		prev = &ret
		return []db.Record{makeRecord(&ret, op)}, nil
	}

	// 若是不是从零开始， 或没有从数据库中读到上次的值，累计将出错
	return nil, nil
}

func statCur(title string, detail *types.BlockDetail, genesis, coin int64) *BStat {
	cur := BStat{
		Height:  detail.Block.Height,
		Time:    detail.Block.BlockTime,
		TxCount: len(detail.Block.Txs),
	}

	for _, tx := range detail.Block.Txs {
		cur.Fee += tx.GetFee()
	}

	if title == "bityuan" || title == "" {
		if cur.Height == 0 {
			cur.Coins = bityuanGenesis()
		}
		cur.Mine = bityuanMiner(cur.Height)
	} else {
		if cur.Height == 0 {
			cur.Coins = genesis
		}
		cur.Mine = coin
	}
	return &cur
}

func bityuanMiner(height int64) int64 {
	// ForkChainParamV2
	if height >= 2270000 {
		return 8 * coin
	} else if height > 0 {
		return 30 * coin
	}
	return 0
}

func bityuanGenesis() int64 {
	return 317430000 * coin
}

// StatRecord StatRecord
type StatRecord struct {
	*db.IKey
	*db.Op
	s *BStat
}

// Append impl
func (r *StatRecord) Append(r2 db.Record) db.Record {
	return r
}

// Value impl
func (r *StatRecord) Value() []byte {
	v, _ := json.Marshal(r.s)
	return v
}

func makeRecord(s *BStat, op int) *StatRecord {
	return &StatRecord{
		IKey: db.NewIKey(DBStatX, DBStatX, s.ID()),
		Op:   db.NewOp(op),
		s:    s,
	}
}
