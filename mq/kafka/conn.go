package kafka

import (
	"strings"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var klog = log15.New("module", "mq/kafka")

const (
	PubConn = "Publish"
	SubConn = "Subscribe"
)

type ConnParams struct {
	Addr    string
	NeedPub bool // for pub
	NeedSub bool // for sub
	AutoAck bool // for sub
	Group   string
}

type Kafka struct {
	Params     *ConnParams
	Brokers    []string
	ConnConfig *sarama.Config
	SubConfig  *cluster.Config
	Pub        sarama.SyncProducer
	Sub        *cluster.Consumer
}

//连接kafka服务
func ConnMQ(connType, addr, group string) (*Kafka, error) {
	var para ConnParams

	para.Addr = addr

	if connType == "Publish" {
		para.NeedPub = true
	} else if connType == "Subscribe" {
		para.NeedSub = true
		para.AutoAck = false
		para.Group = group
	}
	klog.Info("kafka para set success")

	kafka, _ := creator(&para)
	klog.Info("kafka connect success")

	return kafka, nil
}

//Kafka config的配置和producer的连接
func creator(p *ConnParams) (*Kafka, error) {
	k := new(Kafka)
	k.Params = p
	k.Brokers = strings.Split(p.Addr, ",")
	if p.NeedPub {
		kc := sarama.NewConfig()
		kc.Version = sarama.V2_0_0_0                       //当前kafka的版本
		kc.Producer.Compression = sarama.CompressionSnappy //将数据进行压缩传输，提高数据传输的效率
		kc.Producer.RequiredAcks = sarama.WaitForAll       //等待所有同步中的副本都确认消息
		kc.Producer.Retry.Max = 10                         //发送消息重试的次数
		kc.Producer.Return.Successes = true
		kc.Producer.MaxMessageBytes = 5242880 //传递大消息时需要的配置，默认为1M
		pub, err := sarama.NewSyncProducer(k.Brokers, kc)
		if err != nil {
			return nil, err
		}
		k.ConnConfig = kc
		k.Pub = pub
	} else if p.NeedSub {
		kc := cluster.NewConfig()
		kc.Consumer.Return.Errors = true
		kc.Consumer.Offsets.Initial = sarama.OffsetOldest
		kc.Group.Return.Notifications = true
		k.SubConfig = kc
	}
	return k, nil
}
