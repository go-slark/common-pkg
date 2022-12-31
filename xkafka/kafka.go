package xkafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xutils"
	"time"
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
	Brokers       []string `mapstructure:"brokers"`
	Retry         int      `mapstructure:"retry"`
	Ack           int16    `mapstructure:"ack"`
	ReturnSuccess bool     `mapstructure:"return_success"`
	ReturnErrors  bool     `mapstructure:"return_errors"`
}

type ConsumerGroupConf struct {
	Brokers      []string `mapstructure:"brokers"`
	GroupID      string   `mapstructure:"group_id"`
	Topics       []string `mapstructure:"topics"`
	Initial      int64    `mapstructure:"initial"`
	CommitEnable bool     `mapstructure:"commit_enable"`
	ReturnErrors bool     `mapstructure:"return_errors"`
}

type KafkaConf struct {
	Producer      *ProducerConf      `mapstructure:"producer"`
	ConsumerGroup *ConsumerGroupConf `mapstructure:"consumer_group"`
}

func (kp *KafkaProducer) Close() {
	_ = kp.SyncProducer.Close()
	kp.AsyncClose()
}

func (kp *KafkaProducer) SyncSend(ctx context.Context, topic, key string, msg []byte) error {
	traceID, ok := ctx.Value(xutils.TraceID).(string)
	if !ok {
		traceID = xutils.BuildRequestID()
	}
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte(xutils.TraceID),
				Value: []byte(traceID),
			},
		},
	}

	_, _, err := kp.SyncProducer.SendMessage(pm)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (kp *KafkaProducer) AsyncSend(ctx context.Context, topic, key string, msg []byte) error {
	traceID, ok := ctx.Value(xutils.TraceID).(string)
	if !ok {
		traceID = xutils.BuildRequestID()
	}
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte(xutils.TraceID),
				Value: []byte(traceID),
			},
		},
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

func InitKafkaProducer(conf *ProducerConf) *KafkaProducer {
	return &KafkaProducer{
		SyncProducer:  newSyncProducer(conf),
		AsyncProducer: newAsyncProducer(conf),
	}
}

func newSyncProducer(conf *ProducerConf) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(conf.Ack) // WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Retry
	config.Producer.Return.Successes = conf.ReturnSuccess // true
	config.Producer.Return.Errors = conf.ReturnErrors     // true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	producer, err := sarama.NewSyncProducer(conf.Brokers, config)
	if err != nil {
		panic(err)
	}
	return producer
}

func newAsyncProducer(conf *ProducerConf) sarama.AsyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(conf.Ack) // WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Retry
	config.Producer.Return.Successes = conf.ReturnSuccess // true
	config.Producer.Return.Errors = conf.ReturnErrors     // true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	producer, err := sarama.NewAsyncProducer(conf.Brokers, config)
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

func InitKafkaConsumer(conf *ConsumerGroupConf) *KafkaConsumerGroup {
	return &KafkaConsumerGroup{
		ConsumerGroup: newConsumerGroup(conf),
		Topics:        conf.Topics,
	}
}

func newConsumerGroup(conf *ConsumerGroupConf) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = conf.Initial                // sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = conf.CommitEnable // false
	config.Consumer.Return.Errors = conf.ReturnErrors             // true
	if err := config.Validate(); err != nil {
		panic(err)
	}

	consumerGroup, err := sarama.NewConsumerGroup(conf.Brokers, conf.GroupID, config)
	if err != nil {
		panic(err)
	}

	return consumerGroup
}

func (kc *KafkaConsumerGroup) Consume() {
	for {
		err := kc.ConsumerGroup.Consume(context.TODO(), kc.Topics, kc.ConsumerGroupHandler)
		if err != nil {
			logrus.Warnf("consumer group consume fail, err:%+v\n", err)
		}
		time.Sleep(time.Second)
	}
}

func (kc *KafkaConsumerGroup) Start() error {
	kc.Consume()
	return nil
}

func (kc *KafkaConsumerGroup) Stop(_ context.Context) error {
	return kc.Close()
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
