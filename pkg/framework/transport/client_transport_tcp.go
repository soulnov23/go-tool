package transport

import (
	"context"
)

var cltTranUDP = &clientTransportUDP{}

func init() {
	RegisterClientTransport("udp", cltTranUDP)
	RegisterClientTransport("udp4", cltTranUDP)
	RegisterClientTransport("udp6", cltTranUDP)
}

type clientTransportUDP struct{}

func (t *clientTransportUDP) Invoke(ctx context.Context, request []byte) (response []byte, err error) {
	return nil, nil
}
