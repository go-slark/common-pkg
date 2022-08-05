package json

import (
	"encoding/json"
	"github.com/smallfish-root/common-pkg/xencoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"reflect"
)

// std json / proto json

const Name = "json"

var (
	MarshalOptions = protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

type codec struct{}

func init() {
	xencoding.RegisterCodec(&codec{})
}

func (*codec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	case proto.Message:
		return MarshalOptions.Marshal(m)
	default:
		return json.Marshal(m)
	}
}

func (*codec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return UnmarshalOptions.Unmarshal(data, m)
	default:
		rv := reflect.ValueOf(v)
		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = rv.Elem()
		}
		pm, ok := reflect.Indirect(rv).Interface().(proto.Message)
		if ok {
			return UnmarshalOptions.Unmarshal(data, pm)
		}
		return json.Unmarshal(data, m)
	}
}

func (*codec) Name() string {
	return Name
}
