package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/soulnov23/go-tool/pkg/framework/transport"
)

type Handler func(ctx context.Context, request string) (response string, err error)

type service struct {
	name     string
	address  string
	network  string
	protocol string
	timeout  time.Duration

	serverTransport transport.ServerTransport

	handlers map[string]Handler // rpc_name => Handler
}

func newService(name, address, network, protocol string, timeout time.Duration) *service {
	return &service{
		name:     name,
		address:  address,
		network:  network,
		protocol: protocol,
		timeout:  timeout,
		handlers: make(map[string]Handler),
	}
}

func (s *service) register(rpcName string, handler Handler) error {
	s.handlers[rpcName] = handler
	return nil
}

func (s *service) serve() error {
	s.serverTransport = transport.NewServerTransport(s.address, s.network, s.protocol)
	if s.serverTransport == nil {
		return fmt.Errorf("network[%s] not support", s.network)
	}
	return s.serverTransport.ListenAndServe()
}

func (s *service) close() {
	s.serverTransport.Close()
}
