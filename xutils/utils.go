package xutils

import (
	"encoding/json"
	"github.com/google/uuid"
	"math/rand"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

func RandomString(l int) string {
	str := "0123456789AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
	bytes := []byte(str)
	var result []byte = make([]byte, 0, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return BytesToString(result)
}

// BytesToString 0 拷贝转换 slice byte 为 string
func BytesToString(b []byte) (s string) {
	_bptr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	_sptr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	_sptr.Data = _bptr.Data
	_sptr.Len = _bptr.Len
	return s
}

func Int642String(intNum int64) (int64Str string) {
	//10, 代表10进制
	int64Str = strconv.FormatInt(intNum, 10)
	return
}

type BodyMap map[string]interface{}

func (bm BodyMap) JsonBody() (jb string) {
	bs, err := json.Marshal(bm)
	if err != nil {
		return ""
	}
	jb = string(bs)
	return jb
}

const (
	ServerName = "server-name"
	TraceID    = "x-request-id"
)

func BuildRequestID() string {
	return uuid.New().String()
}
