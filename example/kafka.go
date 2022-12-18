package example

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	mq "github.com/smallfish-root/common-pkg/xkafka"
)

type Kafka struct {
	*mq.KafkaClient
}

func NewKafka() *Kafka {
	k := &Kafka{
		KafkaClient: &mq.KafkaClient{
			KafkaProducer:      mq.GetKafkaProducer(),
			KafkaConsumerGroup: mq.GetKafkaConsumerGroup(),
		},
	}
	k.ConsumerGroupHandler = k
	return k
}

func (*Kafka) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*Kafka) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (k *Kafka) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := context.Background()
		if len(msg.Headers) != 0 && msg.Headers[0] != nil {
			ctx = context.WithValue(context.Background(), string(msg.Headers[0].Key), string(msg.Headers[0].Value))
		}

		logrus.WithContext(ctx) // example todo

		// TODO
		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return nil
}
