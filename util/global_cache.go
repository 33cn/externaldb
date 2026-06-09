package util

import (
	"github.com/33cn/externaldb/escli"
	"github.com/pkg/errors"
)

// 全局变量，变量持久化到es里面，但是涉及到需要频繁读写的变量，频繁的查es效率不高，并且es7对于实时查询的支持有限
// 基于目前是一个http服务接收链推送，链那边是顺序推送，并且一个请求结束之后才推送另一个，
// 因此设置全局变量，程序启动时加载到内存里，之后全局变量修改，先修改es, 再修改内存，日常的读首先取内存值，
// 以后如果有了redis 或者类似的高性能读写服务，全局缓存可以移到redis

type lastSyncSeqCache struct {
	number int64
}

var (
	LastOne  = int64(46612200) //  int64(46612097)
	FirstOne = int64(43380000) // 40049898

	//againLast  = int64(30000000)
	//againFirst = int64(20000000)
)

var LastSyncSeqCache = &lastSyncSeqCache{number: -1}

func InitLastSyncSeqCacheFixTool(client escli.ESClient, id string, startSeq int64) error {
	//currentSeqNum, err := LastSyncSeq(client, id)
	var err error
	_ = err
	currentSeqNum := FirstOne
	if err != nil {
		log.Error("InitLastSyncSeqCache failed", "err", err, "module", id)
		return err
	}
	if currentSeqNum == -1 {
		log.Info("last_seq 从ES获取失败，自动使用配置文件sync.startSeq")
	}
	if currentSeqNum < startSeq {
		currentSeqNum = startSeq - 1
	}
	log.Info("last_seq 处理成功", "当前last_seq ", currentSeqNum)
	// 前面步骤，底层包在查询ES的last_seq不存在时，会输出ERROR错误日志，为了方便查看，这里也在ERROR的位置输出last_seq 处理成功, 便于理解
	log.Error("ES last_seq 自动修复处理完毕", "当前last_seq ", currentSeqNum)
	return LastSyncSeqCache.SetNumber(currentSeqNum)
}

func InitLastSyncSeqCache(client escli.ESClient, id string, startSeq int64) error {
	var currentSeqNum int64
	if startSeq > 0 {
		// startSeq > 0 时直接使用配置值，跳过 ES 读取，避免启动时卡在 ES
		currentSeqNum = startSeq - 1
		log.Info("last_seq 使用配置值（跳过ES读取）", "startSeq", startSeq, "当前last_seq", currentSeqNum)
	} else {
		var err error
		currentSeqNum, err = LastSyncSeq(client, id)
		if err != nil {
			log.Error("InitLastSyncSeqCache failed", "err", err, "module", id)
			return err
		}
		if currentSeqNum == -1 {
			log.Info("last_seq 从ES获取失败，自动使用配置文件sync.startSeq")
		}
		if currentSeqNum < startSeq {
			currentSeqNum = startSeq - 1
		}
		log.Info("last_seq 处理成功", "当前last_seq ", currentSeqNum)
		// 前面步骤，底层包在查询ES的last_seq不存在时，会输出ERROR错误日志，为了方便查看，这里也在ERROR的位置输出last_seq 处理成功, 便于理解
		log.Error("ES last_seq 自动修复处理完毕", "当前last_seq ", currentSeqNum)
	}
	return LastSyncSeqCache.SetNumber(currentSeqNum)
}

func (s *lastSyncSeqCache) SetNumber(n int64) error {
	if n < -1 {
		return errors.New("LastSyncSeq Number must >= -1")
	}
	s.number = n
	return nil
}

func (s *lastSyncSeqCache) GetNumber() int64 {
	return s.number
}

// ConvertEsBulk 数据解析结果存储的ES 是否选择批量写入
var ConvertEsBulk bool

func InitConvertEsBulk(bulk bool) {
	ConvertEsBulk = bulk
}
