package xdelay_queue

import (
	"encoding/json"
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

//func (dq *DelayQueue) AddJob(job *Job) error {
//	if job.Id == "" || job.Topic == "" || job.Delay < 0 || job.TTR <= 0 {
//		return errors.New("invalid job")
//	}
//
//	err := dq.addJob(job.Id, job)
//	if err != nil {
//		return errors.WithStack(err)
//	}
//
//	err = dq.addJobToBucketZ(job.Delay, job.Id)
//	if err != nil {
//		return errors.WithStack(err)
//	}
//
//	return nil
//}

// AddJob transaction
func (dq *DelayQueue) AddJob(jobCore *JobCore) error {
	if jobCore == nil {
		return errors.New("job param invalid")
	}

	if jobCore.Id == "" || jobCore.Topic == "" || jobCore.Delay < 0 || jobCore.TTR <= 0 {
		return errors.New("invalid job")
	}

	jobBucket := <-dq.bucketNameChan
	job := &Job{
		JobCore:   jobCore,
		DoneTimes: 0,
		JobBucket: jobBucket,
	}
	value, err := json.Marshal(job)
	if err != nil {
		return errors.WithStack(err)
	}

	err = redis.NewScript(`
        redis.call("SET", KEYS[1], ARGV[1])
        redis.call("ZADD", KEYS[2], ARGV[2], ARGV[3])
    `).Run(dq.Client, []string{job.Id, jobBucket.BucketName}, value, float64(job.Delay), job.Id).Err()
	if err == nil || errors.Is(err, redis.Nil) {
		return nil
	}
	return errors.WithStack(err)
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

	jobBucket := <-dq.bucketNameChan
	job.DoneTimes++
	if job.Times <= 0 || job.DoneTimes < job.Times {
		job.JobBucket = jobBucket
		value, _ := json.Marshal(job)
		_ = redis.NewScript(`
        redis.call("SET", KEYS[1], ARGV[1])
        redis.call("ZADD", KEYS[2], ARGV[2], ARGV[3])
    `).Run(dq.Client, []string{job.Id, jobBucket.BucketName}, value, float64(time.Now().Unix()+job.TTR), job.Id).Err()
	}
	return job, nil
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

func (dq *DelayQueue) UpdateJob(jobCore *JobCore) error {
	if jobCore == nil {
		return errors.New("invalid param")
	}
	job, err := dq.getJob(jobCore.Id)
	if err != nil {
		return errors.WithStack(err)
	}
	if job.DoneTimes > 0 {
		return errors.New("job is doing")
	}
	value, err := json.Marshal(&Job{
		JobCore:   jobCore,
		JobBucket: job.JobBucket,
		DoneTimes: 0,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return redis.NewScript(`
        redis.call("SET", KEYS[1], ARGV[1])
        redis.call("ZADD", KEYS[2], ARGV[2], ARGV[3])
    `).Run(dq.Client, []string{job.Id, job.BucketName}, value, float64(jobCore.Delay), job.Id).Err()
}

func (dq *DelayQueue) BatchAddJob(jobCores []*JobCore) error {
	jobCoreNum := len(jobCores)
	if jobCoreNum == 0 {
		return errors.New("invalid param")
	}
	jobKey := make([]string, 0, jobCoreNum)
	jobValue := make([][]byte, jobCoreNum)
	jobDelay := make([]float64, 0, jobCoreNum)
	jobBucketName := make([]string, 0, jobCoreNum)
	for index, jobCore := range jobCores {
		if jobCore == nil || jobCore.Id == "" || jobCore.Topic == "" || jobCore.Delay < 0 || jobCore.TTR <= 0 {
			return errors.New("invalid job")
		}
		jobBucket := <-dq.bucketNameChan
		jobKey = append(jobKey, jobCore.Id)
		value, err := json.Marshal(&Job{
			JobCore:   jobCore,
			JobBucket: jobBucket,
			DoneTimes: 0,
		})
		if err != nil {
			return errors.WithStack(err)
		}
		jobValue[index] = value
		jobDelay = append(jobDelay, float64(jobCore.Delay))
		jobBucketName = append(jobBucketName, jobBucket.BucketName)
	}

	argv := make([]interface{}, 0, 4*jobCoreNum)
	for _, value := range jobValue {
		argv = append(argv, value)
	}
	for _, name := range jobBucketName {
		argv = append(argv, name)
	}

	for _, delay := range jobDelay {
		argv = append(argv, delay)
	}

	for _, key := range jobKey {
		argv = append(argv, key)
	}

	src := `
        local l = table.getn(KEYS)
        for k, v in pairs(KEYS) do
            redis.call("SET", v, ARGV[k])
            redis.call("ZADD", ARGV[l+k], ARGV[2*l+k], ARGV[3*l+k])
        end
    `
	return redis.NewScript(src).Run(dq.Client, jobKey, argv...).Err()
}

func (dq *DelayQueue) getBucketName() {
	c := make(chan *JobBucket)
	go func() {
		for {
			index := rand.Int() % dq.BucketSize
			c <- &JobBucket{
				BucketName:  buildBucket(dq.BucketName, index),
				BucketIndex: index,
			}
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
			jobBucket := <-dq.bucketNameChan
			job.JobBucket = jobBucket
			value, _ := json.Marshal(job)
			_ = redis.NewScript(`
                redis.call("ZREM", KEYS[1], ARGV[1])
                redis.call("ZADD", KEYS[2], ARGV[2], ARGV[3])
                redis.call("SET", KEYS[3], ARGV[4])
            `).Run(dq.Client, []string{bucketName, jobBucket.BucketName, job.Id}, bucketZ.jobId, float64(job.Delay), bucketZ.jobId, value).Err()
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
