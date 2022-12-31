package xlogrus

import (
	"context"
	"github.com/sirupsen/logrus"
)

// 极简日志接口设计

type Logger interface {
	Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{})
}

type loggerEntity struct {
	*logrus.Logger
}

func NewLoggerEntity(opts ...FuncOpts) *loggerEntity {
	return &loggerEntity{Logger: NewLogger(opts...)}
}

const (
	PanicLevel uint = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

func (l *loggerEntity) Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{}) {
	l.WithContext(ctx).WithFields(fields).Log(logrus.Level(level), msg)
}
