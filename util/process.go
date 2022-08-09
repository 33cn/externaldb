// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"time"

	"github.com/33cn/chain33/types"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/store"
)

var (
	currentSeqN int64
	syncSeqN    int64
)

// ModuleConvert 具体模块填上这个结构提， 由 Proc  控制流程
type ModuleConvert struct {
	Title    string
	Name     string
	StartSeq int64
	// 强制或建议从StartSeq 开始同步， 强制为了测试， 建议是因为之前seq没有需要处理的交易
	ForceSeq    bool
	WriteDB     escli.ESClient
	SeqStore    store.SeqStore
	SeqNumStore store.SeqNumStore
	AppConvert
}

// AppConvert for application
// 随着需求增加，需要各种各样的解析。由app 提供， 可以更灵活的实现
type AppConvert interface {
	ConvertBlock(blockSeq *block.Seq, detail *types.BlockDetail) ([]db.Record, error)
	RecoverStats(client escli.ESClient, lastSeq int64) error
}

// BlockProc deal block
func (mod *ModuleConvert) BlockProc() {
	syncSeq, err := mod.SeqNumStore.LastSeq()
	if err != nil {
		time.Sleep(1 * time.Second)
		log.Error("BlockProc LastSyncSeq1", "err", err, "module", mod.Name)
		return
	}
	syncSeqNum := syncSeq.Number
	currentSeqNum, err := LastSyncSeq(mod.WriteDB, mod.Name)
	if err != nil {
		time.Sleep(1 * time.Second)
		log.Error("BlockProc LastSyncSeq2", "err", err, "module", mod.Name)
		return
	}
	if currentSeqNum < mod.StartSeq {
		currentSeqNum = mod.StartSeq
	} else {
		currentSeqNum++
	}
	err = mod.RecoverStats(mod.WriteDB, currentSeqNum)
	if err != nil {
		time.Sleep(1 * time.Second)
		log.Error("BlockProc RecoverStats", "err", err, "seq", currentSeqNum, "module", mod.Name)
		return
	}

	//若seq记录未发生变化则不重复打印日志
	if currentSeqN != currentSeqNum || syncSeqN != syncSeqNum {
		log.Debug("dealBlock", "cur", currentSeqNum, "sync", syncSeqNum)
		currentSeqN = currentSeqNum
		syncSeqN = syncSeqNum
	}

	for currentSeqNum <= syncSeqNum {
		if ConvertServerStatus.Closed() {
			return
		}
		blockSeq, err := mod.SeqStore.GetSeq(currentSeqNum)
		if err != nil {
			// 在已经同步到的高度, 却没有找到对应的seq-block,
			// es中没有事务, 可能是 seq先更新, seq-block后完成索引导致
			if err == db.ErrDBNotFound {
				if currentSeqNum < syncSeqNum {
					currentSeqNum++
					continue
				}
				log.Error("Position not read, again", "seq", currentSeqNum, "err", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			time.Sleep(1 * time.Second)
			log.Error("BlockProc GetSyncBlock", "err", err, "seq", currentSeqNum, "module", mod.Name)
			continue
		} else if blockSeq == nil && err == nil {
			continue
		}

		records, err := mod.dealBlock(blockSeq)
		if err != nil {
			log.Error("BlockProc", "err", err, "block", blockSeq, "module", mod.Name)
			time.Sleep(1 * time.Second)
			continue
		}

		err = SaveToES(mod.WriteDB, records)
		if err != nil {
			log.Error("BlockProc", "err", err, "block", blockSeq, "module", mod.Name, "op", "save")
			time.Sleep(1 * time.Second)
			continue
		}
		currentSeqNum = mod.SeqStore.CommitSeqAck(currentSeqNum)
	}
}

func (mod *ModuleConvert) dealBlock(blockSeq *block.Seq) ([]db.Record, error) {
	var detail types.BlockDetail
	err := types.Decode(blockSeq.BlockDetail, &detail)
	if err != nil {
		return nil, err
	}
	log.Debug("dealBlock", "block", detail.Block.Height, "tx", len(detail.Block.Txs), "receipt", len(detail.Receipts))
	// TODO 跳过处理， 后续离线用脚本添加数据
	if len(detail.Block.Txs) != len(detail.Receipts) {
		log.Error("dealBlock size not match， skip", "block", detail.Block.Height, "tx", len(detail.Block.Txs), "receipt", len(detail.Receipts))
		return nil, nil
	}

	records, err := mod.ConvertBlock(blockSeq, &detail)

	lastRecord := NewLastRecord(mod.Name, int64(blockSeq.SyncSeq))
	log.Info("newLastRecord", "seq", blockSeq.SyncSeq)
	records = append(records, lastRecord)

	// 如果是回滚操作的话， 所有的状态都需要逆序处理
	if blockSeq.Type == db.SeqTypeDel {
		records = reverse(records)
	}
	return records, err
}

// SingleDealBlock 处理区块时只处理当前区块，不添加lastRecord
func (mod *ModuleConvert) SingleDealBlock(blockSeq *block.Seq) ([]db.Record, error) {
	var detail types.BlockDetail
	err := types.Decode(blockSeq.BlockDetail, &detail)
	if err != nil {
		return nil, err
	}
	log.Debug("SingleDealBlock", "block", detail.Block.Height, "tx", len(detail.Block.Txs), "receipt", len(detail.Receipts))
	// TODO 跳过处理， 后续离线用脚本添加数据
	if len(detail.Block.Txs) != len(detail.Receipts) {
		log.Error("SingleDealBlock size not match， skip", "block", detail.Block.Height, "tx", len(detail.Block.Txs), "receipt", len(detail.Receipts))
		return nil, nil
	}
	records, err := mod.ConvertBlock(blockSeq, &detail)
	// 如果是回滚操作的话， 所有的状态都需要96-逆序处理
	if blockSeq.Type == db.SeqTypeDel {
		records = reverse(records)
	}
	return records, err
}

// Reverse reverse records
// https://github.com/golang/go/wiki/SliceTricks#reversing
func reverse(a []db.Record) []db.Record {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}

type convertServerStatus struct {
	closed bool
}

var ConvertServerStatus = &convertServerStatus{}

func (c *convertServerStatus) CloseServer() {
	c.closed = true
}

func (c *convertServerStatus) Closed() bool {
	return c.closed
}
