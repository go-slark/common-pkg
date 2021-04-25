package xmongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	client  *mongo.Client
	mongoDB *mongo.Database
)

type MongoConf struct {
	Url      string `json:"url"`
	DateBase string `json:"date_base"`
	Timeout  int    `json:"timeout"`
}

func InitMongoDB(conf *MongoConf) error {
	if conf == nil {
		return errors.New("conf param invalid")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(conf.Url))
	if err != nil {
		return errors.WithStack(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Timeout))
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	mongoDB = client.Database(conf.DateBase)
	return nil
}

type MongoCollection struct {
	*mongo.Collection
}

func NewMongoCollection(coll string) *MongoCollection {
	return &MongoCollection{mongoDB.Collection(coll)}
}

func CloseMongo() error {
	if client == nil {
		return errors.New("mongo client invalid")
	}

	return client.Disconnect(context.TODO())
}
