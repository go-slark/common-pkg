package json

import (
	"fmt"
	"github.com/smallfish-root/common-pkg/xencoding"
	"math"
	"testing"
)

// compare std json with proto json

func TestJson(t *testing.T) {
	mock := &Person{
		Name: "ChengDu",
		Age:  math.MaxInt64,
		Numbers: []*Person_PhoneNumber{{
			Number: "100000000001",
			Type:   PhoneType_Home,
		},
		},
	}

	// marshal
	result, err := xencoding.GetCodec(Name).Marshal(mock)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(result))

	// unmarshal
	rsp := &Person{}
	xencoding.GetCodec(Name).Unmarshal(result, rsp)
	fmt.Printf("%+v\n", rsp)
}
