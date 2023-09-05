package mysql

import "time"

type Options struct {
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大打开连接数
	ConnMaxLifetime time.Duration // 连接重用最大时间
	ConnMaxIdleTime time.Duration // 连接空闲最大时间
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
