package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xerror"
	httpLogger "github.com/smallfish-root/gin-http-logger"
)

func ErrLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		context := ctx.Request.Context()
		for _, err := range ctx.Errors {
			ce, ok := err.Err.(*xerror.CustomError)
			if !ok {
				logrus.WithContext(context).WithFields(logrus.Fields{"meta": err.Meta}).WithError(err.Err).Error("系统异常")
			} else {
				fields := logrus.Fields{
					"surplus": ce.Surplus,
					"meta":    ce.Metadata,
					"code":    ce.Code,
				}
				logrus.WithContext(context).WithFields(fields).WithError(ce.GetError()).Error(ce.Message)
			}
		}
	}
}

func Logger(excludePaths ...string) gin.HandlerFunc {
	l := httpLogger.AccessLoggerConfig{
		//LogrusLogger:   logger,
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
