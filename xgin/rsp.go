package xgin

import (
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xgin/xrender"
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
	rsp := &Response{
		Code: 0,
		Msg:  "成功",
		Data: data,
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
