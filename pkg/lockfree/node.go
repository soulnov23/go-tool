package lockfree

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	value interface{}
	next  unsafe.Pointer
}

func load(addr *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(addr))
}

func cas(addr *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(addr, unsafe.Pointer(old), unsafe.Pointer(new))
}
