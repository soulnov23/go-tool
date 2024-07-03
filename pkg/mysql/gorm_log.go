package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/soulnov23/go-tool/pkg/log"
	"gorm.io/gorm/logger"
)

const (
	infoFormatter  = "[%.3fms] [rows:%v] %s"
	warnFormatter  = "[%s] [%.3fms] [rows:%v] %s"
	errorFormatter = "[%s] [%.3fms] [rows:%v] %s"
)

// New initialize gormLogger
func new(logger log.Logger, opts ...Option) logger.Interface {
	defaultOpts := &Options{
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      false,
	}
	for _, opt := range opts {
		opt(defaultOpts)
	}
	return &gormLogger{
		Logger:  logger,
		Options: defaultOpts,
	}
}

type gormLogger struct {
	log.Logger
	*Options
}

// LogMode log mode
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info print info
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Infof(msg, data...)
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Warnf(msg, data...)
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Errorf(msg, data...)
}

// Trace print sql message
//
//nolint:cyclop
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		if rows == -1 {
			l.Logger.Errorf(errorFormatter, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Errorf(errorFormatter, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		return
	}

	if elapsed > l.SlowThreshold && l.SlowThreshold != 0 {
		if rows == -1 {
			l.Logger.Warnf(warnFormatter, fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Warnf(warnFormatter, fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		return
	}

	if rows == -1 {
		l.Logger.Infof(infoFormatter, float64(elapsed.Nanoseconds())/1e6, "-", sql)
	} else {
		l.Logger.Infof(infoFormatter, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}

// ParamsFilter filter params
func (l *gormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Options.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
