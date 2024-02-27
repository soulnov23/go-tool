package codec

import "sync"

var (
	svrTrans    = make(map[string]ServerTransport) // network => ServerTransport
	muxSvrTrans = sync.RWMutex{}

	cltTrans    = make(map[string]ClientTransport) // network => ClientTransport
	muxCltTrans = sync.RWMutex{}
)

type Codec interface {
	// Encode pack the body into binary buffer.
	// client: Encode(msg, reqBody)(request-buffer, err)
	// server: Encode(msg, rspBody)(response-buffer, err)
	Encode(message Msg, body []byte) (buffer []byte, err error)

	// Decode unpack the body from binary buffer
	// server: Decode(msg, request-buffer)(reqBody, err)
	// client: Decode(msg, response-buffer)(rspBody, err)
	Decode(message Msg, buffer []byte) (body []byte, err error)
}
