package xdelay_queue

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"time"
)

func (dq *DelayQueue) addJobToReadyQueue(queueName string, jobId string) error {
	return dq.RPush(buildQueue(dq.QueueName, queueName), jobId).Err()
}

func (dq *DelayQueue) getJobFromReadyQueue(queues []string) (string, error) {
	queueKey := make([]string, 0, len(queues))
	for _, queue := range queues {
		queueKey = append(queueKey, buildQueue(dq.QueueName, queue))
	}

	result, err := dq.BLPop(time.Duration(dq.QueueBlockTimeout)*time.Second, queueKey...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", errors.WithStack(err)
	}
	if len(result) != 2 {
		return "", nil
	}

	return result[1], nil
}
