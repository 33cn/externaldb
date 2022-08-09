package sync

import (
	"errors"
	"math"
	"sync"

	"github.com/33cn/externaldb/proto"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/store"
)

var (
	rollbackSeq []int64
	log         = l.New("module", "sync")
)

func initRollbackSeq(seqs []int64) {
	if seqs != nil {
		log.Info("badBlock", "seqs", seqs)
		rollbackSeq = seqs
	}
}

// NewSeqsProc 将推送和处理等功能组合起来
func NewSeqsProc(seqNumStore store.SeqNumStore, seqStore store.SeqStore, chain *proto.Chain33) (*SeqsProc, error) {
	initRollbackSeq(chain.RollbackSeq)

	s := &SeqsProc{
		seqStore:    seqStore,
		seqNumStore: seqNumStore,
	}
	s.blockCh = make(chan *types.BlockSeqs)
	s.resultCh = make(chan error)

	// recover when init SeqsProcl
	db, err := s.seqNumStore.LastSeq()
	if err != nil {
		return nil, err
	}
	db.From = chain.Host
	log.Info("recover", "db_status", db)

	s.seqNum = db

	return s, nil
}

var _ SeqSaver = &SeqsProc{}

// SeqsProc SeqsProc
type SeqsProc struct {
	mutex    sync.Mutex
	blockCh  chan *types.BlockSeqs
	resultCh chan error

	seqStore    store.SeqStore
	seqNumStore store.SeqNumStore
	seqNum      *store.SeqNum

	// delBlocksVersion string
}

// Save save seq
func (p *SeqsProc) Save(block *types.BlockSeqs) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.blockCh <- block
	err := <-p.resultCh
	return err
}

func (p *SeqsProc) Proc(startSeq int64) {
	// 配置是从0开始同步(指next), number 表示已经同步了的seq
	// 所以 next = max(number+1, startSeq)
	if p.seqNum.Number < startSeq-1 {
		p.seqNum.Number = startSeq - 1
	}

	for {
		blocks := <-p.blockCh
		if len(blocks.Seqs) > 0 {
			log.Info("to deal request", "seq", blocks.Seqs[0].Num, "count", len(blocks.Seqs))
		}
		p.resultCh <- DealBlocks(blocks, p.seqNum, p.seqNumStore, p.seqStore)
	}
}

// 保存区块步骤
// 1. 记录 seqNumber ->  seq
// 2. 记录 lastseq
// 3. 更新高度
//
// 重启恢复
// 1. 看高度， 对应高度是已经完成的
// 2. 继续重新下一个高度即可。 重复写， 幂等
// 所以不需要恢复过程， 读出高度即可

// DealBlocks 处理输入流程
func DealBlocks(seqs *types.BlockSeqs, d *store.SeqNum, seqNumStore store.SeqNumStore, seqStore store.SeqStore) error {
	beg := types.Now()
	defer func() {
		log.Info("dealBlocks", "cost", types.Since(beg))
	}()
	count, start, seqsMap, err := parseSeqs(seqs)
	if err != nil {
		log.Error("deal request", "parseSeqs", err)
		return err
	}
	log.Info("dealBlocks", "p1", "parseSeqs", "cost", types.Since(beg))

	// 在app 端保存成功， 但回复ok时，程序挂掉, 记录日志
	if start <= d.Number {
		log.Error("deal request", "seq", start, "current_seq", d.Number)
	}
	if start+int64(count-1) <= d.Number {
		return nil
	}

	// 一般在配置有误时发生， 在同一个节点上配置相同的推送名
	if start > d.Number+1 {
		log.Error("deal request", "seq", start, "current_seq", d.Number)
		return errors.New("bad seq")
	}

	number := d.Number
	records := make([]db.Record, 0)
	for i := 0; i < count; i++ {
		if start+int64(i) <= d.Number {
			continue
		}

		number++
		if seq, ok := seqsMap[int64(i)+start]; ok {
			blockSeq := &block.Seq{
				SyncSeq:     int(number),
				From:        d.From,
				Number:      int(seq.Num),
				Hash:        common.ToHex(seq.Seq.Hash),
				Type:        int(seq.Seq.Type),
				BlockDetail: types.Encode(seq.Detail),
			}
			seqRecord := block.NewSeqRecord(blockSeq)
			records = append(records, seqRecord)
		} else {
			continue
		}
	}
	lastSeq := block.NewLastRecord(number)
	log.Info("dealBlocks", "p2", "genRecord", "cost", types.Since(beg))
	log.Info("newLastRecord", "seq", lastSeq.ID())

	// save
	err = seqStore.SaveSeqs(records)
	if err == nil {
		err = seqNumStore.UpdateLastSeq(lastSeq)
	}

	if err == nil {
		d.Number = number
	}
	log.Info("dealBlocks", "p3", "callSaveDB", "cost", types.Since(beg))

	return err
}

// 检查输入是否有问题, 并解析输入
func parseSeqs(seqs *types.BlockSeqs) (totalCount int, start int64, seqsOrder map[int64]*types.BlockSeq, err error) {
	seqsOrder = make(map[int64]*types.BlockSeq)
	totalCount = len(seqs.Seqs)
	count := len(seqs.Seqs)
	start = math.MaxInt64 //int64(^uint64(0) >> 1)

dealRollback:
	for i := 0; i < totalCount; i++ {
		if seqs.Seqs[i].Num < start {
			start = seqs.Seqs[i].Num
		}
		for _, v := range rollbackSeq {
			if seqs.Seqs[i].Num == v {
				log.Error("rollbackSeq", "seq", seqs.Seqs[i].Num)
				count--
				continue dealRollback
			}
		}
		if seqs.Seqs[i].Seq.Type != db.SeqTypeAdd && seqs.Seqs[i].Seq.Type != db.SeqTypeDel {
			log.Error("parseSeqs seq op not support", "seq", seqs.Seqs[i].Num, "seqOp", seqs.Seqs[i].Seq.Type)
			err = errors.New("bad seq type")
			return
		}
		if len(seqs.Seqs[i].Detail.Block.Txs) != len(seqs.Seqs[i].Detail.Receipts) {
			log.Error("parseSeqs tx size != reciepti size", "seq", seqs.Seqs[i].Num,
				"txs", len(seqs.Seqs[i].Detail.Block.Txs), "receipts", len(seqs.Seqs[i].Detail.Receipts))
			err = errors.New("bad tx/receipt size")
			return
		}
		seqsOrder[seqs.Seqs[i].Num] = seqs.Seqs[i]
	}

	if len(seqsOrder) != count {
		err = errors.New("dup seq")
		return
	}

dealRollback2:
	for i := 0; i < totalCount; i++ {
		if _, ok := seqsOrder[int64(i)+start]; !ok {
			for _, v := range rollbackSeq {
				if int64(i)+start == v {
					continue dealRollback2
				}
			}

			err = errors.New("seq not continuous")
			break
		}
	}
	return
}
