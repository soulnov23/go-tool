package log

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/soulnov23/go-tool/pkg/log/writer"
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

func (z *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{
		l: z.l.With(fields...),
	}
}

func (z *ZapLogger) Debug(args ...any) {
	z.l.Debug(fmt.Sprint(args...))
}

func (z *ZapLogger) Debugf(formatter string, args ...any) {
	z.l.Debug(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) DebugFields(msg string, fields ...zap.Field) {
	z.l.Debug(msg, fields...)
}

func (z *ZapLogger) Info(args ...any) {
	z.l.Info(fmt.Sprint(args...))
}

func (z *ZapLogger) Infof(formatter string, args ...any) {
	z.l.Info(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) InfoFields(msg string, fields ...zap.Field) {
	z.l.Info(msg, fields...)
}

func (z *ZapLogger) Warn(args ...any) {
	z.l.Warn(fmt.Sprint(args...))
}

func (z *ZapLogger) Warnf(formatter string, args ...any) {
	z.l.Warn(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) WarnFields(msg string, fields ...zap.Field) {
	z.l.Warn(msg, fields...)
}

func (z *ZapLogger) Error(args ...any) {
	z.l.Error(fmt.Sprint(args...))
}

func (z *ZapLogger) Errorf(formatter string, args ...any) {
	z.l.Error(fmt.Sprintf(formatter, args...))
}

func (z *ZapLogger) ErrorFields(msg string, fields ...zap.Field) {
	z.l.Error(msg, fields...)
}

func (z *ZapLogger) Fatal(args ...any) {
	z.l.Fatal(fmt.Sprint(args...))
}

func (z *ZapLogger) Fatalf(formatter string, args ...any) {
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
	"":      zapcore.DebugLevel,
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

func New(c LogConfig) (Logger, error) {
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
		l: zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(c.CallerSkip), zap.AddStacktrace(zapcore.ErrorLevel)),
	}, nil
}

func newConsoleCore(c CoreConfig) zapcore.Core {
	return zapcore.NewCore(newEncoder(c), zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level]))
}

func newFileCore(c CoreConfig) (zapcore.Core, error) {
	opts := []writer.Option{
		writer.WithMaxSize(c.WriteConfig.MaxSize),
		writer.WithMaxBackups(c.WriteConfig.MaxBackups),
		writer.WithMaxAge(c.WriteConfig.MaxAge),
		writer.WithCompress(c.WriteConfig.Compress),
		writer.WithRotationTime(c.WriteConfig.TimeFormat),
	}
	writer, err := writer.New(c.WriteConfig.FileName, opts...)
	if err != nil {
		return nil, errors.New("new roll writer: " + err.Error())
	}
	ws := zapcore.AddSync(writer)
	return zapcore.NewCore(newEncoder(c), ws, zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level])), nil
}

func newEncoder(c CoreConfig) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:       getLogEncoderKey("time", c.FormatConfig.TimeKey),
		LevelKey:      getLogEncoderKey("level", c.FormatConfig.LevelKey),
		NameKey:       getLogEncoderKey("name", c.FormatConfig.NameKey),
		CallerKey:     getLogEncoderKey("caller", c.FormatConfig.CallerKey),
		FunctionKey:   c.FormatConfig.FunctionKey,
		MessageKey:    getLogEncoderKey("msg", c.FormatConfig.MessageKey),
		StacktraceKey: c.FormatConfig.StacktraceKey,
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
