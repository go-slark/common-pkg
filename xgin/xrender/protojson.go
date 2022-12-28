package xrender

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type ProtoJson struct {
	Code    int
	TraceID interface{}
	Msg     string
	proto.Message
}

var MarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: true,
}

func (p ProtoJson) Render(w http.ResponseWriter) (err error) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
	jsonBytes, err := MarshalOptions.Marshal(p)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		panic(err)
	}
	return
}

func (p ProtoJson) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}
