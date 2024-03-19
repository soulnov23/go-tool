package transport

import (
	"context"
	"errors"
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
func RegisterClientTransport(network string, t ClientTransport) error {
	tv := reflect.ValueOf(t)
	if t == nil || tv.Kind() == reflect.Pointer && tv.IsNil() {
		return errors.New("register nil client transport")
	}
	if network == "" {
		return errors.New("register empty network of client transport")
	}
	muxCltTrans.Lock()
	cltTrans[network] = t
	muxCltTrans.Unlock()
	return nil
}

// GetClientTransport gets the ClientTransport.
func GetClientTransport(network string) ClientTransport {
	muxCltTrans.RLock()
	t := cltTrans[network]
	muxCltTrans.RUnlock()
	return t
}
