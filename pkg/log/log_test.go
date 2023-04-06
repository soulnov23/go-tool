package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	clog, err := NewZapLog(ConsoleConfig)
	if err != nil {
		t.Logf("NewZapLog: %s", err.Error())
		return
	}
	clog = clog.Named("clog")
	clog.Debug("hello world")
	clog.Debugf("%s %s", "hello", "world")
	clog.DebugFields("hello world", zap.String("hello", "world"))

	jlog, err := NewZapLog(JsonConfig)
	if err != nil {
		t.Logf("NewZapLog: %s", err.Error())
		return
	}
	jlog = jlog.Named("jlog")
	jlog.Debug("hello world")
	jlog.Debugf("%s %s", "hello", "world")
	jlog.DebugFields("hello world", zap.String("hello", "world"))
}

func TestCutLog(t *testing.T) {
	config := LogConfig{
		CallerSkip: 1,
		CoreConfig: []CoreConfig{
			{
				Level:     "debug",
				Formatter: "json",
				FormatConfig: FormatConfig{
					TimeKey:       "time",
					LevelKey:      "level",
					NameKey:       "name",
					CallerKey:     "caller",
					FunctionKey:   "func",
					MessageKey:    "msg",
					StacktraceKey: "stack",
				},
				Writer: logTypeConsole,
			},
			{
				Level:     "debug",
				Formatter: "json",
				FormatConfig: FormatConfig{
					TimeKey:       "time",
					LevelKey:      "level",
					NameKey:       "name",
					CallerKey:     "caller",
					FunctionKey:   "func",
					MessageKey:    "msg",
					StacktraceKey: "stack",
				},
				Writer: logTypeFile,
				WriteConfig: WriteConfig{
					FileName:   "run.log",
					TimeFormat: ".%Y-%m-%d",
					MaxSize:    1,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
		},
	}

	jlog, err := NewZapLog(config)
	if err != nil {
		t.Logf("NewZapLog: %s", err.Error())
		return
	}
	jlog = jlog.Named("jlog")

	for {
		jlog.DebugFields("hello world", zap.String("hello", "world"))
	}
}
