package log

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logTypeConsole = "console"
	logTypeFile    = "file"
)

type ZapLogger struct {
	l *zap.Logger
}

func (z *ZapLogger) Named(s string) Logger {
	return &ZapLogger{
		l: z.l.Named(s),
	}
}

func (z *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{
		l: z.l.With(fields...),
	}
}

func (z *ZapLogger) Debug(args ...interface{}) {
	z.l.Debug(fmt.Sprint(args...))
}

func (z *ZapLogger) Debugf(formatter string, args ...interface{}) {
	z.l.Debug(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) DebugFields(msg string, fields ...zap.Field) {
	z.l.Debug(msg, fields...)
}

func (z *ZapLogger) Info(args ...interface{}) {
	z.l.Info(fmt.Sprint(args...))
}

func (z *ZapLogger) Infof(formatter string, args ...interface{}) {
	z.l.Info(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) InfoFields(msg string, fields ...zap.Field) {
	z.l.Info(msg, fields...)
}

func (z *ZapLogger) Warn(args ...interface{}) {
	z.l.Warn(fmt.Sprint(args...))
}

func (z *ZapLogger) Warnf(formatter string, args ...interface{}) {
	z.l.Warn(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) WarnFields(msg string, fields ...zap.Field) {
	z.l.Warn(msg, fields...)
}

func (z *ZapLogger) Error(args ...interface{}) {
	z.l.Error(fmt.Sprint(args...))
}

func (z *ZapLogger) Errorf(formatter string, args ...interface{}) {
	z.l.Error(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) ErrorFields(msg string, fields ...zap.Field) {
	z.l.Error(msg, fields...)
}

func (z *ZapLogger) Fatal(args ...interface{}) {
	z.l.Fatal(fmt.Sprint(args...))
}

func (z *ZapLogger) Fatalf(formatter string, args ...interface{}) {
	z.l.Fatal(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) FatalFields(msg string, fields ...zap.Field) {
	z.l.Fatal(msg, fields...)
}

func (z *ZapLogger) Sync() error {
	return z.l.Sync()
}

// Levels is the map from string to zapcore.Level.
var zapCoreLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

func NewZapLog(c LogConfig) (Logger, error) {
	var cores []zapcore.Core
	for _, cfg := range c.CoreConfig {
		if cfg.Writer == logTypeConsole {
			core := newConsoleCore(cfg)
			cores = append(cores, core)
		} else if cfg.Writer == logTypeFile {
			core, err := newFileCore(cfg)
			if err != nil {
				return nil, errors.New("new file core: " + err.Error())
			}
			cores = append(cores, core)
		} else {
			return nil, errors.New("writer type " + cfg.Writer + " not support")
		}
	}
	return &ZapLogger{
		l: zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(c.CallerSkip)),
	}, nil
}

func newConsoleCore(c CoreConfig) zapcore.Core {
	level := zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level])
	core := zapcore.NewCore(newEncoder(c), zapcore.Lock(os.Stdout), level)
	return core
}

func newFileCore(c CoreConfig) (zapcore.Core, error) {
	opts := []Option{
		WithMaxSize(c.WriteConfig.MaxSize),
		WithMaxBackups(c.WriteConfig.MaxBackups),
		WithMaxAge(c.WriteConfig.MaxAge),
		WithCompress(c.WriteConfig.Compress),
		WithRotationTime(c.WriteConfig.TimeFormat),
	}
	writer, err := NewRollWriter(c.WriteConfig.FileName, opts...)
	if err != nil {
		return nil, errors.New("new roll writer: " + err.Error())
	}
	ws := zapcore.AddSync(writer)

	level := zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level])
	core := zapcore.NewCore(newEncoder(c), ws, level)
	return core, nil
}

func newEncoder(c CoreConfig) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:       getLogEncoderKey("Time", c.FormatConfig.TimeKey),
		LevelKey:      getLogEncoderKey("Level", c.FormatConfig.LevelKey),
		NameKey:       getLogEncoderKey("Name", c.FormatConfig.NameKey),
		CallerKey:     getLogEncoderKey("Caller", c.FormatConfig.CallerKey),
		FunctionKey:   getLogEncoderKey(zapcore.OmitKey, c.FormatConfig.FunctionKey),
		MessageKey:    getLogEncoderKey("Message", c.FormatConfig.MessageKey),
		StacktraceKey: getLogEncoderKey("Stacktrace", c.FormatConfig.StacktraceKey),
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendByteString(defaultTimeFormat(t))
		},
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}
	switch c.Formatter {
	case "console":
		return zapcore.NewConsoleEncoder(cfg)
	case "json":
		return zapcore.NewJSONEncoder(cfg)
	default:
		return zapcore.NewConsoleEncoder(cfg)
	}
}

func getLogEncoderKey(defKey, key string) string {
	if key == "" {
		return defKey
	}
	return key
}

// defaultTimeFormat returns the default time formatter.
func defaultTimeFormat(t time.Time) []byte {
	t = t.Local()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	micros := t.Nanosecond() / 1000

	buf := make([]byte, 23)
	buf[0] = byte((year/1000)%10) + '0'
	buf[1] = byte((year/100)%10) + '0'
	buf[2] = byte((year/10)%10) + '0'
	buf[3] = byte(year%10) + '0'
	buf[4] = '-'
	buf[5] = byte((month)/10) + '0'
	buf[6] = byte((month)%10) + '0'
	buf[7] = '-'
	buf[8] = byte((day)/10) + '0'
	buf[9] = byte((day)%10) + '0'
	buf[10] = ' '
	buf[11] = byte((hour)/10) + '0'
	buf[12] = byte((hour)%10) + '0'
	buf[13] = ':'
	buf[14] = byte((minute)/10) + '0'
	buf[15] = byte((minute)%10) + '0'
	buf[16] = ':'
	buf[17] = byte((second)/10) + '0'
	buf[18] = byte((second)%10) + '0'
	buf[19] = '.'
	buf[20] = byte((micros/100000)%10) + '0'
	buf[21] = byte((micros/10000)%10) + '0'
	buf[22] = byte((micros/1000)%10) + '0'
	return buf
}
