package ring

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	cacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})
)

type node struct {
	seq   uint32
	value any
}

// 为了获得高性能，使用伪共享填充在多线程环境下确保read和write不共享相同的缓存行
type Ring struct {
	capacity uint32
	mask     uint32
	_        [cacheLinePadSize - 8]byte
	read     uint32
	_        [cacheLinePadSize - 4]byte
	write    uint32
	_        [cacheLinePadSize - 4]byte
	nodes    []*node
}

func New(capacity uint32) *Ring {
	capacity = roundUpToPower2(capacity)
	if capacity < 2 {
		capacity = 2
	}

	r := &Ring{
		capacity: capacity,
		mask:     capacity - 1,
		nodes:    make([]*node, capacity),
	}
	for index := range r.nodes {
		r.nodes[index].seq = uint32(index)
	}
	return r
}

func roundUpToPower2(v uint32) uint32 {
	// 非2的幂
	if v&(v-1) != 0 {
		// 依次将最高位1右边的第1位、第2~3位，第4~7位，第8~15位，第16~31位置为1
		v |= v >> 1
		v |= v >> 2
		v |= v >> 4
		v |= v >> 8
		v |= v >> 16
	}
	return v
}
