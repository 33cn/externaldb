package kafka

import (
	"encoding/json"
	"fmt"
	l "log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Student struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// func msgCb(msg *KafkaMsg) error {
// 	var student Student
// 	j++
// 	l.Printf("receive msg [%d] --> body:[%s] with ID [%s]\n", j, msg.GetBody(), msg.GetID())
//
// 	json.Unmarshal(msg.GetBody(), &student)
// 	//Sqlcli.ExecInsert("insert into student(id,name) values(?,?)", student.ID, student.Name)
// 	//Escli.Set(student, "student", "student", msg.GetID())
// 	return nil
// }

func testConnMQ(t *testing.T, connType string) *Kafka {
	assert := assert.New(t)
	k, err := ConnMQ(connType, "127.0.0.1:9093", "group-1")
	assert.Nil(err)
	return k
}

func TestConnMQ(t *testing.T) {
	testConnMQ(t, PubConn)
}

func TestPublic(t *testing.T) {
	assert := assert.New(t)
	k := testConnMQ(t, PubConn)
	var student Student
	var err error
	for i := 15; i < 16; i++ {
		go func(i int) {
			student.ID = i
			student.Name = fmt.Sprintf("t%d", i)
			body, _ := json.Marshal(student)
			key := strconv.Itoa(i)
			pid, offset, err := k.Publish("d", body, key)
			if err != nil {
				l.Fatal("pub err", err)
			}
			l.Printf("publish msg body:[%s] with ID [%s] - pid:[%d] offset:[%d]\n", body, key, pid, offset)
		}(i)
	}
	assert.Nil(err)
	select {}
}

func TestSubscribe(t *testing.T) {
	assert := assert.New(t)
	k := testConnMQ(t, SubConn)
	consumer, err := k.AddConsumer("d")
	k.Subscribe(consumer)
	assert.Nil(err)
	select {}
}
