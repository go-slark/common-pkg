package mq

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

/*
[kafka]
    [kafka.producer]
       brokers = ["192.168.4.14:9092"]
       topic = "exam_msg_prod"
       retry = 5

    [kafka.consumer_group]
       brokers = ["192.168.4.14:9092"]
       group_id = "logic-server"
       topics = ["exam_prod"]

func InitKafka() {
	conf := &mq.KafkaConf{}
	err := viper.UnmarshalKey("kafka", conf, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "mapstructure"
	})
	if err != nil {
		panic(fmt.Sprintf("parse kafka config %v\n", err))
	}

	mq.InitKafkaProducer(conf)
    mq.InitKafkaConsumer(conf)
}

*/

type KafkaProducer struct {
	sarama.SyncProducer
	sarama.AsyncProducer
}

type ProducerConf struct {
	Brokers []string `mapstructure:"brokers"`
	Retry   int      `mapstructure:"retry"`
}

type ConsumerGroupConf struct {
	Brokers []string `mapstructure:"brokers"`
	GroupId string   `mapstructure:"group_id"`
	Topics  []string `mapstructure:"topics"`
}

type KafkaConf struct {
	Producer      *ProducerConf      `mapstructure:"producer"`
	ConsumerGroup *ConsumerGroupConf `mapstructure:"consumer_group"`
}

func (kp *KafkaProducer) Close() {
	_ = kp.SyncProducer.Close()
	kp.AsyncClose()
}

func (kp *KafkaProducer) SyncSend(topic, key string, msg []byte) error {
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
	}

	_, _, err := kp.SyncProducer.SendMessage(pm)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (kp *KafkaProducer) AsyncSend(topic, key string, msg []byte) error {
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
	}

	kp.AsyncProducer.Input() <- pm
	select {
	case <-kp.AsyncProducer.Successes():

	case err := <-kp.AsyncProducer.Errors():
		return errors.WithStack(err)

	default:

	}
	return nil
}

func InitKafkaProducer(conf *KafkaConf) {
	kafkaProducer = &KafkaProducer{
		SyncProducer:  newSyncProducer(conf),
		AsyncProducer: newAsyncProducer(conf),
	}
}

func newSyncProducer(conf *KafkaConf) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Producer.Retry
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	producer, err := sarama.NewSyncProducer(conf.Producer.Brokers, config)
	if err != nil {
		panic(err)
	}
	return producer
}

func newAsyncProducer(conf *KafkaConf) sarama.AsyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Producer.Retry
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	producer, err := sarama.NewAsyncProducer(conf.Producer.Brokers, config)
	if err != nil {
		panic(err)
	}

	return producer
}

type KafkaConsumerGroup struct {
	sarama.ConsumerGroup
	sarama.ConsumerGroupHandler
	Topics []string
}

func InitKafkaConsumer(conf *KafkaConf) {
	kafkaConsumerGroup = &KafkaConsumerGroup{
		ConsumerGroup: newConsumerGroup(conf),
		Topics:        conf.ConsumerGroup.Topics,
	}
}

func newConsumerGroup(conf *KafkaConf) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Return.Errors = true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	consumerGroup, err := sarama.NewConsumerGroup(conf.ConsumerGroup.Brokers, conf.ConsumerGroup.GroupId, config)
	if err != nil {
		panic(err)
	}

	return consumerGroup
}

func (kc *KafkaConsumerGroup) Consume() {
	for {
		err := kc.ConsumerGroup.Consume(context.TODO(), kc.Topics, kc.ConsumerGroupHandler)
		if err != nil {
			fmt.Printf("consumer group consume fail, err:%+v\n", err)
			break
		}
	}
}

func (kc *KafkaConsumerGroup) Close() {
	kc.Close()
}

var (
	kafkaProducer      *KafkaProducer
	kafkaConsumerGroup *KafkaConsumerGroup
)

func GetKafkaProducer() *KafkaProducer {
	return kafkaProducer
}

func GetKafkaConsumerGroup() *KafkaConsumerGroup {
	return kafkaConsumerGroup
}
