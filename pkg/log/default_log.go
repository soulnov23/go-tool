package log

import "fmt"

func GetDefaultLogger() (Logger, error) {
	config := &Config{
		CallerSkip: 1,
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
				},
			},
		},
	}
	defaultLogger, err := New(config)
	if err != nil {
		return nil, fmt.Errorf("init default log: %v", err)
	}
	return defaultLogger, nil
}
