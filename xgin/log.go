package xgin

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func Logger() gin.HandlerFunc {
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
