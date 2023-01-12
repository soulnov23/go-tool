package log

// LogConfig is the output config, includes console, file and remote.
type LogConfig struct {
	// Level controls the log level, like debug, info or error.
	Level string

	// CallerSkip controls the nesting depth of log function.
	CallerSkip int `yaml:"caller_skip"`

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
	// TimeKey is the time key of log output, default as "T".
	TimeKey string `yaml:"time_key"`
	// LevelKey is the level key of log output, default as "L".
	LevelKey string `yaml:"level_key"`
	// NameKey is the name key of log output, default as "N".
	NameKey string `yaml:"name_key"`
	// CallerKey is the caller key of log output, default as "C".
	CallerKey string `yaml:"caller_key"`
	// FunctionKey is the function key of log output, default as "", which means not to print
	// function name.
	FunctionKey string `yaml:"function_key"`
	// MessageKey is the message key of log output, default as "M".
	MessageKey string `yaml:"message_key"`
	// StackTraceKey is the stack trace key of log output, default as "S".
	StacktraceKey string `yaml:"stacktrace_key"`
}

// 标准输出打印
var DefaultConsoleLogConfig = &LogConfig{
	Level:      "debug",
	CallerSkip: 0,
	Formatter:  "console",
	Writer:     logTypeConsole,
}

// 日志文件打印
var DefaultFileLogConfig = &LogConfig{
	Level:      "debug",
	CallerSkip: 0,
	Formatter:  "console",
	Writer:     logTypeFile,
	WriteConfig: WriteConfig{
		FileName:   "app.log",
		MaxSize:    128,
		MaxBackups: 10,
		MaxAge:     7,
		Compress:   false,
	},
}
