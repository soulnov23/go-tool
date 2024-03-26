package transport

type ServerTransportOptions struct {
	coreSize int
}

type ServerTransportOption func(*ServerTransportOptions)

func WithCoreSize(coreSize int) ServerTransportOption {
	return func(o *ServerTransportOptions) {
		o.coreSize = coreSize
	}
}
