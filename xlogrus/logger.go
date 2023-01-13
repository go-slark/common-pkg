package xlogrus

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xlogger"
)

type logrusEntity struct {
	*logrus.Logger
}

func NewLogrusEntity(opts ...FuncOpts) *logrusEntity {
	return &logrusEntity{Logger: NewLogger(opts...)}
}

func (l *logrusEntity) Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{}) {
	var logrusLevel logrus.Level
	switch level {
	case xlogger.DebugLevel:
		logrusLevel = logrus.DebugLevel
	case xlogger.InfoLevel:
		logrusLevel = logrus.InfoLevel
	case xlogger.WarnLevel:
		logrusLevel = logrus.WarnLevel
	case xlogger.ErrorLevel:
		logrusLevel = logrus.ErrorLevel
	case xlogger.FatalLevel:
		logrusLevel = logrus.FatalLevel
	case xlogger.PanicLevel:
		logrusLevel = logrus.PanicLevel
	case xlogger.TraceLevel:
		logrusLevel = logrus.TraceLevel
	default:
		logrusLevel = logrus.DebugLevel
	}
	l.WithContext(ctx).WithFields(fields).Log(logrusLevel, msg)
}
