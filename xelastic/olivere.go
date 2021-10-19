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

		for _, url := range conf.Url {
			_, _, err = client.Ping(url).Do(context.TODO())
			if err != nil {
				panic(errors.WithStack(err))
			}
		}
	})
}

func Stop() {
	client.Stop()
}

func Index(index, typ string, doc interface{}) (*elastic.IndexResponse, error) {
	rsp, err := client.Index().Index(index).Type(typ).BodyJson(doc).Do(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rsp, nil
}

func Get(index, typ string) (*elastic.GetResult, error) {
	result, err := client.Get().Index(index).Type(typ).Do(context.TODO())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}
