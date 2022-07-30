package xgin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/smallfish-root/common-pkg/xcrypto"
	"io/ioutil"
	"net/http"
)

func DecryptRequestParam(key []byte) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		switch ctx.Request.Method {
		case http.MethodGet:
			result, err := xcrypto.Base64AesCBCDecrypt(ctx.Request.URL.RawQuery, key)
			if err != nil {
				ctx.Abort()
				return
			}
			ctx.Request.URL.RawQuery = string(result)
		case http.MethodPost:
			bodyBytes, _ := ioutil.ReadAll(ctx.Request.Body)
			result, err := xcrypto.Base64AesCBCDecrypt(string(bodyBytes), key)
			if err != nil {
				ctx.Abort()
				return
			}
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(result))
		}
	}
}
