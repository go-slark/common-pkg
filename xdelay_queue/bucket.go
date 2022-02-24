package xdelay_queue

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type BucketZ struct {
	timeScore float64
	jobId     string
}

func (dq *DelayQueue) addJobToBucketZ(timestamp int64, jobId string) error {
	jobBucket := <-dq.bucketNameChan
	return dq.ZAdd(jobBucket.BucketName, redis.Z{
		Score:  float64(timestamp),
		Member: jobId,
	}).Err()
}

func (dq *DelayQueue) getJobFromBucketZ(key string) (*BucketZ, error) {
	//result, err := dq.ZRangeWithScores(key, 0, 0).Result()
	result, err := dq.ZRangeByScore(key, redis.ZRangeBy{
		Max:    fmt.Sprintf("%d", time.Now().Unix()),
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}

	bz := &BucketZ{
		jobId: result[0],
	}
	return bz, nil
}

func (dq *DelayQueue) removeJobFromBucketZ(bucketName string, jobId string) error {
	return dq.ZRem(bucketName, jobId).Err()
}
