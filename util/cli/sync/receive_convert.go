package sync

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/util/cli/convert"
	"github.com/pkg/errors"

	// 以下包导入不显示使用，主要使用里面的init函数，初始化一些配置，复制来源 cmd/convert/main.go 里面的导入逻辑
	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/externaldb/db/account"
	_ "github.com/33cn/externaldb/db/coins"
	_ "github.com/33cn/externaldb/db/evm"
	_ "github.com/33cn/externaldb/db/evm/erc1155"
	_ "github.com/33cn/externaldb/db/evm/erc721"
	_ "github.com/33cn/externaldb/db/evm/nft"
	_ "github.com/33cn/externaldb/db/evmxgo"
	_ "github.com/33cn/externaldb/db/filepart"
	_ "github.com/33cn/externaldb/db/filesummary"
	_ "github.com/33cn/externaldb/db/multisig"
	_ "github.com/33cn/externaldb/db/proof"
	_ "github.com/33cn/externaldb/db/proof_config"
	_ "github.com/33cn/externaldb/db/ticket"
	_ "github.com/33cn/externaldb/db/token"
	_ "github.com/33cn/externaldb/db/trade"
	_ "github.com/33cn/externaldb/db/unfreeze"
	_ "github.com/33cn/externaldb/stat/block"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

type ReceiverConvert interface {
	RecoverStats() error
	Register() error
	ReceiveLoop() // 内部启动http, http请求内部直接解析请求数据，写入rpc es查询库

	// AnalysisBlock 在http请求里面，解析区块    CheckBlock 检查区块   ConvertBlock 转换处理区块数据
	//AnalysisBlock()
	//CheckBlock()
	//ConvertBlock()
}

func CreateReceiverConvert(cfg *proto.ConfigNew, EsWrite escli.ESClient) (ReceiverConvert, error) {
	err := checkPushFormat(cfg.Sync.PushFormat)
	if err != nil {
		log.Error("checkPushFormat failed", "err", err.Error())
		return nil, err
	}

	p := pusher{
		pushServer: cfg.Chain.Host,
		name:       cfg.Sync.PushName,
		url:        "http://" + cfg.Sync.PushHost,
		encode:     cfg.Sync.PushFormat,

		height:    cfg.Sync.StartHeight,
		seq:       cfg.Sync.StartSeq,
		blockHash: cfg.Sync.StartBlockHash,
	}

	// 兼容老版本
	if p.height == 0 && p.blockHash == "" {
		p.seq = 0
	}

	convert.EsWrite = EsWrite
	mod := &util.ModuleConvert{
		Name:        cfg.Convert.AppName,
		SeqNumStore: nil,
		SeqStore:    nil,
		WriteDB:     EsWrite,
		StartSeq:    cfg.Convert.StartSeq,
		ForceSeq:    false,
		AppConvert:  convert.NewApp(cfg),
	}

	return &receiverConvert{
		p:        &p,
		bindAddr: cfg.Sync.PushBind,
		mod:      mod,
		chain:    cfg.Chain,
	}, nil
}

type receiverConvert struct {
	p        *pusher
	bindAddr string
	mod      *util.ModuleConvert
	// pushVersion int32
	chain *proto.Chain33
}

func (r *receiverConvert) RecoverStats() error {
	err := r.mod.RecoverStats(r.mod.WriteDB, util.LastSyncSeqCache.GetNumber())
	if err != nil {
		log.Error("BlockProc RecoverStats", "err", err, "seq", util.LastSyncSeqCache.GetNumber(), "module", r.mod.Name)
	}
	return err
}

func (r *receiverConvert) Register() error {
	_, err := NewReceiver(r.p, r.bindAddr)
	return err
}

func (r *receiverConvert) ReceiveLoop() {
	handler := func(req []byte) error {
		return handleConvertRequest(req, r.p.encode, r.mod, r.chain)
	}
	startHTTPService(r.bindAddr, "*", handler)
}

func handleConvertRequest(body []byte, format string, mod *util.ModuleConvert, chain *proto.Chain33) error {
	beg := types.Now()
	defer func() {
		log.Info("HandleConvertRequest", "total cost", types.Since(beg))
	}()

	// 解析请求，转为结构体
	var req types.BlockSeqs
	var err error
	if format == "json" {
		err = types.JSONToPB(body, &req)
	} else {
		err = types.Decode(body, &req)
	}
	if err != nil {
		log.Error("handleRequest", "JSONToPB", err, "req", &req)
		return err
	}

	// 检查数据是否正确
	count, start, seqsMap, err := parseSeqs(&req)
	if err != nil {
		log.Error("deal request", "parseSeqs", err)
		return err
	}
	log.Info("parseSeqs", "seqsMap", seqsMap)
	log.Info("parseSeqs", "cost", types.Since(beg), "count", count, "start", start)

	currentSeqNum := util.LastSyncSeqCache.GetNumber()
	// 在app 端保存成功， 但回复ok时，程序挂掉, 记录日志
	log.Info("GetNumber", "current_seq", currentSeqNum)
	if start+int64(count) <= currentSeqNum {
		return nil
	}

	// 一般在配置有误时发生， 在同一个节点上配置相同的推送名
	if start > currentSeqNum+1 {
		log.Error("GetNumber", "seq", start, "current_seq", currentSeqNum)
		return errors.New("bad seq")
	}

	// 将所有区块信息取出，并处理
	number := currentSeqNum
	bulkRecords := make([]db.Record, 0, count*2)

	for i := 0; i < count; i++ {
		if start+int64(i) <= currentSeqNum {
			log.Info("for 循环", "start+int64(i)", start+int64(i), "currentSeqNum", currentSeqNum)
			continue
		}

		number++

		seq, ok := seqsMap[int64(i)+start]
		if !ok {
			log.Info("for 循环", "seq", seq, "ok", ok)
			continue
		}
		blockSeq := &block.Seq{
			SyncSeq:     int(number),
			From:        chain.Host, // cfg.Chain.Host
			Number:      int(seq.Num),
			Hash:        common.ToHex(seq.Seq.Hash),
			Type:        int(seq.Seq.Type),
			BlockDetail: types.Encode(seq.Detail),
		}

		// 解析处理数据
		records, err := mod.SingleDealBlock(blockSeq)
		if err != nil {
			log.Error("BlockProc", "err", err, "block", blockSeq, "module", mod.Name)
			// 数据解析错误，可能是数据格式问题，为了后续区块能够接收处理，发生数据解析错误后，跳过此条信息
			continue
		}
		bulkRecords = append(bulkRecords, records...)
	}

	lastSeq := util.NewLastRecord(db.LastSeqDB, number)
	bulkRecords = append(bulkRecords, lastSeq)
	log.Info("deal request over", "cost", types.Since(beg), "number", number, "bulkRecords", bulkRecords)

	// 存入ES
	err = util.SaveToESSelectBulk(mod.WriteDB, bulkRecords, util.ConvertEsBulk)
	if err != nil {
		log.Error("SaveToESSelectBulk", "err", err, "bulkRecords", bulkRecords, "module", mod.Name, "op", "save")
		return err
	}
	err = util.LastSyncSeqCache.SetNumber(number)

	log.Info("response", "err", err)
	return err
}
