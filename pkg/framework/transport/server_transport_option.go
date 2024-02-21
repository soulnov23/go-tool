package transport

type ServerTransportOptions struct {
	loopSize  int
	eventSize int
	backlog   int
}

type ServerTransportOption func(*ServerTransportOptions)

func WithLoopSize(loopSize int) ServerTransportOption {
	return func(o *ServerTransportOptions) {
		o.loopSize = loopSize
	}
}

func WithEventSize(eventSize int) ServerTransportOption {
	return func(o *ServerTransportOptions) {
		o.eventSize = eventSize
	}
}

func WithBacklog(backlog int) ServerTransportOption {
	return func(o *ServerTransportOptions) {
		o.backlog = backlog
	}
}
