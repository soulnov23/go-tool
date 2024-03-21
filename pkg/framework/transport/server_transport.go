package transport

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type serverTransportFunc func(network, address, protocol string, opts ...ServerTransportOption) (ServerTransport, error)

var (
	serverTransportFuncs = map[string]serverTransportFunc{}
	sMutex               = sync.RWMutex{}
)

type ServerTransport interface {
	ListenAndServe(ctx context.Context) error
}

func RegisterServerTransportFunc(network string, fn serverTransportFunc) {
	value := reflect.ValueOf(fn)
	if fn == nil || value.Kind() == reflect.Pointer && value.IsNil() {
		panic("register nil server transport")
	}
	if network == "" {
		panic("register empty network of server transport")
	}
	sMutex.Lock()
	defer sMutex.Unlock()
	serverTransportFuncs[network] = fn
}

func NewServerTransport(network, address, protocol string, opts ...ServerTransportOption) (ServerTransport, error) {
	fn, ok := serverTransportFuncs[network]
	if !ok {
		return nil, fmt.Errorf("network[%s] not support", network)
	}
	return fn(network, address, protocol, opts...)
}
