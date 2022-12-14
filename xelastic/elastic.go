package xelastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
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

func Create(index, ID, docType string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Create(index, ID, buf, client.Create.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// create / update index (if index not exist, create index and doc, create: id default nil)

func Index(index, docType string, doc interface{}, ID ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	var id string
	if len(ID) != 0 {
		id = ID[0]
	}

	opt := []func(*esapi.IndexRequest){
		client.Index.WithDocumentID(id),
		client.Index.WithDocumentType(docType),
	}
	rsp, err := client.Index(index, buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// bulk create / update / index doc

func Bulk(index, docType string, docs []interface{}) ([]byte, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	buf := &bytes.Buffer{}
	for _, doc := range docs {
		//meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, "id", "\n")) // 批量操作指定_id(_id存在则更新)
		meta := []byte(fmt.Sprintf(`{ "index" : {} }%s`, "\n"))
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	opt := []func(*esapi.BulkRequest){
		client.Bulk.WithIndex(index),
		client.Bulk.WithDocumentType(docType),
	}
	rsp, err := client.Bulk(buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// update doc

func Update(index, docType, ID string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	rsp, err := client.Update(index, ID, buf, client.Update.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// update by condition

func UpdateByQuery(index []string, docType string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	opt := []func(*esapi.UpdateByQueryRequest){
		client.UpdateByQuery.WithDocumentType(docType),
		client.UpdateByQuery.WithBody(buf),
		client.UpdateByQuery.WithContext(context.Background()),
		client.UpdateByQuery.WithPretty(),
	}
	rsp, err := client.UpdateByQuery(index, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// delete doc

func Delete(index, docType, ID string) ([]byte, error) {
	rsp, err := client.Delete(index, ID, client.Delete.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func DeleteBulk(index, docType string, ids []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, id := range ids {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, id, "\n"))
		buf.Grow(len(meta))
		buf.Write(meta)
	}
	opt := []func(*esapi.BulkRequest){
		client.Bulk.WithIndex(index),
		client.Bulk.WithDocumentType(docType),
	}
	rsp, err := client.Bulk(buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func DeleteByQuery(index []string, query interface{}, docType ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	rsp, err := client.DeleteByQuery(index, buf, client.DeleteByQuery.WithDocumentType(docType...))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// query doc

func Get(index, docType, ID string) ([]byte, error) {
	rsp, err := client.Get(index, ID, client.Get.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// search 返回的hit total每次可能不同(不准确)

func Search(index, docType string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	opt := []func(*esapi.SearchRequest){
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithDocumentType(docType),
		client.Search.WithBody(buf),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		//client.Search.WithFrom(0),
		//client.Search.WithSize(3),
		//client.Search.WithSort([]string{"_source:{name:desc}", "_score:asc", "_id:desc"}...), // 多字段排序
		//client.Search.WithScroll(),
	}
	rsp, err := client.Search(opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// count(返回总数准确)

func Count(index, docType []string) ([]byte, error) {
	opt := []func(*esapi.CountRequest){
		client.Count.WithIndex(index...),
		client.Count.WithDocumentType(docType...),
	}
	rsp, err := client.Count(opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}
