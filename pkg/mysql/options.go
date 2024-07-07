package mysql

import "time"

type Options struct {
	MaxIdleConns              int           // 最大空闲连接数
	MaxOpenConns              int           // 最大打开连接数
	ConnMaxLifetime           time.Duration // 连接重用最大时间
	ConnMaxIdleTime           time.Duration // 连接空闲最大时间
	SlowThreshold             time.Duration // TraceLog打印慢查询日志时的阈值
	IgnoreRecordNotFoundError bool          // TraceLog打印错误日志时是否忽略RecordNotFound错误
	ParameterizedQueries      bool          // TraceLog打印日志时SQL语句是否使用?占位符代替实际的参数
	DryRun                    bool          // 生成SQL但不执行
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

func WithDryRun(b bool) Option {
	return func(o *Options) {
		o.DryRun = b
	}
}
