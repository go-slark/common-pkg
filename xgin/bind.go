package xgin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xvalidator"
	"reflect"
)

const requestKey = "api-request"

const (
	jsonFormatReq uint8 = iota
	formFormatReq
	uriFormatReq
	protoFormatReq
)

func bindRequest(reqObj interface{}, format uint8) gin.HandlerFunc {
	reqType := reflect.TypeOf(reqObj)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	return func(ctx *gin.Context) {
		var (
			f func(obj interface{}) error
			b binding.Binding
		)
		switch format {
		case jsonFormatReq:
			b = binding.JSON
		case formFormatReq:
			b = binding.Form
		case protoFormatReq:
			b = binding.ProtoBuf
		case uriFormatReq:
		default:
			r := Error(xerror.NewError(xerror.FormatInvalidCode, xerror.FormatInvalid, xerror.FormatInvalid).WithError(errors.New(fmt.Sprintf("invalid format:%d", format))))
			_ = ctx.Error(r.Err())
			ctx.Render(r.Code(), r)
			ctx.Abort()
			return
		}

		if b == nil {
			f = ctx.ShouldBindUri
		} else {
			f = func(obj interface{}) error {
				return ctx.ShouldBindWith(obj, b)
			}
		}

		req := reflect.New(reqType).Interface()
		if err := f(req); err != nil {
			te := xvalidator.ParseError(err)
			if len(te) != 0 {
				err = errors.New(fmt.Sprintf("%v", te))
			}
			r := Error(xerror.NewError(xerror.ParamValidCode, xerror.ParamValid, err.Error()).WithError(err))
			_ = ctx.Error(r.Err())
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

func BindProto(req interface{}) gin.HandlerFunc {
	return bindRequest(req, protoFormatReq)
}

func DefaultRequest(ctx *gin.Context) interface{} {
	return ctx.MustGet(requestKey)
}
