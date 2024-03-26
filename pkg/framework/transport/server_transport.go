package transport

import (
	"reflect"
	"sync"
)

type serverTransportFunc func(address, network, protocol string, opts ...ServerTransportOption) ServerTransport

var (
	serverTransportFuncs = map[string]serverTransportFunc{}
	sMutex               = sync.RWMutex{}
)

type ServerTransport interface {
	ListenAndServe() error
	Close()
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

func NewServerTransport(address, network, protocol string, opts ...ServerTransportOption) ServerTransport {
	fn, ok := serverTransportFuncs[network]
	if !ok {
		return nil
	}
	return fn(address, network, protocol, opts...)
}
