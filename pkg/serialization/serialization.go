package serialization

import (
	"sync"
)

type Serializer interface {
	Unmarshal(in []byte, body interface{}) error
	Marshal(body interface{}) (out []byte, err error)
}

const (
	SerializationTypeJSON = 0
	SerializationTypeXML  = 1
	SerializationTypeYAML = 2
)

var (
	serializers = make(map[int]Serializer)
	rwLock      = sync.RWMutex{}
)

func RegisterSerializer(serializationType int, s Serializer) {
	rwLock.Lock()
	serializers[serializationType] = s
	rwLock.Unlock()
}

func GetSerializer(serializationType int) Serializer {
	rwLock.RLock()
	s := serializers[serializationType]
	rwLock.RUnlock()
	return s
}
