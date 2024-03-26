package framework

import (
	"context"
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

func newService(name, address, network, protocol string, timeout time.Duration) (*service, error) {
	serverTransport, err := transport.NewServerTransport(address, network, protocol)
	if err != nil {
		return nil, err
	}
	s := &service{
		name:            name,
		address:         address,
		network:         network,
		protocol:        protocol,
		timeout:         timeout,
		serverTransport: serverTransport,
		handlers:        make(map[string]Handler),
	}
	return s, nil
}

func (s *service) register(rpcName string, handler Handler) error {
	s.handlers[rpcName] = handler
	return nil
}

func (s *service) serve() error {
	return s.serverTransport.ListenAndServe()
}

func (s *service) close() {
	s.serverTransport.Close()
}
