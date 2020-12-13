package xkafka

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	sarama "gopkg.in/Shopify/sarama.v1"
	"time"
)

type ProducerConfig struct {
	Topic     string
	Partition int32
	Broker    []string //[]string{"IP:9092","IP:9092","IP:9092"}集群节点信息
}

func NewSyncProducer(pc *ProducerConfig) sarama.SyncProducer {
	if pc == nil {
		return nil
	}

	c := sarama.NewConfig()
	c.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	c.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	c.Producer.Return.Successes = true
	sp, err := sarama.NewSyncProducer(pc.Broker, c)
	if err != nil {
		logrus.Errorf("new kafka sync producer fail, conf:%+v, err:%+v", *pc, err)
		return nil
	}
	return sp
}

func SendMsgBySyncProducer(c *ProducerConfig, key string, msg interface{}) (sarama.SyncProducer, error) {
	p := NewSyncProducer(c)
	//defer p.Close()
	b, err := json.Marshal(msg)
	if err != nil {
		return p, errors.WithStack(err)
	}
	//partition决定topic按照什么样的方式分配到partition，随机还是轮循；key决定按照哪个string分配都partition
	m := &sarama.ProducerMessage{
		Key:       sarama.StringEncoder(key),
		Topic:     c.Topic,
		Partition: c.Partition, //sarama.NewRandomPartitioner
		Value:     sarama.ByteEncoder(b),
	}

	if _, _, err = p.SendMessage(m); err != nil {
		logrus.Errorf("send msg fail, msg:%+v, err:%+v", msg, err)
		return p, errors.WithStack(err)
	}
	return p, nil
}

func NewAsyncProducer(pc *ProducerConfig) sarama.AsyncProducer {
	if pc == nil {
		return nil
	}

	c := sarama.NewConfig()
	//等待服务器所有副本都保存成功后的响应
	c.Producer.RequiredAcks = sarama.WaitForAll
	//随机的分区类型
	c.Producer.Partitioner = sarama.NewRandomPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	c.Version = sarama.V0_11_0_0

	ap, err := sarama.NewAsyncProducer(pc.Broker, c)
	if err != nil {
		logrus.Errorf("new kafka async producer fail, conf:%+v, err:%+v", *pc, err)
		return nil
	}
	return ap
}

func SendMsgByAsyncProducer(c *ProducerConfig, key string, msg interface{}) (sarama.AsyncProducer, error) {
	p := NewAsyncProducer(c)
	//defer p.Close()
	b, err := json.Marshal(msg)
	if err != nil {
		return p, errors.WithStack(err)
	}
	//partition决定topic按照什么样的方式分配到partition，随机还是轮循；key决定按照哪个string分配都partition
	m := &sarama.ProducerMessage{
		Key:       sarama.StringEncoder(key),
		Topic:     c.Topic,
		Partition: c.Partition, //sarama.NewRandomPartitioner
		Value:     sarama.ByteEncoder(b),
	}

	p.Input() <- m
	select {
	case succ := <-p.Successes():
		logrus.Debugf("kafka async producer send msg success, succ:%+v, msg:%+v", succ, msg)
	case err = <-p.Errors():
		logrus.Errorf("kafka async producer send msg fail, err:%+v, msg:%+v", err, msg)
		return p, errors.WithStack(err)
	}

	return p, nil
}

type Config struct {
	Hosts       []string      `mapstructure:"hosts"`//broker
	Topics      []string      `mapstructure:"topics"`
	GroupID     string        `mapstructure:"group_id"`
	ClientID    string        `mapstructure:"client_id"`
	MaxWaitTime time.Duration `mapstructure:"max_wait_time"`
	Username    string        `mapstructure:"username"`
	Password    string        `mapstructure:"password"`
}

func NewConsumerGroup(config *Config) sarama.ConsumerGroup {
	c := sarama.NewConfig()
	c.Version = sarama.V1_0_0_0
	c.Consumer.Offsets.Initial = sarama.OffsetOldest
	c.Consumer.Return.Errors = true
	c.Consumer.MaxWaitTime = time.Duration(config.MaxWaitTime) * time.Second

	if config.Username != "" {
		c.Net.SASL.Enable = true
		c.Net.SASL.User = config.Username
		c.Net.SASL.Password = config.Password
	}

	c.ClientID = config.ClientID
	if err := c.Validate(); err != nil {
		return nil
	}

	//hosts == broker
	consumerGroup, err := sarama.NewConsumerGroup(config.Hosts, config.GroupID, c)
	if err != nil {
		return nil
	}
	return consumerGroup
}

func Consume(ctx context.Context, conf *Config, handler sarama.ConsumerGroupHandler) (sarama.ConsumerGroup, error) {
	consumer := NewConsumerGroup(conf)
	go func() {
		//监听consumer消费过程中出现的错误
		for err := range consumer.Errors() {
			logrus.Errorf("kafka consume msg fail, err:%+v", err)
		}
	}()

	err := consumer.Consume(ctx, conf.Topics, handler)//handler是消费消息的具体处理
	if err != nil {
		logrus.Errorf("Kafka consume error, err:%+v", err)
		return consumer, errors.WithStack(err)
	}
	return consumer, nil
}

/*
open interface outside
type ConsumerHandler struct{}

func (ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ch ConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//consumer msg concrete handle
		fmt.Printf("msg:%+v", msg)
		sess.MarkMessage(msg, "")
	}
	return nil
}
 */