package xmysql

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type FuncOpts func(l *customizedLogger)

func WithLogLevel(level logger.LogLevel) FuncOpts {
	return func(l *customizedLogger) {
		l.LogLevel = level
	}
}

func WithColor(color bool) FuncOpts {
	return func(l *customizedLogger) {
		l.Colorful = color
	}
}

func WithSlowThreshold(tm time.Duration) FuncOpts {
	return func(l *customizedLogger) {
		l.SlowThreshold = tm
	}
}

func WithRecordNotFound(i bool) FuncOpts {
	return func(l *customizedLogger) {
		l.IgnoreRecordNotFoundError = i
	}
}

type customizedLogger struct {
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func newCustomizedLogger(opts ...FuncOpts) logger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	cl := &customizedLogger{
		Config: logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
	for _, opt := range opts {
		opt(cl)
	}

	if cl.Colorful {
		cl.infoStr = logger.Green + "%s\n" + logger.Reset + logger.Green + "[info] " + logger.Reset
		cl.warnStr = logger.BlueBold + "%s\n" + logger.Reset + logger.Magenta + "[warn] " + logger.Reset
		cl.errStr = logger.Magenta + "%s\n" + logger.Reset + logger.Red + "[error] " + logger.Reset
		cl.traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		cl.traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		cl.traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}
	return cl
}

func (l *customizedLogger) LogMode(level logger.LogLevel) logger.Interface {
	nl := *l
	nl.LogLevel = level
	return &nl
}

func (l customizedLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel < logger.Info {
		return
	}
	logrus.WithContext(ctx).Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l customizedLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel < logger.Warn {
		return
	}
	logrus.WithContext(ctx).Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l customizedLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel < logger.Error {
		return
	}
	logrus.WithContext(ctx).Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l customizedLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			logrus.WithContext(ctx).Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logrus.WithContext(ctx).Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			logrus.WithContext(ctx).Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logrus.WithContext(ctx).Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			logrus.WithContext(ctx).Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logrus.WithContext(ctx).Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
