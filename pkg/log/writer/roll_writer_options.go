package writer

// Options is the rollWriter call options.
type Options struct {
	// MaxSize is max size by byte of the log file.
	MaxSize int64

	// MaxBackups is the max number of log files.
	MaxBackups int

	// MaxAge is the max expire time by day of log files.
	MaxAge int

	// whether the log file should be compressed.
	Compress bool

	// TimeFormat is the time format to split log file by time.
	TimeFormat string

	// CloseFileDelay is the delay duration before closing rotated files (in milliseconds)
	CloseFileDelay int

	// CloseFileChanSize is the size of channel for delayed file closing
	CloseFileChanSize int
}

// Option modifies the Options.
type Option func(*Options)

// WithMaxSize returns an Option which sets the max size(MB) of log files.
func WithMaxSize(n int) Option {
	return func(o *Options) {
		o.MaxSize = int64(n) * 1024 * 1024
	}
}

// WithMaxAge returns an Option which sets the max expire time(Day) of log files.
func WithMaxAge(n int) Option {
	return func(o *Options) {
		o.MaxAge = n
	}
}

// WithMaxBackups returns an Option which sets the max number of backup log files.
func WithMaxBackups(n int) Option {
	return func(o *Options) {
		o.MaxBackups = n
	}
}

// WithCompress returns an Option which sets whether log files should be compressed.
func WithCompress(b bool) Option {
	return func(o *Options) {
		o.Compress = b
	}
}

// WithRotationTime returns an Option which sets the time format(%Y%m%d) to roll logs.
func WithRotationTime(s string) Option {
	return func(o *Options) {
		o.TimeFormat = s
	}
}

// WithCloseFileDelay returns an Option which sets the delay time in milliseconds before closing rotated files.
func WithCloseFileDelay(n int) Option {
	return func(o *Options) {
		o.CloseFileDelay = n
	}
}

// WithCloseFileChanSize returns an Option which sets the size of channel for delayed file closing.
func WithCloseFileChanSize(n int) Option {
	return func(o *Options) {
		o.CloseFileChanSize = n
	}
}
