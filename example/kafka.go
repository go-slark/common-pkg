package example

import (
	"github.com/Shopify/sarama"
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
		// TODO
		sess.MarkMessage(msg, "")
		sess.Commit()
	}
	return nil
}
