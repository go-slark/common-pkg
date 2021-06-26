package xelastic

import (
	"context"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"sync"
)

var (
	once   sync.Once
	client *elastic.Client
)

type ElasticConf struct {
	Url      []string `json:"url"` //cluster
	UserName string   `json:"user_name"`
	Password string   `json:"password"`
}

func GetElastic() *elastic.Client {
	return client
}

func InitElastic(conf *ElasticConf) {
	once.Do(func() {
		var err error
		options := []elastic.ClientOptionFunc{
			elastic.SetURL(conf.Url...),
			elastic.SetSniff(false),
			//elastic.SetRetrier(),
		}
		if len(conf.UserName) != 0 && len(conf.Password) != 0 {
			options = append(options, elastic.SetBasicAuth(conf.UserName, conf.Password))
		}
		client, err = elastic.NewClient(options...)
		if err != nil {
			panic(errors.WithStack(err))
		}

		_, _, err = client.Ping(conf.Url[0]).Do(context.TODO())
		if err != nil {
			panic(errors.WithStack(err))
		}
	})
}
