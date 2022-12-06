package xgin

import (
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xgin/xrender"
	"google.golang.org/protobuf/proto"
	"net/http"
)

const SimpleRsp = "X-Response"

type Response struct {
	Code    int         `json:"code"`
	TraceID interface{} `json:"trace_id"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func Success(data interface{}) xrender.Render {
	if data == nil {
		data = struct{}{}
	}
	var (
		msg  = "成功"
		code int
		rsp  interface{}
	)
	switch data.(type) {
	case proto.Message:
		rsp = &xrender.ProtoResponse{
			Code:    code,
			Msg:     msg,
			Message: data.(proto.Message),
		}
	default:
		rsp = &Response{
			Code: code,
			Msg:  msg,
			Data: data,
		}
	}
	return JSON(http.StatusOK, rsp, nil)
}

func Error(err error) xrender.Render {
	e := xerror.GetErr(err)
	return JSON(http.StatusOK, &Response{
		Code: int(e.Status.Code),
		Msg:  e.Status.Message,
		Data: struct{}{},
	}, err)
}

func Reply(data interface{}, err error) xrender.Render {
	if err != nil {
		return Error(err)
	}
	return Success(data)
}
