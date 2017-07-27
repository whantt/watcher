package kafka

import (
	l "log"
	"os"
	"time"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/Shopify/sarama"
	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/harvester"
	"github.com/dearcode/watcher/meta"
)

var (
	kh = kafkHarvester{}
)

type kafkHarvester struct {
	consumer *cluster.Consumer
}

func init() {
	harvester.Register("kafka", &kh)
}

func (kh *kafkHarvester) Init(hc config.HarvesterConfig) error {
	cc := cluster.NewConfig()
	cc.ClientID = hc.ClientID
	cc.Consumer.Offsets.Initial = sarama.OffsetOldest
	cc.Consumer.MaxWaitTime = time.Second
	cc.Consumer.Return.Errors = true
	cc.Group.Return.Notifications = true

	sarama.Logger = l.New(os.Stdout, "", l.Lshortfile|l.LstdFlags)

	consumer, err := cluster.NewConsumer(hc.Brokers, hc.Group, hc.Topics, cc)
	if err != nil {
		return errors.Annotatef(err, "NewConsumer:%+v", hc)
	}

	kh.consumer = consumer

	return nil
}

func (kh *kafkHarvester) Start() <-chan *meta.Message {
	c := make(chan *meta.Message)

	go kh.run(c)

	return c
}

func (kh *kafkHarvester) Stop() {
	kh.consumer.Close()
}

func (kh *kafkHarvester) run(c chan *meta.Message) {
	for {
		select {
		case msg, ok := <-kh.consumer.Messages():
			if !ok {
				log.Errorf("consumer Messages error")
				return
			}
			c <- meta.NewMessage(msg.Topic, string(msg.Value))
			log.Debugf("topic:%v, offset:%v, value:%v", msg.Topic, msg.Offset, string(msg.Value))
			kh.consumer.MarkOffset(msg, "")
		case err, ok := <-kh.consumer.Errors():
			log.Errorf("consumer Error:%v, ok:%v", err, ok)
			return
		case ntf, ok := <-kh.consumer.Notifications():
			if ok {
				log.Infof("Rebalanced: %#v", ntf)
			}
		}
	}
}
