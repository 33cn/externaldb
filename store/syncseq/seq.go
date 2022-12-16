package syncseq

import (
	"encoding/json"
	"errors"
	"fmt"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/Shopify/sarama"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/mq/kafka"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/store"
	"github.com/33cn/externaldb/util"
)

var (
	log   = l.New("module", "syncSeq")
	mqMsg *sarama.ConsumerMessage
)

type seqNumStore struct {
	seqNumClient escli.ESClient
}

type esSeqStore struct {
	esSeqClient escli.ESClient
	bulk        bool
}

type mqSeqStore struct {
	mqSeqClient *kafka.Kafka
	topic       string
}

// NewEsSeqStore  SeqStore
func NewSeqStore(cfg *proto.ConfigNew) (store.SeqNumStore, store.SeqStore, error) {
	client, err := escli.NewESLongConnect(cfg.SyncEs.Host, cfg.SyncEs.Prefix, cfg.EsVersion, cfg.SyncEs.User, cfg.SyncEs.Pwd)
	if err != nil {
		return nil, nil, err
	}
	seqNumClient := &seqNumStore{
		seqNumClient: client,
	}

	if cfg.Dbtype == "mq" {
		kafka, err := kafka.ConnMQ(kafka.PubConn, cfg.Kafka.Host, "")
		if err != nil {
			return seqNumClient, nil, err
		}
		return seqNumClient, &mqSeqStore{
			mqSeqClient: kafka,
			topic:       cfg.Kafka.Topic,
		}, nil
	} else if cfg.Dbtype == "es" {
		err = util.InitIndex(client, block.StatusDB, block.StatusDB, block.Mapping)
		if err != nil {
			return seqNumClient, nil, err
		}
		return seqNumClient, &esSeqStore{
			esSeqClient: client,
			bulk:        cfg.SyncEs.Bulk,
		}, nil
	}
	l.Error("cfg.Dbtype set err")
	return seqNumClient, nil, errors.New("cfg.Dbtype set err")
}

// NewEsSeqStore  SeqStore
func NewGetSeq(cfg *proto.ConfigNew) (store.SeqNumStore, store.SeqStore, error) {
	client, err := escli.NewESLongConnect(cfg.SyncEs.Host, cfg.SyncEs.Prefix, cfg.EsVersion, cfg.SyncEs.User, cfg.SyncEs.Pwd)
	if err != nil {
		return nil, nil, err
	}
	seqNumClient := &seqNumStore{
		seqNumClient: client,
	}
	if cfg.Dbtype == "mq" {
		kafka, err := kafka.ConnMQ(kafka.SubConn, cfg.Kafka.Host, cfg.Kafka.Group)
		if err != nil {
			return seqNumClient, nil, err
		}
		consumer, err := kafka.AddConsumer(cfg.Kafka.Topic)
		if err != nil {
			return seqNumClient, nil, err
		}
		kafka.Sub = consumer
		return seqNumClient, &mqSeqStore{
			mqSeqClient: kafka,
		}, nil
	} else if cfg.Dbtype == "es" {
		return seqNumClient, &esSeqStore{
			esSeqClient: client,
		}, nil
	}
	return seqNumClient, nil, errors.New("cfg.Dbtype set err")
}

// LastSeq LastSeq
func (s *seqNumStore) LastSeq() (*store.SeqNum, error) {
	num, err := util.LastSyncSeq(s.seqNumClient, block.SyncSeq)
	if err != nil {
		return nil, err
	}
	return &store.SeqNum{Number: num}, nil
}

// UpdateLastSeq 更新最后同步的seq
func (s *seqNumStore) UpdateLastSeq(seq db.Record) error {
	err := s.seqNumClient.Update(seq.Index(), seq.Type(), seq.ID(), string(seq.Value()))
	log.Debug("save", "ID", seq.Key(), "value", string(seq.Value()))
	if err != nil {
		log.Error("save", "ID", seq.Key(), "err", err)
		return err
	}
	return nil
}

// SaveSeqs 输出同步区块信息到ES
func (s *esSeqStore) SaveSeqs(blockItems []db.Record) error {
	if len(blockItems) == 0 {
		return nil
	}
	if s.bulk {
		var rs []db.Record
		for i, v := range blockItems {
			rs = append(rs, v)
			log.Debug("save", "op", v.OpType(), "idx", i, "ID", v.Key(), "v", string(v.Value()))
		}
		return s.esSeqClient.BulkUpdate(rs)
	}

	for i, v := range blockItems {
		err := s.esSeqClient.Update(v.Index(), v.Type(), v.ID(), string(v.Value()))
		log.Debug("save", "idx", i, "ID", v.Key(), "value", string(v.Value()))
		if err != nil {
			log.Error("save", "idx", i, "ID", v.Key(), "err", err)
			return err
		}
	}
	return nil
}

// GetSeq get block from es
func (s *esSeqStore) GetSeq(seqNum int64) (*block.Seq, error) {
	id := fmt.Sprintf("%d", seqNum)
	result, err := s.esSeqClient.Get(block.StatusDB, block.StatusDB, id)
	if err != nil {
		return nil, err
	}
	var seq block.Seq
	err = json.Unmarshal([]byte(*result), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

// CommitSeqAck continue next seq
func (s *esSeqStore) CommitSeqAck(seqNum int64) int64 {
	return seqNum + 1
}

// SaveSeqs 输出同步区块信息到MQ
func (s *mqSeqStore) SaveSeqs(blockItems []db.Record) error {
	if len(blockItems) == 0 {
		return nil
	}
	for _, v := range blockItems {
		pid, offset, err := s.mqSeqClient.Publish(s.topic, v.Value(), v.Key())
		if err != nil {
			log.Error("pub failed", "err", err, "ID", v.Key(), "value.len", len(v.Value()))
			return err
		}
		log.Debug("publish msg success", "ID", v.Key(), "pid", pid, "offset", offset, "value.len", len(v.Value()))
	}
	return nil
}

// GetSeq get block from mq
func (s *mqSeqStore) GetSeq(seqNum int64) (*block.Seq, error) {
	msg, err := s.mqSeqClient.Subscribe(s.mqSeqClient.Sub)
	if msg == nil && err != nil {
		return nil, err
	} else if msg == nil && err == nil {
		return nil, nil
	}
	if msg != nil && err == nil {
		var seq block.Seq
		err = json.Unmarshal(msg.Value, &seq)
		if err != nil {
			return nil, err
		}
		mqMsg = msg
		return &seq, nil
	}
	return nil, nil
}

// CommitSeqAck commit ack to mq
func (s *mqSeqStore) CommitSeqAck(seqNum int64) int64 {
	s.mqSeqClient.Ack(mqMsg)
	return mqMsg.Offset + 1
}
