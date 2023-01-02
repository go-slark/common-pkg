package xlogger

import (
	"context"
)

// 极简日志接口设计

type Logger interface {
	Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{})
}

//level

const (
	PanicLevel uint = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)
