package xelastic

import (
	"fmt"
	"testing"
)

// es6.x之后，一个index下只能创建一个type
func TestES(t *testing.T) {
	Init(ElasticConf{
		Addr: []string{"http://localhost:9200"},
	})
	type T struct {
		Name string `json:"name"`
	}
	tt := &T{Name: "oooo"}
	r, err := Index("city_index11", "doc", tt)
	fmt.Println(string(r))
	fmt.Println(err)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": "oooo",
			},
		},
	}
	rsp, err := Search("city_index11", "doc", query)
	fmt.Println(err)
	fmt.Println(string(rsp))
}
