package xgin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xvalidator"
	"reflect"
)

const requestKey = "api-request"

const (
	jsonFormatReq uint8 = iota
	formFormatReq
	uriFormatReq
)

func bindRequest(reqObj interface{}, format uint8) gin.HandlerFunc {
	reqType := reflect.TypeOf(reqObj)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	return func(ctx *gin.Context) {
		var f func(obj interface{}) error
		switch format {
		case jsonFormatReq, formFormatReq:
			f = ctx.ShouldBind
		case uriFormatReq:
			f = ctx.ShouldBindUri
		default:
			r := Error(errors.New(fmt.Sprintf("req format invalid, req format is :%v", format)))
			ctx.Render(r.Code(), r)
			ctx.Abort()
			return
		}

		req := reflect.New(reqType).Interface()
		if err := f(req); err != nil {
			te := xvalidator.Error(err)
			if len(te) != 0 {
				err = errors.New(fmt.Sprintf("%v", te))
			}
			r := Error(err)
			ctx.Render(r.Code(), r)
			ctx.Abort()
			return
		}
		ctx.Set(requestKey, req)
		//ctx.Next()
	}
}

func BindJson(req interface{}) gin.HandlerFunc {
	return bindRequest(req, jsonFormatReq)
}

func BindUri(req interface{}) gin.HandlerFunc {
	return bindRequest(req, uriFormatReq)
}

func BindForm(req interface{}) gin.HandlerFunc {
	return bindRequest(req, formFormatReq)
}

func DefaultRequest(ctx *gin.Context) interface{} {
	return ctx.MustGet(requestKey)
}
