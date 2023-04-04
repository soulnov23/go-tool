package log

import "go.uber.org/zap"

type PrintfFunc func(formatter string, args ...any)

type Logger interface {
	Named(s string) Logger
	With(fields ...zap.Field) Logger
	Debug(args ...any)
	Debugf(formatter string, args ...any)
	DebugFields(msg string, fields ...zap.Field)
	Info(args ...any)
	Infof(formatter string, args ...any)
	InfoFields(msg string, fields ...zap.Field)
	Warn(args ...any)
	Warnf(formatter string, args ...any)
	WarnFields(msg string, fields ...zap.Field)
	Error(args ...any)
	Errorf(formatter string, args ...any)
	ErrorFields(msg string, fields ...zap.Field)
	Fatal(args ...any)
	Fatalf(formatter string, args ...any)
	FatalFields(msg string, fields ...zap.Field)
	Sync() error
}
