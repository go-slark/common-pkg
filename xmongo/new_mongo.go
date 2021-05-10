package xmongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	client  *mongo.Client
	mongoDB *mongo.Database
	once    sync.Once
)

type MongoConf struct {
	Url         string `json:"url"`
	DateBase    string `json:"date_base"`
	Timeout     int    `json:"timeout"`
	MaxPoolSize uint64 `json:"max_pool_size"`
	MinPoolSize uint64 `json:"min_pool_size"`
}

func InitMongoDB(conf *MongoConf) error {
	if conf == nil {
		return errors.New("conf param invalid")
	}

	var err error
	opts := options.Client()
	if conf.MaxPoolSize != 0 {
		opts.SetMaxPoolSize(conf.MaxPoolSize)
	}
	if conf.MinPoolSize != 0 {
		opts.SetMinPoolSize(conf.MinPoolSize)
	}
	opts.ApplyURI(conf.Url)
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Timeout))
		defer cancel()
		client, err = mongo.Connect(ctx, opts)
		if err != nil {
			panic(errors.WithStack(err))
		}

		mongoDB = client.Database(conf.DateBase)
	})

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func NewMongoCollection(coll string) *mongo.Collection {
	return mongoDB.Collection(coll)
}

func CloseMongo() error {
	if client == nil {
		return errors.New("mongo client invalid")
	}

	return client.Disconnect(context.TODO())
}
