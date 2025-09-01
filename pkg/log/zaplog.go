package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lestrrat-go/strftime"
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

func New(c *Config) (Logger, error) {
	var cores []zapcore.Core
	for _, cfg := range c.CoreConfig {
		if cfg == nil {
			return nil, errors.New("core config is nil")
		}
		if cfg.Formatter == "json" && cfg.FormatConfig == nil {
			return nil, errors.New("format config is nil")
		}
		if cfg.Writer == logTypeFile && cfg.WriteConfig == nil {
			return nil, errors.New("write config is nil")
		}
		switch cfg.Writer {
		case logTypeConsole:
			core := newConsoleCore(cfg)
			cores = append(cores, core)
		case logTypeFile:
			core, err := newFileCore(cfg)
			if err != nil {
				return nil, errors.New("new file core: " + err.Error())
			}
			cores = append(cores, core)
		default:
			return nil, fmt.Errorf("writer type[%s] not support", cfg.Writer)
		}
	}
	return &ZapLogger{
		l: zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(c.CallerSkip), zap.AddStacktrace(zapcore.FatalLevel)),
	}, nil
}

func newConsoleCore(c *CoreConfig) zapcore.Core {
	return zapcore.NewCore(newEncoder(c), zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level]))
}

func newFileCore(c *CoreConfig) (zapcore.Core, error) {
	if err := os.MkdirAll(filepath.Dir(c.WriteConfig.FileName), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %v", err)
	}
	pattern, err := strftime.New(c.WriteConfig.FileName + c.WriteConfig.TimeFormat)
	if err != nil {
		return nil, fmt.Errorf("get file pattern: %v", err)
	}
	file, err := os.OpenFile(pattern.FormatString(time.Now()), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("open log file: %v", err)
	}
	return zapcore.NewCore(newEncoder(c), zapcore.Lock(file), zap.NewAtomicLevelAt(zapCoreLevelMap[c.Level])), nil
}

func newEncoder(c *CoreConfig) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:          getLogEncoderKey("time", c.FormatConfig.TimeKey),
		LevelKey:         getLogEncoderKey("level", c.FormatConfig.LevelKey),
		NameKey:          getLogEncoderKey("name", c.FormatConfig.NameKey),
		CallerKey:        getLogEncoderKey("caller", c.FormatConfig.CallerKey),
		FunctionKey:      c.FormatConfig.FunctionKey,
		MessageKey:       getLogEncoderKey("msg", c.FormatConfig.MessageKey),
		StacktraceKey:    getLogEncoderKey("stack", c.FormatConfig.StacktraceKey),
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout(time.DateTime + ".000"),
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
