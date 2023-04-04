package log

import "go.uber.org/zap"

type PrintfFunc func(formatter string, args ...interface{})

type Logger interface {
	Named(s string) Logger
	With(fields ...zap.Field) Logger
	Debug(args ...interface{})
	Debugf(formatter string, args ...interface{})
	DebugFields(msg string, fields ...zap.Field)
	Info(args ...interface{})
	Infof(formatter string, args ...interface{})
	InfoFields(msg string, fields ...zap.Field)
	Warn(args ...interface{})
	Warnf(formatter string, args ...interface{})
	WarnFields(msg string, fields ...zap.Field)
	Error(args ...interface{})
	Errorf(formatter string, args ...interface{})
	ErrorFields(msg string, fields ...zap.Field)
	Fatal(args ...interface{})
	Fatalf(formatter string, args ...interface{})
	FatalFields(msg string, fields ...zap.Field)
	Sync() error
}
