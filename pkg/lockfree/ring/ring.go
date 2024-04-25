package ringbuffer

import (
	"errors"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	cacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})
)

type node struct {
	/*----------------CacheLine----------------*/
	enSeq *atomic.Uint64
	_     [cacheLinePadSize - 8]byte
	/*----------------CacheLine----------------*/
	deSeq *atomic.Uint64
	_     [cacheLinePadSize - 8]byte
	/*----------------CacheLine----------------*/
	value any
}

// 为了获得高性能，使用伪共享填充在多线程环境下确保read和write不共享相同的缓存行
type Ring struct {
	/*----------------CacheLine----------------*/
	capacity uint64
	size     *atomic.Uint64
	mask     uint64
	_        [cacheLinePadSize - 24]byte
	/*----------------CacheLine----------------*/
	head *atomic.Uint64
	_    [cacheLinePadSize - 8]byte
	/*----------------CacheLine----------------*/
	tail *atomic.Uint64
	_    [cacheLinePadSize - 8]byte
	/*----------------CacheLine----------------*/
	nodes []*node
}

func New(capacity uint64) *Ring {
	capacity = roundUpToPower2(capacity)
	ring := &Ring{
		capacity: capacity,
		size:     &atomic.Uint64{},
		mask:     capacity - 1,
		head:     &atomic.Uint64{},
		tail:     &atomic.Uint64{},
		nodes:    make([]*node, capacity),
	}
	for index := range ring.nodes {
		node := &node{
			enSeq: &atomic.Uint64{},
			deSeq: &atomic.Uint64{},
		}
		node.enSeq.Store(uint64(index))
		node.deSeq.Store(uint64(index))
		ring.nodes[index] = node
	}
	return ring
}

func roundUpToPower2(v uint64) uint64 {
	if v == 0 {
		return 1
	}
	// 非2的幂
	if v&(v-1) != 0 {
		// 依次将最高位1右边的第1位、第2~3位，第4~7位，第8~15位，第16~31位，第32~63位置为1
		v |= v >> 1
		v |= v >> 2
		v |= v >> 4
		v |= v >> 8
		v |= v >> 16
		v |= v >> 32
		// 进一位，将最右边所有的1都置为0，只保留最高位为1，就是2的幂
		v += 1
	}
	return v
}

func (ring *Ring) Enqueue(value any) error {
	for {
		if ring.Size() == ring.capacity {
			return errors.New("queue is full")
		}
		// 抢占pos
		tail := ring.tail.Load()
		if !ring.tail.CompareAndSwap(tail, tail+1) {
			continue
		}
		// 抢到位置后，就没有数据竞争了
		ring.size.Add(1)
		node := ring.nodes[tail&ring.mask]
		for {
			// 当Dequeue更新ring.head后，还没有更新node.deSeq，这里需要判断是否已经被读取，避免被覆盖
			if node.enSeq.Load() == node.deSeq.Load() {
				node.value = value
				node.enSeq.Add(ring.capacity)
				return nil
			}
		}
	}
}

func (ring *Ring) Dequeue() any {
	for {
		if ring.Size() == 0 {
			return nil
		}
		// 抢占pos
		head := ring.head.Load()
		if !ring.head.CompareAndSwap(head, head+1) {
			continue
		}
		// 抢到位置后，就没有数据竞争了
		ring.size.Add(^uint64(0))
		node := ring.nodes[head&ring.mask]
		for {
			// 当Enqueue更新ring.tail后，还没有更新node.enSeq，这里需要判断是否已经被写入，避免取旧值
			if node.enSeq.Load() == node.deSeq.Load()+ring.capacity {
				value := node.value
				node.deSeq.Add(ring.capacity)
				return value
			}
		}
	}
}

// Size 实际大小
func (ring *Ring) Size() uint64 {
	return ring.size.Load()
}

// Capacity 最大容量
func (ring *Ring) Capacity() uint64 {
	return ring.capacity
}
