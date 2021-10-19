package xdelay_queue

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/smallfish-root/common-pkg/xredis"
	"time"
)

type AppConf struct {
	BucketSize        int    `json:"bucket_size"`
	BucketName        string `json:"bucket_name"`         // zset key
	QueueName         string `json:"queue_name"`          // ready queue key
	QueueBlockTimeout int    `json:"queue_block_timeout"` // blpop time out: s
}

type DelayQueue struct {
	*redis.Client
	timers         []*time.Ticker
	bucketNameChan <-chan string

	AppConf
	xredis.RedisClientConfig
}

func buildBucket(bucketName string, index int) string {
	return fmt.Sprintf("%s:%d", bucketName, index)
}

func buildQueue(queue, name string) string {
	return fmt.Sprintf("%s:%s", queue, name)
}
