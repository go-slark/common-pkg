package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/smallfish-root/common-pkg/xgin/xrender"
	"io"
	"net/http"
)

type decoratorHandlerFunc func(*gin.Context) xrender.Render

type handlerFunc func(*gin.Context) error

func HandlerDecorator(fn decoratorHandlerFunc, fs ...handlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, f := range fs {
			err := f(ctx)
			if err != nil {
				r := Error(err)
				ctx.Render(r.Code(), r)
				ctx.Abort()
				_ = ctx.Error(err)
				return
			}
		}

		r := fn(ctx)
		if r == nil {
			return
		}
		err := r.Err()
		if err != nil {
			_ = ctx.Error(err)
		}
		ctx.Render(r.Code(), r)
	}
}

func JSON(code int, obj interface{}, err error) xrender.Render {
	r := xrender.JSON{}
	r.Code_ = code
	r.Data = obj
	r.Error.Update(err)
	return r
}

func String(code int, formant string, values ...interface{}) xrender.Render {
	r := xrender.String{}
	r.Code_ = code
	r.Format = formant
	r.Data = values
	return r
}

func DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) xrender.Render {
	r := xrender.Reader{}
	r.Code_ = code
	r.Headers = extraHeaders
	r.ContentType = contentType
	r.ContentLength = contentLength
	r.Reader.Reader = reader
	return r
}

func Redirect(code int, location string, req *http.Request) xrender.Render {
	r := xrender.Redirect{}
	r.Code_ = -1
	r.Redirect.Code = code
	r.Location = location
	r.Request = req
	return r
}

func Data(code int, contentType string, data []byte) xrender.Render {
	r := xrender.Data{}
	r.Code_ = code
	r.ContentType = contentType
	r.Data.Data = data
	return r
}
