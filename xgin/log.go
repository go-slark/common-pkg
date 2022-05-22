package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	httpLogger "github.com/smallfish-root/gin-http-logger"
	"io"
	"os"
)

func ErrLogger() gin.HandlerFunc {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	formatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	logger.SetFormatter(formatter)
	logger.SetOutput(io.MultiWriter([]io.Writer{os.Stdout}...))
	return func(ctx *gin.Context) {
		ctx.Next()
		for _, err := range ctx.Errors {
			if err != nil {
				logger.Errorf("%+v", err.Err)
			}
		}
	}
}

func Logger(excludePaths ...string) gin.HandlerFunc {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	formatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	logger.SetFormatter(formatter)
	logger.SetOutput(io.MultiWriter([]io.Writer{os.Stdout}...))
	l := httpLogger.AccessLoggerConfig{
		LogrusLogger:   logger,
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
