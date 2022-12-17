package net

type Options struct {
	loopSize  int
	eventSize int
	backlog   int
}

type Option func(*Options)

func WithLoopSize(loopSize int) Option {
	return func(o *Options) {
		o.loopSize = loopSize
	}
}

func WithEventSize(eventSize int) Option {
	return func(o *Options) {
		o.eventSize = eventSize
	}
}

func WithBacklog(backlog int) Option {
	return func(o *Options) {
		o.backlog = backlog
	}
}
