package xelastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"io/ioutil"
	"strings"
)

var client *elasticsearch.Client

type ElasticConf struct {
	Addr     []string `json:"addr"` //cluster
	UserName string   `json:"user_name"`
	Password string   `json:"password"`
}

func Init(c ElasticConf) {
	cfg := elasticsearch.Config{
		Addresses: c.Addr,
	}

	if len(c.UserName) != 0 && len(c.Password) != 0 {
		cfg.Username = c.UserName
		cfg.Password = c.Password
	}

	var err error
	client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// cluster info
	rsp, err := client.Info()
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	body := map[string]interface{}{}
	err = json.NewDecoder(rsp.Body).Decode(&body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("elastic search info:%+v\n", body)
}

func Stop() {}

// create index

func CreateIndex(index, str string) ([]byte, error) {
	rsp, err := client.Indices.Create(index, client.Indices.Create.WithBody(strings.NewReader(str)))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// query index

func GetIndex(index []string) ([]byte, error) {
	rsp, err := client.Indices.Get(index)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// create doc

func Create(index, ID string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Create(index, ID, buf)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// create / update index (if index not exist, create index and doc, create: id default nil)

func Index(index string, doc interface{}, ID ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	var id string
	if len(ID) != 0 {
		id = ID[0]
	}
	rsp, err := client.Index(index, buf, client.Index.WithDocumentID(id))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// bulk create doc

func CreateBulk(index string, docs ...interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, doc := range docs {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, "id", "\n")) // TODO
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	rsp, err := client.Bulk(buf, client.Bulk.WithIndex(index))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// update doc

func Update(index, ID string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	rsp, err := client.Update(index, ID, buf, client.Update.WithDocumentType("doc"))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// update by condition

func UpdateByQuery(index []string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	rsp, err := client.UpdateByQuery(index, client.UpdateByQuery.WithDocumentType("doc"),
		client.UpdateByQuery.WithBody(buf),
		client.UpdateByQuery.WithContext(context.Background()),
		client.UpdateByQuery.WithPretty())
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// delete doc

func Delete(index, ID string) ([]byte, error) {
	rsp, err := client.Delete(index, ID)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func DeleteBulk(index string, ids []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, id := range ids {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, id, "\n"))
		buf.Grow(len(meta))
		buf.Write(meta)
	}
	rsp, err := client.Bulk(buf, client.Bulk.WithIndex(index))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func DeleteByQuery(index []string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	rsp, err := client.DeleteByQuery(index, buf)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// query doc

func Get(index, ID string) ([]byte, error) {
	rsp, err := client.Get(index, ID)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func Search(index string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	rsp, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(buf),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}
