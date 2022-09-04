package xgin

import (
	"github.com/smallfish-root/common-pkg/xerror"
	rsp "github.com/smallfish-root/common-pkg/xgin/protorsp"
	"github.com/smallfish-root/common-pkg/xgin/xrender"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) xrender.Render {
	switch m := data.(type) {
	case proto.Message:
		any, _ := anypb.New(m)
		return ProtoJSON(http.StatusOK, &rsp.Response{
			Code: 0,
			Msg:  "成功",
			Data: any,
		}, nil)
	default:
		if data == nil {
			data = struct{}{}
		}
		return JSON(http.StatusOK, &Response{
			Code: 0,
			Msg:  "成功",
			Data: data,
		}, nil)
	}
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
