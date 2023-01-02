package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xerror"
	"github.com/smallfish-root/common-pkg/xlogger"
	httpLogger "github.com/smallfish-root/gin-http-logger"
)

func ErrLogger(logger xlogger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		context := ctx.Request.Context()
		for _, err := range ctx.Errors {
			ce, ok := err.Err.(*xerror.CustomError)
			if !ok {
				logger.Log(context, xlogger.ErrorLevel, map[string]interface{}{"meta": err.Meta, "error": err.Err}, "系统异常")
			} else {
				fields := map[string]interface{}{
					"surplus": ce.Surplus,
					"meta":    ce.Metadata,
					"code":    ce.Code,
					"error":   ce.GetError(),
				}
				logger.Log(context, xlogger.ErrorLevel, fields, ce.Message)
			}
		}
	}
}

func Logger(excludePaths ...string) gin.HandlerFunc {
	l := httpLogger.AccessLoggerConfig{
		LogrusLogger:   logrus.StandardLogger(),
		BodyLogPolicy:  httpLogger.LogAllBodies,
		MaxBodyLogSize: 1024 * 16, //16k
		DropSize:       1024 * 10, //10k
	}

	l.ExcludePaths = map[string]struct{}{}
	for _, excludePath := range excludePaths {
		l.ExcludePaths[excludePath] = struct{}{}
	}
	return httpLogger.New(l)
}
