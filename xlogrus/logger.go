package xlogrus

import (
	"context"
	"github.com/sirupsen/logrus"
)

type logrusEntity struct {
	*logrus.Logger
}

func NewLogrusEntity(opts ...FuncOpts) *logrusEntity {
	return &logrusEntity{Logger: NewLogger(opts...)}
}

func (l *logrusEntity) Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{}) {
	l.WithContext(ctx).WithFields(fields).Log(logrus.Level(level), msg)
}
