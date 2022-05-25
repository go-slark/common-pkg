package mq

type KafkaClient struct {
	*KafkaProducer
	*KafkaConsumerGroup
}

func (k *KafkaClient) AsyncProduce(topic, key string, msg []byte) error {
	return k.AsyncSend(topic, key, msg)
}

func (k *KafkaClient) SyncProduce(topic, key string, msg []byte) error {
	return k.SyncSend(topic, key, msg)
}

func (k *KafkaClient) Consume() error {
	go k.KafkaConsumerGroup.Consume()
	return nil
}

type queue interface {
	Produce(topic, key string, msg []byte) error
	Consume() error
}

func (k *KafkaClient) Produce(topic, key string, msg []byte) error {
	return k.AsyncSend(topic, key, msg)
}

var _ queue = &KafkaClient{}
