package log

import (
	"go.uber.org/zap"
)

var DefaultLogger Logger

func init() {
	var err error
	config := &LogConfig{
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
	if DefaultLogger, err = New(config); err != nil {
		panic("init default log: " + err.Error())
	}
}

func With(fields ...zap.Field) Logger {
	return DefaultLogger.With(fields...)
}

func Debug(args ...any) {
	DefaultLogger.Debug(args...)
}

func Debugf(formatter string, args ...any) {
	DefaultLogger.Debugf(formatter, args...)
}

func DebugFields(msg string, fields ...zap.Field) {
	DefaultLogger.DebugFields(msg, fields...)
}

func Info(args ...any) {
	DefaultLogger.Info(args...)
}

func Infof(formatter string, args ...any) {
	DefaultLogger.Infof(formatter, args...)
}

func InfoFields(msg string, fields ...zap.Field) {
	DefaultLogger.InfoFields(msg, fields...)
}

func Warn(args ...any) {
	DefaultLogger.Warn(args...)
}

func Warnf(formatter string, args ...any) {
	DefaultLogger.Warnf(formatter, args...)
}

func WarnFields(msg string, fields ...zap.Field) {
	DefaultLogger.WarnFields(msg, fields...)
}

func Error(args ...any) {
	DefaultLogger.Error(args...)
}

func Errorf(formatter string, args ...any) {
	DefaultLogger.Errorf(formatter, args...)
}

func ErrorFields(msg string, fields ...zap.Field) {
	DefaultLogger.ErrorFields(msg, fields...)
}

func Fatal(args ...any) {
	DefaultLogger.Fatal(args...)
}

func Fatalf(formatter string, args ...any) {
	DefaultLogger.Fatalf(formatter, args...)
}

func FatalFields(msg string, fields ...zap.Field) {
	DefaultLogger.FatalFields(msg, fields...)
}

func Sync() error {
	return DefaultLogger.Sync()
}
