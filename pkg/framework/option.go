package framework

import "time"

type Options struct {
	ServiceName string
	Network     string
	Address     string
	Protocol    string
	Timeout     time.Duration
}

type Option func(*Options)

func WithServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

func WithAddress(address string) Option {
	return func(o *Options) {
		o.Address = address
	}
}

func WithNetwork(network string) Option {
	return func(o *Options) {
		o.Network = network
	}
}

func WithProtocol(protocol string) Option {
	return func(o *Options) {
		o.Protocol = protocol
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}
