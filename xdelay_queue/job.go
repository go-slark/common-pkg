package xdelay_queue

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type JobBucket struct {
	BucketName  string `json:"bucket_name"`
	BucketIndex int    `json:"bucket_index"`
}

type JobCore struct {
	Topic string `json:"topic"`
	Id    string `json:"id"`
	Delay int64  `json:"delay"`
	TTR   int64  `json:"ttr"`
	Times int64  `json:"times"`
	Body  []byte `json:"body"`
	Url   string `json:"url"`
}

type Job struct {
	*JobCore
	*JobBucket
	DoneTimes int64 `json:"done_times"`
}

func (dq *DelayQueue) getJob(key string) (*Job, error) {
	result, err := dq.Get(key).Bytes()
	if err != nil {
		//if errors.Is(err, redis.Nil) {
		//	return nil, nil
		//}
		return nil, errors.WithStack(err)
	}
	if len(result) == 0 {
		return nil, nil
	}

	job := &Job{}
	err = json.Unmarshal(result, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (dq *DelayQueue) addJob(key string, job *Job) error {
	value, err := json.Marshal(job)
	if err != nil {
		return errors.WithStack(err)
	}

	return dq.Set(key, value, 0).Err()
}

func (dq *DelayQueue) deleteJob(key string) error {
	return dq.Del(key).Err()
}
