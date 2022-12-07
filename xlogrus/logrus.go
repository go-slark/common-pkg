package xlogrus

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xutils"
	"io"
	"os"
)

type logger struct {
	srvName   string
	level     logrus.Level
	levels    []logrus.Level
	formatter logrus.Formatter
	writer    io.Writer
	writers   map[logrus.Level]io.Writer
}

type funcOpts func(*logger)

func WithSrvName(name string) funcOpts {
	return func(l *logger) {
		l.srvName = name
	}
}

func WithLevel(level string) funcOpts {
	return func(l *logger) {
		lv, err := logrus.ParseLevel(level)
		if err != nil {
			panic(errors.Errorf("logrus parse level fail, level:%sm err:%+v", level, err))
		}
		l.level = lv
	}
}

func WithLevels(levels []string) funcOpts {
	return func(l *logger) {
		lvs := make([]logrus.Level, 0, len(levels))
		for _, l := range levels {
			lv, err := logrus.ParseLevel(l)
			if err != nil {
				panic(errors.Errorf("logrus parse level fail, levle:%s, err:%+v", l, err))
			}
			lvs = append(lvs, lv)
		}
		l.levels = lvs
	}
}

func WithFormatter(formatter logrus.Formatter) funcOpts {
	return func(l *logger) {
		l.formatter = formatter
	}
}

func WithWriter(writer io.Writer) funcOpts {
	return func(l *logger) {
		l.writer = writer
	}
}

func WithDispatcher(dispatcher map[string]io.Writer) funcOpts {
	return func(l *logger) {
		l.levels = make([]logrus.Level, 0, len(dispatcher))
		l.writers = make(map[logrus.Level]io.Writer, len(dispatcher))
		maxLevel := logrus.Level(len(logrus.AllLevels))
		for level, writer := range dispatcher {
			lv, err := logrus.ParseLevel(level)
			if err != nil {
				continue
			}

			if maxLevel <= lv {
				continue
			}
			l.writers[lv] = writer
			l.levels = append(l.levels, lv)
		}
	}
}

func NewLogger(opts ...funcOpts) *logrus.Logger {
	l := &logger{
		srvName:   "Default-Server",
		level:     logrus.DebugLevel,
		levels:    logrus.AllLevels,
		formatter: &logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"},
		writer:    os.Stdout,
	}
	for _, opt := range opts {
		opt(l)
	}
	stdLogger := logrus.StandardLogger()
	stdLogger.SetFormatter(l.formatter)
	stdLogger.SetLevel(l.level)
	stdLogger.SetOutput(l.writer)
	stdLogger.SetReportCaller(true)
	stdLogger.AddHook(l)
	return stdLogger
}

func (l *logger) Levels() []logrus.Level {
	return l.levels
}

func (l *logger) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}
	entry.Data[xutils.TraceID] = ctx.Value(xutils.TraceID)
	entry.Data[xutils.ServerName] = l.srvName

	// 日志统一分发 es mongo kafka
	writer, ok := l.writers[entry.Level]
	if !ok {
		return nil
	}
	eb, err := entry.Bytes()
	if err != nil {
		return err
	}
	_, err = writer.Write(eb)
	return err
}
