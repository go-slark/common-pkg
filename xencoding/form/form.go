package form

import (
	"github.com/go-playground/form/v4"
	"github.com/smallfish-root/common-pkg/xencoding"
	"google.golang.org/protobuf/proto"
	"net/url"
	"reflect"
)

const (
	Name    = "x-www-form-urlencoded"
	nullStr = "null"
)

func init() {
	decoder := form.NewDecoder()
	decoder.SetTagName("json")
	xencoding.RegisterCodec(&codec{
		decoder: decoder,
	})
}

type codec struct {
	encoder *form.Encoder
	decoder *form.Decoder
}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	return []byte{}, nil
}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, vs)
	} else if m, ok := reflect.Indirect(reflect.ValueOf(v)).Interface().(proto.Message); ok {
		return DecodeValues(m, vs)
	}

	return c.decoder.Decode(v, vs)
}

func (*codec) Name() string {
	return Name
}
