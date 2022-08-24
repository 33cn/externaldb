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

var LastSyncSeqCache = &lastSyncSeqCache{number: 0}

func InitLastSyncSeqCache(client escli.ESClient, id string, startSeq int64) error {
	currentSeqNum, err := LastSyncSeq(client, id)
	if err != nil {
		log.Error("InitLastSyncSeqCache failed", "err", err, "module", id)
		return err
	}
	if currentSeqNum == -1 {
		log.Info("last_seq 从ES获取失败，自动使用配置文件sync.startSeq")
	}
	if currentSeqNum < startSeq {
		currentSeqNum = startSeq
	}
	log.Info("last_seq 处理成功", "当前last_seq ", currentSeqNum)
	// 前面步骤，底层包在查询ES的last_seq不存在时，会输出ERROR错误日志，为了方便查看，这里也在ERROR的位置输出last_seq 处理成功, 便于理解
	log.Error("ES last_seq 自动修复处理完毕", "当前last_seq ", currentSeqNum)
	return LastSyncSeqCache.SetNumber(currentSeqNum)
}

func (s *lastSyncSeqCache) SetNumber(n int64) error {
	if n < 0 {
		return errors.New("LastSyncSeq Number has to be greater than 0")
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
