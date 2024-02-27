package framework

import (
	"context"
	"time"
)

type Handler func(ctx context.Context, request string) (response string, err error)

type service struct {
	name     string
	network  string
	address  string
	protocol string
	timeout  time.Duration

	ctx      context.Context
	cancel   context.CancelFunc
	handlers map[string]Handler // rpc_name => Handler
}

func newService(name, network, address, protocol string, timeout time.Duration) *service {
	ctx, cancel := context.WithCancel(context.Background())
	s := &service{
		name:     name,
		network:  network,
		address:  address,
		protocol: protocol,
		timeout:  timeout,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]Handler),
	}
	return s
}

func (s *service) register(rpcName string, handler Handler) error {
	s.handlers[rpcName] = handler
	return nil
}
