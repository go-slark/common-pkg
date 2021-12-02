package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/smallfish-root/common-pkg/xdelay_queue"
	"github.com/smallfish-root/common-pkg/xredis"
	"time"
)

func initDelayQueue() {
	conf := []*xdelay_queue.DelayQueue{{
		AppConf:           xdelay_queue.AppConf{
			BucketSize:        3,
			BucketName:        "test_delay_queue_7#",
			QueueName:         "test_queue_7#",
			QueueBlockTimeout: 5,
		},
		RedisClientConfig: xredis.RedisClientConfig{
			Alias:              "campus",
			Address:            "addr:port",
			Password:           "password",
			DB:                 13,
			DialTimeout:        5,
			ReadTimeout:        5,
			WriteTimeout:       5,
			IdleTimeout:        120,
		},
	}}
	xdelay_queue.InitDelayQueue(conf)
}

func main() {
	initDelayQueue()
	dq := xdelay_queue.GetDelayQueue("campus")
	//go func() {
	//	for {
			//time.Sleep(3*time.Second)
			err := dq.AddJob(&xdelay_queue.Job{
				Topic: "^^^",
				Id:    uuid.NewString(),
				Delay: time.Now().Unix() + 20,
				TTR:   3, // time to retry
				Body:  []byte("test delay queue"),
			})
			if err != nil {
				fmt.Println("^^^^ err:", err)
			}

			err = dq.AddJob(&xdelay_queue.Job{
				Topic: "***",
				Id:    uuid.NewString(),
				Delay: time.Now().Unix() + 10,
				TTR:   3,
				Body:  []byte("test delay queue-----"),
			})
			if err != nil {
				fmt.Println("**** err:", err)
			}
	//	}
	//}()

	for {
		job, err := dq.GetJob([]string{"^^^", "***"})
		if err != nil {
			fmt.Println("get job fail:", err)
			break
		}
		if job == nil {
			fmt.Println("------------------")
			continue
		}

		_ = dq.DeleteJob(job.Id) // prevent repeated to exec, avoid invalid execution multiple times!!!  be commented out for multi times to exec
		fmt.Println("body:", job.Body)
	}
}

