package log

// LogConfig is the output config, includes console, file and remote.
type LogConfig struct {
	// CallerSkip controls the nesting depth of log function.
	CallerSkip int          `yaml:"caller_skip"`
	CoreConfig []CoreConfig `yaml:"core_config"`
}

// CoreConfig is console, file output config.
type CoreConfig struct {
	// Level controls the log level, like debug, info or error.
	Level string

	// Formatter is the format of log, such as console or json.
	Formatter    string
	FormatConfig FormatConfig `yaml:"formatter_config"`

	// Writer is the output of log, such as console or file.
	Writer      string
	WriteConfig WriteConfig `yaml:"writer_config"`
}

// WriteConfig is the local file config.
type WriteConfig struct {
	// FileName is the file name like /var/run/log/server.log.
	FileName string `yaml:"file_name"`
	// MaxSize is the max size of log file(MB).
	MaxSize int `yaml:"max_size"`
	// MaxBackups is the max backup files.
	MaxBackups int `yaml:"max_backups"`
	// MaxAge is the max expire times(day).
	MaxAge int `yaml:"max_age"`
	// Compress defines whether log should be compressed.
	Compress bool `yaml:"compress"`
}

// FormatConfig is the log format config.
type FormatConfig struct {
	// TimeKey is the time key of log output, default as "Time".
	TimeKey string `yaml:"time_key"`
	// LevelKey is the level key of log output, default as "Level".
	LevelKey string `yaml:"level_key"`
	// NameKey is the name key of log output, default as "Name".
	NameKey string `yaml:"name_key"`
	// CallerKey is the caller key of log output, default as "Caller".
	CallerKey string `yaml:"caller_key"`
	// FunctionKey is the function key of log output, default as "", which means not to print function name.
	FunctionKey string `yaml:"function_key"`
	// MessageKey is the message key of log output, default as "Message".
	MessageKey string `yaml:"message_key"`
	// StackTraceKey is the stack trace key of log output, default as "Stacktrace".
	StacktraceKey string `yaml:"stacktrace_key"`
}

// 默认日志配置
var DefaultLogConfig = LogConfig{
	CallerSkip: 0,
	CoreConfig: []CoreConfig{
		{
			Level:     "debug",
			Formatter: "console",
			Writer:    logTypeConsole,
		},
		{
			Level:     "debug",
			Formatter: "console",
			Writer:    logTypeFile,
			WriteConfig: WriteConfig{
				FileName:   "app.log",
				MaxSize:    128,
				MaxBackups: 10,
				MaxAge:     7,
				Compress:   false,
			},
		},
	},
}

// 标准输出日志配置
var DefaultConsoleLogConfig = LogConfig{
	CallerSkip: 0,
	CoreConfig: []CoreConfig{
		{
			Level:     "debug",
			Formatter: "console",
			Writer:    logTypeConsole,
		},
	},
}

// 本地文件日志配置
var DefaultFileLogConfig = LogConfig{
	CallerSkip: 0,
	CoreConfig: []CoreConfig{
		{
			Level:     "debug",
			Formatter: "console",
			Writer:    logTypeFile,
			WriteConfig: WriteConfig{
				FileName:   "app.log",
				MaxSize:    128,
				MaxBackups: 10,
				MaxAge:     7,
				Compress:   false,
			},
		},
	},
}
