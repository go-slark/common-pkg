package xdelay_queue

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xredis"
	"math/rand"
	"time"
)

var delayQueue = make(map[string]*DelayQueue)

func InitDelayQueue(dqs []*DelayQueue) {
	rand.Seed(time.Now().UnixNano())
	for _, dq := range dqs {
		xredis.AppendRedisClients([]*xredis.RedisClientConfig{&dq.RedisClientConfig})
		delayQueue[dq.Alias] = dq
		dq.Client = xredis.GetRedisClient(dq.Alias)
		dq.initTimers()
		dq.getBucketName()
	}
}

func GetDelayQueue(alias string) *DelayQueue {
	return delayQueue[alias]
}

func (dq *DelayQueue) AddJob(job *Job) error {
	if job.Id == "" || job.Topic == "" || job.Delay < 0 || job.TTR <= 0 {
		return errors.New("invalid job")
	}

	err := dq.addJob(job.Id, job)
	if err != nil {
		return errors.WithStack(err)
	}

	err = dq.addJobToBucketZ(job.Delay, job.Id)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (dq *DelayQueue) GetJob(topics []string) (*Job, error) {
	jobId, err := dq.getJobFromReadyQueue(topics)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	job, err := dq.getJob(jobId)
	if err != nil {
		return job, errors.WithStack(err)
	}

	if job == nil {
		return nil, nil
	}

	timestamp := time.Now().Unix() + job.TTR
	err = dq.addJobToBucketZ(timestamp, job.Id)
	return job, err
}

func (dq *DelayQueue) DeleteJob(jobId string) error {
	return dq.deleteJob(jobId)
}

func (dq *DelayQueue) QueryJob(jobId string) (*Job, error) {
	job, err := dq.getJob(jobId)
	if err != nil {
		return job, errors.WithStack(err)
	}

	if job == nil {
		return nil, nil
	}
	return job, err
}

func (dq *DelayQueue) getBucketName() {
	c := make(chan string)
	go func() {
		for {
			c <- buildBucket(dq.BucketName, rand.Int()%dq.BucketSize)
		}
	}()
	dq.bucketNameChan = c
}

func (dq *DelayQueue) initTimers() {
	dq.timers = make([]*time.Ticker, dq.BucketSize)
	for i := 0; i < dq.BucketSize; i++ {
		dq.timers[i] = time.NewTicker(1 * time.Second)
		go dq.waitTicker(dq.timers[i], buildBucket(dq.BucketName, i))
	}
}

func (dq *DelayQueue) waitTicker(timer *time.Ticker, bucketName string) {
	defer timer.Stop()
	for {
		select {
		case t := <-timer.C:
			dq.handleTicker(t, bucketName)
		}
	}
}

func (dq *DelayQueue) handleTicker(t time.Time, bucketName string) {
	for {
		bucketZ, err := dq.getJobFromBucketZ(bucketName)
		if err != nil {
			return
		}

		if bucketZ == nil {
			return
		}

		if int64(bucketZ.timeScore) > t.Unix() {
			return
		}

		job, err := dq.getJob(bucketZ.jobId)
		if err != nil {
			continue
		}

		if job == nil {
			_ = dq.removeJobFromBucketZ(bucketName, bucketZ.jobId)
			continue
		}

		if job.Delay > t.Unix() {
			//_ = dq.removeJobFromBucketZ(bucketName, bucketZ.jobId)
			//_ = dq.addJobToBucketZ(job.Delay, bucketZ.jobId)
			_ = redis.NewScript(`
                redis.call("ZREM", KEYS[1], ARGV[1])
                redis.call("ZADD", KEYS[2], ARGV[2])
            `).Run(dq.Client, []string{bucketName, <-dq.bucketNameChan}, bucketZ.jobId, redis.Z{
				Score:  float64(job.Delay),
				Member: bucketZ.jobId,
			}).Err()
			continue
		}

		//err = dq.addJobToReadyQueue(job.Topic, bucketZ.jobId)
		//if err != nil {
		//	continue
		//}
		//
		//err = dq.removeJobFromBucketZ(bucketName, bucketZ.jobId)
		//if err != nil {
		//}

		_ = redis.NewScript(`
               redis.call("RPUSH", KEYS[1], ARGV[1])
               redis.call("ZREM", KEYS[2], ARGV[2])
           `).Run(dq.Client, []string{buildQueue(dq.QueueName, job.Topic), bucketName}, bucketZ.jobId, bucketZ.jobId).Err()
	}
}
