package xdelay_queue

import (
	"errors"
	"github.com/go-redis/redis"
)

type BucketZ struct {
	timeScore float64
	jobId     string
}

func (dq *DelayQueue) addJobToBucketZ(timestamp int64, jobId string) error {
	return dq.ZAdd(<- dq.bucketNameChan, redis.Z{
		Score:  float64(timestamp),
		Member: jobId,
	}).Err()
}

func (dq *DelayQueue) getJobFromBucketZ(key string) (*BucketZ, error) {
	result, err := dq.ZRangeWithScores(key, 0, 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}

	bz := &BucketZ{
		timeScore: result[0].Score,
	}
	bz.jobId = result[0].Member.(string)
	return bz, nil
}

func (dq *DelayQueue) removeJobFromBucketZ(bucketName string, jobId string) error {
	return dq.ZRem(bucketName, jobId).Err()
}
