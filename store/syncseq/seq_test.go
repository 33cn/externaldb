package syncseq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/33cn/externaldb/mq/kafka"
)

func TestGetSeq(t *testing.T) {
	assert := assert.New(t)
	var err error
	kafka, _ := kafka.ConnMQ("Subscribe", "127.0.0.1:9092", "group-1")

	mqStore := &mqSeqStore{kafka, "sync3"}
	for i := 0; i < 10; i++ {
		_, err = mqStore.GetSeq(1)
	}
	assert.Nil(err)
}
