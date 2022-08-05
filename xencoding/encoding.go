package xencoding

import "strings"

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

var codecRegister = make(map[string]Codec)

func RegisterCodec(codec Codec) {
	if codec == nil {
		panic("codec nil")
	}
	if len(codec.Name()) == 0 {
		panic("codec name empty")
	}
	codecRegister[strings.ToLower(codec.Name())] = codec
}

func GetCodec(name string) Codec {
	return codecRegister[name]
}
