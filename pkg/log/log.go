package log

import (
	"errors"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

const (
	logTypeConsole = "console"
	logTypeFile    = "file"
)

// Levels is the map from string to zapcore.Level.
var zapCoreLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

func NewZapLog(c LogConfig) (*zap.SugaredLogger, error) {
	var cores []zapcore.Core
	for _, cfg := range c.CoreConfig {
		if cfg.Writer == logTypeConsole {
			core := newConsoleCore(cfg)
			cores = append(cores, core)
		} else if cfg.Writer == logTypeFile {
			core, err := newFileCore(cfg)
			if err != nil {
				return nil, errors.New("log.newFileCore: " + err.Error())
			}
			cores = append(cores, core)
		} else {
			return nil, errors.New("writer type " + cfg.Writer + " not support")
		}
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCallerSkip(c.CallerSkip), zap.AddCaller()).Sugar(), nil
}

func newConsoleCore(c CoreConfig) zapcore.Core {
	level := zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level])
	core := zapcore.NewCore(newEncoder(c), zapcore.Lock(os.Stdout), level)
	return core
}

func newFileCore(c CoreConfig) (zapcore.Core, error) {
	opts := []Option{
		WithMaxAge(c.WriteConfig.MaxAge),
		WithMaxBackups(c.WriteConfig.MaxBackups),
		WithCompress(c.WriteConfig.Compress),
		WithMaxSize(c.WriteConfig.MaxSize),
		WithRotationTime(".%Y-%m-%d"),
	}
	writer, err := NewRollWriter(c.WriteConfig.FileName, opts...)
	if err != nil {
		return nil, errors.New("log.NewRollWriter: " + err.Error())
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

// defaultTimeFormat returns the default time format.
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
