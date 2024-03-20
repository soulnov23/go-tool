package transport

type ServerTransportOptions struct {
	loopSize int
}

type ServerTransportOption func(*ServerTransportOptions)

func WithLoopSize(loopSize int) ServerTransportOption {
	return func(o *ServerTransportOptions) {
		o.loopSize = loopSize
	}
}
