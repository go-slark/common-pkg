package xdelay_queue

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/smallfish-root/common-pkg/xredis"
	"time"
)

type AppConf struct {
	BucketSize        int    `mapstructure:"bucket_size" json:"bucket_size"`
	BucketName        string `mapstructure:"bucket_name" json:"bucket_name"`                 // zset key
	QueueName         string `mapstructure:"queue_name" json:"queue_name"`                   // ready queue key
	QueueBlockTimeout int    `mapstructure:"queue_block_timeout" json:"queue_block_timeout"` // blpop time out: s
}

type DelayQueue struct {
	*redis.Client
	timers         []*time.Ticker
	bucketNameChan <-chan string

	AppConf                  `mapstructure:"app_conf" json:"app_conf"`
	xredis.RedisClientConfig `mapstructure:"redis_conf" json:"redis_conf"`
}

func buildBucket(bucketName string, index int) string {
	return fmt.Sprintf("%s:%d", bucketName, index)
}

func buildQueue(queue, name string) string {
	return fmt.Sprintf("%s:%s", queue, name)
}
