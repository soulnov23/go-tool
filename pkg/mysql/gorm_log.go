package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/soulnov23/go-tool/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Logger.Errorf("[%s] [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Errorf("[%s] [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Logger.Warnf("[%s] [%.3fms] [rows:%v] %s", slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Warnf("[%s] [%.3fms] [rows:%v] %s", slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			l.Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter filter params
func (l *gormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Options.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
