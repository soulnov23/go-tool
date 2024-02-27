package transport

import (
	"context"
	"reflect"
	"sync"

	"github.com/soulnov23/go-tool/pkg/errors"
	"github.com/soulnov23/go-tool/pkg/framework"
)

type serverTransportFunc func(network, address, protocol string, opts ...ServerTransportOption) (ServerTransport, error)

var (
	serverTransportFuncs = map[string]serverTransportFunc{}
	sMutex               = sync.RWMutex{}
)

type ServerTransport interface {
	ListenAndServe(ctx context.Context)
}

func RegisterServerTransportFunc(network string, fn serverTransportFunc) error {
	value := reflect.ValueOf(fn)
	if fn == nil || value.Kind() == reflect.Pointer && value.IsNil() {
		return errors.NewInternalServerError(framework.Unknown, "register nil server transport")
	}
	if network == "" {
		return errors.NewInternalServerError(framework.Unknown, "register empty network of server transport")
	}
	sMutex.Lock()
	defer sMutex.Unlock()
	serverTransportFuncs[network] = fn
	return nil
}

func NewServerTransport(network, address, protocol string, opts ...ServerTransportOption) (ServerTransport, error) {
	fn, ok := serverTransportFuncs[network]
	if !ok {
		return nil, errors.NewInternalServerError(framework.InvalidConfig, "network[%s] not support", network)
	}
	return fn(network, address, protocol, opts...)
}
