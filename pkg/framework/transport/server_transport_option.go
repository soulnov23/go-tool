package transport

type ServerTransportOptions struct {
	loopSize  int
	eventSize int
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
