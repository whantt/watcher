package kafka

import (
	"fmt"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/zssky/log"

	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/harvester"
	"github.com/dearcode/tracker/meta"
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
	cc.Consumer.Return.Errors = true
	cc.Group.Return.Notifications = true

	consumer, err := cluster.NewConsumer(hc.Brokers, hc.Group, hc.Topics, cc)
	if err != nil {
		return fmt.Errorf("Error NewConsumer: %v", err)
	}

	kh.consumer = consumer

	return nil
}

func (kh *kafkHarvester) Start() <-chan *meta.Message {
	c := make(chan *meta.Message)

	go kh.run(c)

	return c
}

func (kh *kafkHarvester) run(c chan *meta.Message) {
	for {
		select {
		case msg, ok := <-kh.consumer.Messages():
			if !ok {
				log.Errorf("consumer Messages error")
				continue
			}
			c <- meta.NewMessage(msg.Topic, string(msg.Value))
			kh.consumer.MarkOffset(msg, "")
		case err, ok := <-kh.consumer.Errors():
			if ok {
				log.Errorf("consumer Error: %v", err)
			}
		case ntf, ok := <-kh.consumer.Notifications():
			if ok {
				log.Infof("Rebalanced: %v", ntf)
			}
		}
	}
}
