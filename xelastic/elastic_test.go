package xelastic

import (
	"encoding/json"
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
	if err != nil {
		fmt.Println("index err:", err)
		return
	}
	fmt.Println("index result:", string(r))

	// bulk create
	ttt := make([]interface{}, 0)
	ttt = append(ttt, tt, tt)
	rt, err := Bulk("city_index11", "doc", ttt)
	if err != nil {
		fmt.Println("create bulk err:", err)
		return
	}
	fmt.Println("create bulk result:", string(rt))
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": "oooo",
			},
		},
	}
	rsp, err := Search("city_index11", "doc", query)
	if err != nil {
		fmt.Println("search err:", err)
		return
	}

	rp := &response{}
	err = json.Unmarshal(rsp, rp)
	if err != nil {
		fmt.Println("search json unmarshal err:", err)
		return
	}
	fmt.Printf("search result:%+v\n", rp)

	// count
	rc, err := Count([]string{"city_index11"}, []string{"doc"})
	if err != nil {
		fmt.Println("count err:", err)
		return
	}
	fmt.Println("count:", string(rc))
}

type response struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string                 `json:"_index"`
			Type   string                 `json:"_type"`
			ID     string                 `json:"_id"`
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
		}
	} `json:"hits"`
}
