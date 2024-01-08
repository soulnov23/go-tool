package log

// Config is the output config, includes console, file and remote.
type Config struct {
	// CallerSkip controls the nesting depth of log function.
	CallerSkip int           `yaml:"caller_skip"`
	CoreConfig []*CoreConfig `yaml:"core_config"`
}

// CoreConfig is console, file output config.
type CoreConfig struct {
	// Level controls the log level, like debug, info or error.
	Level string

	// Formatter is the format of log, such as console or json.
	Formatter    string
	FormatConfig *FormatConfig `yaml:"formatter_config"`

	// Writer is the output of log, such as console or file.
	Writer      string
	WriteConfig *WriteConfig `yaml:"writer_config"`
}

// WriteConfig is the local file config.
type WriteConfig struct {
	// FileName is the file name like /var/run/log/server.log.
	FileName string `yaml:"file_name"`
	// TimeFormat is the time format to split log file by time.
	TimeFormat string `yaml:"time_format"`
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
	// TimeKey is the time key of log output, default as "time".
	TimeKey string `yaml:"time_key"`
	// LevelKey is the level key of log output, default as "level".
	LevelKey string `yaml:"level_key"`
	// NameKey is the name key of log output, default as "name".
	NameKey string `yaml:"name_key"`
	// CallerKey is the caller key of log output, default as "caller".
	CallerKey string `yaml:"caller_key"`
	// FunctionKey is the function key of log output, default as "", which means not to print function name.
	FunctionKey string `yaml:"function_key"`
	// MessageKey is the message key of log output, default as "msg".
	MessageKey string `yaml:"message_key"`
	// StackTraceKey is the stack trace key of log output, default as "", which means not to print stacktrace name.
	StacktraceKey string `yaml:"stacktrace_key"`
}
