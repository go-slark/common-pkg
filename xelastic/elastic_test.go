package xelastic

import (
	"fmt"
	"testing"
)

func TestES(t *testing.T) {
	Init(ElasticConf{
		Addr: []string{"http://localhost:9200"},
	})
	type T struct {
		Name string
	}
	ts := make([]*T, 0)
	ts = append(ts, &T{Name: "8888"})
	r, err := CreateBulk("city_index", ts)
	fmt.Println(string(r))
	fmt.Println(err)
}
