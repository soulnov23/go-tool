package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	clog, err := New(ConsoleConfig)
	if err != nil {
		t.Logf("new log: %s", err.Error())
		return
	}
	clog = clog.With(zap.String("name", "clog"))
	clog.Debug("hello world")
	clog.Debugf("%s %s", "hello", "world")
	clog.DebugFields("hello world", zap.String("hello", "world"))

	jlog, err := New(JsonConfig)
	if err != nil {
		t.Logf("new log: %s", err.Error())
		return
	}
	jlog = jlog.With(zap.String("name", "jlog"))
	jlog.Debug("hello world")
	jlog.Debugf("%s %s", "hello", "world")
	jlog.DebugFields("hello world", zap.String("hello", "world"))
	jlog.DebugFields("hello world", zap.Reflect("meta", nil))
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

	jlog, err := New(config)
	if err != nil {
		t.Logf("new log: %s", err.Error())
		return
	}
	jlog = jlog.With(zap.String("name", "jlog"))

	for {
		jlog.DebugFields("hello world", zap.String("hello", "world"))
	}

}

func TestStack(t *testing.T) {
	jlog, err := New(JsonConfig)
	if err != nil {
		t.Logf("new log: %s", err.Error())
		return
	}
	jlog = jlog.With(zap.String("name", "jlog"))
	defer func() {
		if err := recover(); err != nil {
			jlog.DebugFields("[PANIC]", zap.Stack("stack"))
			jlog.Errorf("[PANIC] %v", err)
		}
	}()
	panic("test")
}
