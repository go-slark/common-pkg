package xmongo

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xjson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	client  = make(map[string]*mongo.Client)
	mongoDB = make(map[string]*mongo.Database)
	once    sync.Once
)

type MongoConf struct {
	Alias       string `json:"alias"`
	Url         string `json:"url"`
	DateBase    string `json:"data_base"`
	Timeout     int    `json:"timeout"`
	MaxPoolSize uint64 `json:"max_pool_size"`
	MinPoolSize uint64 `json:"min_pool_size"`
}

func createMongoClient(c *MongoConf) (*mongo.Client, error) {
	opts := options.Client()
	if c.MaxPoolSize != 0 {
		opts.SetMaxPoolSize(c.MaxPoolSize)
	}
	if c.MinPoolSize != 0 {
		opts.SetMinPoolSize(c.MinPoolSize)
	}
	opts.ApplyURI(c.Url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout))
	defer cancel()
	cli, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = cli.Ping(context.TODO(), nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return cli, nil
}

func InitMongoDB(conf []*MongoConf) {
	once.Do(func() {
		for _, c := range conf {
			if _, ok := client[c.Alias]; ok {
				panic(errors.New("duplicate mongo client: " + c.Alias))
			}

			if _, ok := mongoDB[c.Alias]; ok {
				panic(errors.New("duplicate mongo db: " + c.Alias))
			}

			cli, err := createMongoClient(c)
			if err != nil {
				panic(errors.New(fmt.Sprintf("redis pool %s error %v", xjson.SafeMarshal(c), err)))
			}
			client[c.Alias] = cli
			mongoDB[c.Alias] = cli.Database(c.DateBase)
		}
	})
}

func NewMongoCollection(alias, coll string) *mongo.Collection {
	return mongoDB[alias].Collection(coll)
}

func GetMongoDB(alias string) *mongo.Database {
	return mongoDB[alias]
}

func CloseMongo() {
	for _, cli := range client {
		_ = cli.Disconnect(context.TODO())
	}
}
