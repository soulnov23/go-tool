package transport

import (
	"context"
)

var cltTranTCP = &clientTransportTCP{}

func init() {
	RegisterClientTransport("tcp", cltTranTCP)
	RegisterClientTransport("tcp4", cltTranTCP)
	RegisterClientTransport("tcp6", cltTranTCP)
}

type clientTransportTCP struct{}

func (t *clientTransportTCP) Invoke(ctx context.Context, request []byte) (response []byte, err error) {
	return nil, nil
}
