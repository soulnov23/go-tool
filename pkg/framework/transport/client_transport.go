package transport

import (
	"context"
	"reflect"
	"sync"
)

var (
	cltTrans    = make(map[string]ClientTransport) // network => ClientTransport
	muxCltTrans = sync.RWMutex{}
)

type ClientTransport interface {
	Invoke(ctx context.Context, request []byte) (response []byte, err error)
}

// RegisterClientTransport register a ClientTransport.
func RegisterClientTransport(network string, t ClientTransport) {
	tv := reflect.ValueOf(t)
	if t == nil || tv.Kind() == reflect.Pointer && tv.IsNil() {
		panic("register nil client transport")
	}
	if network == "" {
		panic("register empty network of client transport")
	}
	muxCltTrans.Lock()
	cltTrans[network] = t
	muxCltTrans.Unlock()
}

// GetClientTransport gets the ClientTransport.
func GetClientTransport(network string) ClientTransport {
	muxCltTrans.RLock()
	t := cltTrans[network]
	muxCltTrans.RUnlock()
	return t
}
