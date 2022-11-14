package xencoding

import (
	"bytes"
	"encoding/gob"
	"strings"
)

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

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(src)
	if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
