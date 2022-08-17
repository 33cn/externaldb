package kafka

import (
	"errors"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type MsgCb func(m *Msg) error

type Msg struct {
	body   *sarama.ConsumerMessage
	handle *Kafka
}

//发布消息
func (k *Kafka) Publish(topic string, body []byte, key string) (partition int32, offset int64, err error) {
	m := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(body),
		Timestamp: time.Now(),
	}
	partition, offset, err = k.Pub.SendMessage(m)
	return partition, offset, err
}

func (k *Kafka) AddConsumer(topic string) (consumer *cluster.Consumer, err error) {
	consumer, err = cluster.NewConsumer(k.Brokers, k.Params.Group, []string{topic}, k.SubConfig)
	if err != nil {
		return nil, err
	}
	k.Sub = consumer
	return consumer, nil
}

//订阅消息
func (k *Kafka) Subscribe(consumer *cluster.Consumer) (message *sarama.ConsumerMessage, err error) {
	select {
	case err = <-consumer.Errors():
		klog.Error("consumer err", "err", err)
		return nil, err
	case n := <-consumer.Notifications():
		klog.Info("consumer rebalanced", "notify", n)
		return nil, nil
	case msg, ok := <-consumer.Messages():
		if !ok {
			klog.Info("msg queue closed", "brokers", strings.Join(k.Brokers, ","),
				"group", k.Params.Group)
			return nil, errors.New("msg queue closed")
		}
		klog.Info("receive msg success", "ID", string(msg.Key), "pid", msg.Partition, "offset", msg.Offset)
		if k.Params.AutoAck {
			k.Sub.MarkOffset(msg, "") // 提交offset
		}
		return msg, nil
	}
}

//Ack 手动提交ack
func (k *Kafka) Ack(msg *sarama.ConsumerMessage) {
	k.Sub.MarkOffset(msg, "")
}

//重置offset，让同一消费组的消费者重新消费
func (k *Kafka) ResetOffset(topic string) error {
	client, om, _ := k.newOffsetManagerFromClient()
	defer om.Close()

	partitionIds, _ := client.Partitions(topic)
	for _, partitionID := range partitionIds {
		pm, _ := om.ManagePartition(topic, partitionID)
		pm.ResetOffset(0, "")
		klog.Info("reset success", "pid", partitionID)
	}

	sarama.NewConsumer(k.Brokers, k.ConnConfig)

	return nil
}

//重置offset，指定分区和偏移量
func (k *Kafka) ResetOffsetAppoint(topic string, partition int32, offset int64) error {
	_, om, _ := k.newOffsetManagerFromClient()
	defer om.Close()

	pm, err := om.ManagePartition(topic, partition)
	if err != nil {
		klog.Error("ManagePartition failed", "err", err)
		return err
	}

	pm.ResetOffset(offset, "")
	klog.Info("reset success")

	sarama.NewConsumer(k.Brokers, k.ConnConfig)

	return nil
}

func (k *Kafka) Close() {
	if k.Pub != nil {
		k.Pub.Close()
	}
	if k.Sub != nil {
		k.Sub.Close()
	}
}

func (m *Msg) GetBody() []byte {
	return m.body.Value
}

func (m *Msg) GetID() string {
	return string(m.body.Key)
}

func (m *Msg) GetPid() int32 {
	return m.body.Partition
}

func (m *Msg) GetOffset() int64 {
	return m.body.Offset
}

func (m *Msg) GetTopic() string {
	return m.body.Topic
}

func (m *Msg) Ack() error {
	m.handle.Sub.MarkOffset(m.body, "")
	return nil
}

func (k *Kafka) newOffsetManagerFromClient() (sarama.Client, sarama.OffsetManager, error) {
	client, err := sarama.NewClient(k.Brokers, k.ConnConfig)
	if err != nil {
		klog.Error("sarama.NewClient failed", "err", err)
		return nil, nil, err
	}

	om, err := sarama.NewOffsetManagerFromClient("group-1", client)
	if err != nil {
		klog.Error("sarama.NewOffsetManagerFromClient failed", "err", err)
		return nil, nil, err
	}
	return client, om, nil
}
