package proto

import (
	"github.com/smallfish-root/common-pkg/xencoding"
	"google.golang.org/protobuf/proto"
)

const Name = "proto"

type codec struct{}

func init() {
	xencoding.RegisterCodec(&codec{})
}

func (*codec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (*codec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (*codec) Name() string {
	return Name
}
