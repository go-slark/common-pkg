package xredis

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xjson"
	"sync"
	"time"
)

var (
	redisClients = make(map[string]*redis.Client)
	redisOnce    sync.Once
)

type RedisClientConfig struct {
	Alias        string `json:"alias"`
	Address      string `json:"address"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

func InitRedisClients(configs []*RedisClientConfig) error {
	redisOnce.Do(func() {
		for _, c := range configs {
			if _, ok := redisClients[c.Alias]; ok {
				panic(errors.New("duplicate redis client: " + c.Alias))
			}
			client, err := createRedisClient(c)
			if err != nil {
				panic(errors.New(fmt.Sprintf("redis client %s error %v", xjson.SafeMarshal(c), err)))
			}
			redisClients[c.Alias] = client
		}
	})

	return nil
}

func createRedisClient(c *RedisClientConfig) (*redis.Client, error) {
	options := &redis.Options{
		Network:     "tcp",
		Addr:        c.Address,
		Password:    c.Password,
		DB:          c.DB,
		IdleTimeout: time.Duration(c.IdleTimeout) * time.Second,
	}

	if c.ReadTimeout != 0 {
		options.ReadTimeout = time.Duration(c.ReadTimeout) * time.Second
	}
	if c.WriteTimeout != 0 {
		options.WriteTimeout = time.Duration(c.WriteTimeout) * time.Second
	}
	redisClient := redis.NewClient(options)
	_, err := redisClient.Ping().Result()
	return redisClient, err
}

func GetRedisClient(alias string) *redis.Client {
	return redisClients[alias]
}

func CloseRedisClients() {
	for _, client := range redisClients {
		if client == nil {
			continue
		}

		_ = client.Close()
	}
}
