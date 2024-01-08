package log

import (
	"go.uber.org/zap"
)

var defaultLogger Logger

func init() {
	var err error
	config := &Config{
		CallerSkip: 2,
		CoreConfig: []*CoreConfig{
			{
				Level:     "debug",
				Formatter: "json",
				FormatConfig: &FormatConfig{
					TimeKey:       "time",
					LevelKey:      "level",
					NameKey:       "name",
					CallerKey:     "caller",
					FunctionKey:   "",
					MessageKey:    "msg",
					StacktraceKey: "stack",
				},
				Writer: logTypeConsole,
			},
			{
				Level:     "debug",
				Formatter: "json",
				FormatConfig: &FormatConfig{
					TimeKey:       "time",
					LevelKey:      "level",
					NameKey:       "name",
					CallerKey:     "caller",
					FunctionKey:   "",
					MessageKey:    "msg",
					StacktraceKey: "stack",
				},
				Writer: logTypeFile,
				WriteConfig: &WriteConfig{
					FileName:   "run.log",
					TimeFormat: ".%Y-%m-%d",
					MaxSize:    1,
					MaxBackups: 0,
					MaxAge:     0,
					Compress:   false,
				},
			},
		},
	}
	if defaultLogger, err = New(config); err != nil {
		panic("init default log: " + err.Error())
	}
}

func With(fields ...zap.Field) Logger {
	return defaultLogger.With(fields...)
}

func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

func Debugf(formatter string, args ...any) {
	defaultLogger.Debugf(formatter, args...)
}

func DebugFields(msg string, fields ...zap.Field) {
	defaultLogger.DebugFields(msg, fields...)
}

func Info(args ...any) {
	defaultLogger.Info(args...)
}

func Infof(formatter string, args ...any) {
	defaultLogger.Infof(formatter, args...)
}

func InfoFields(msg string, fields ...zap.Field) {
	defaultLogger.InfoFields(msg, fields...)
}

func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

func Warnf(formatter string, args ...any) {
	defaultLogger.Warnf(formatter, args...)
}

func WarnFields(msg string, fields ...zap.Field) {
	defaultLogger.WarnFields(msg, fields...)
}

func Error(args ...any) {
	defaultLogger.Error(args...)
}

func Errorf(formatter string, args ...any) {
	defaultLogger.Errorf(formatter, args...)
}

func ErrorFields(msg string, fields ...zap.Field) {
	defaultLogger.ErrorFields(msg, fields...)
}

func Fatal(args ...any) {
	defaultLogger.Fatal(args...)
}

func Fatalf(formatter string, args ...any) {
	defaultLogger.Fatalf(formatter, args...)
}

func FatalFields(msg string, fields ...zap.Field) {
	defaultLogger.FatalFields(msg, fields...)
}

func Sync() error {
	return defaultLogger.Sync()
}
