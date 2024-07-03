package mysql

import "time"

type Options struct {
	MaxIdleConns              int           // 最大空闲连接数
	MaxOpenConns              int           // 最大打开连接数
	ConnMaxLifetime           time.Duration // 连接重用最大时间
	ConnMaxIdleTime           time.Duration // 连接空闲最大时间
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
}

type Option func(*Options)

func WithMaxIdleConns(n int) Option {
	return func(o *Options) {
		o.MaxIdleConns = n
	}
}

func WithMaxOpenConns(n int) Option {
	return func(o *Options) {
		o.MaxOpenConns = n
	}
}

func WithConnMaxLifetime(t time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxLifetime = t
	}
}

func WithConnMaxIdleTime(t time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxIdleTime = t
	}
}

func WithSlowThreshold(t time.Duration) Option {
	return func(o *Options) {
		o.SlowThreshold = t
	}
}

func WithIgnoreRecordNotFoundError(b bool) Option {
	return func(o *Options) {
		o.IgnoreRecordNotFoundError = b
	}
}

func WithParameterizedQueries(b bool) Option {
	return func(o *Options) {
		o.ParameterizedQueries = b
	}
}
