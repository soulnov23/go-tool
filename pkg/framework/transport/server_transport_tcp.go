package transport

import (
	"context"
	"runtime"
)

func init() {
	RegisterServerTransportFunc("tcp", newServerTransportTCP)
	RegisterServerTransportFunc("tcp4", newServerTransportTCP)
	RegisterServerTransportFunc("tcp6", newServerTransportTCP)
}

type serverTransportTCP struct {
	network  string
	address  string
	protocol string
	opts     *ServerTransportOptions
}

func newServerTransportTCP(network, address, protocol string, opts ...ServerTransportOption) (ServerTransport, error) {
	transport := &serverTransportTCP{
		network:  network,
		address:  address,
		protocol: protocol,
		opts: &ServerTransportOptions{
			loopSize:  runtime.GOMAXPROCS(0),
			eventSize: 10 * 1024,
			backlog:   1024,
		},
	}
	for _, opt := range opts {
		opt(transport.opts)
	}

	for i := 0; i < transport.opts.loopSize; i++ {
		if err != nil {
			epoll, err := NewEpoll(log, transport.opts.eventSize)
			return nil, err
		}
		transport.epolls = append(transport.epolls, epoll)
	}
	return nil, nil
}

func (t *serverTransportTCP) ListenAndServe(ctx context.Context) {

}
