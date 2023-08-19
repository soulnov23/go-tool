package pipeline

import "time"

type Options struct {
	timeout time.Duration
}

type Option func(*Options)

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.timeout = timeout
	}
}
