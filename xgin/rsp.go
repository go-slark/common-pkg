package xgin

import (
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xgin/xrender"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) xrender.Render {
	if data == nil {
		data = struct{}{}
	}
	return JSON(http.StatusOK, &Response{
		Code: 0,
		Msg:  "成功",
		Data: data,
	})
}

func Error(err error) xrender.Render {
	e := xerror.GetErr(err)
	return JSON(http.StatusOK, &Response{
		Code: e.Code,
		Msg:  e.Msg,
		Data: struct{}{},
	})
}

func Reply(data interface{}, err error) xrender.Render {
	if err != nil {
		return Error(err)
	}
	return Success(data)
}
