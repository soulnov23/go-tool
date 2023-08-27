package ringbuffer

import (
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	cacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})
)

type node struct {
	enSeq uint64
	_     [cacheLinePadSize - 8]byte
	deSeq uint64
	_     [cacheLinePadSize - 8]byte
	value any
}

// 为了获得高性能，使用伪共享填充在多线程环境下确保read和write不共享相同的缓存行
type RingBuffer struct {
	capacity uint64
	mask     uint64
	_        [cacheLinePadSize - 16]byte
	head     uint64
	_        [cacheLinePadSize - 8]byte
	tail     uint64
	_        [cacheLinePadSize - 8]byte
	nodes    []*node
}

func New(capacity uint64) *RingBuffer {
	capacity = roundUpToPower2(capacity)
	ring := &RingBuffer{
		capacity: capacity,
		mask:     capacity - 1,
		nodes:    make([]*node, capacity),
	}
	for index := range ring.nodes {
		node := &node{
			enSeq: uint64(index),
			deSeq: uint64(index),
		}
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

func (ring *RingBuffer) Enqueue(value any) {
	for {
		var size uint64
		headPos := atomic.LoadUint64(&ring.head) & ring.mask
		tail := atomic.LoadUint64(&ring.tail)
		tailPos := tail & ring.mask
		if tailPos >= headPos {
			size = tailPos - headPos + 1
		} else {
			// tail已经循环一圈过来了
			size = tailPos + ring.capacity - headPos + 1
		}
		if size >= ring.capacity {
			continue
		}
		// 如果tail已经被其它线程移动了，重新开始
		if tail != atomic.LoadUint64(&ring.tail) {
			continue
		}
		// 抢占pos
		if !atomic.CompareAndSwapUint64(&ring.tail, tail, tail+1) {
			continue
		}
		// 抢到位置后，就没有数据竞争了
		node := ring.nodes[tail&ring.mask]
		enSeq := atomic.LoadUint64(&node.enSeq)
		deSeq := atomic.LoadUint64(&node.deSeq)
		for {
			// 当Dequeue更新ring.head后，还没有更新node.deSeq，这里需要判断是否已经被读取，避免被覆盖
			if enSeq == deSeq {
				node.value = value
				atomic.AddUint64(&node.enSeq, ring.capacity)
				break
			}
		}
		break
	}
}

func (ring *RingBuffer) Dequeue() any {
	for {
		var size uint64
		head := atomic.LoadUint64(&ring.head)
		headPos := head & ring.mask
		tailPos := atomic.LoadUint64(&ring.tail) & ring.mask
		if tailPos >= headPos {
			size = tailPos - headPos + 1
		} else {
			// tail已经循环一圈过来了
			size = tailPos + ring.capacity - headPos + 1
		}
		if size < 1 {
			continue
		}
		// 如果head已经被其它线程移动了，重新开始
		if head != atomic.LoadUint64(&ring.head) {
			continue
		}
		// 抢占pos
		if !atomic.CompareAndSwapUint64(&ring.head, head, head+1) {
			continue
		}
		// 抢到位置后，就没有数据竞争了
		node := ring.nodes[head&ring.mask]
		enSeq := atomic.LoadUint64(&node.enSeq)
		deSeq := atomic.LoadUint64(&node.deSeq)
		for {
			// 当Enqueue更新ring.tail后，还没有更新node.enSeq，这里需要判断是否已经被写入，避免取旧值
			if enSeq == deSeq+ring.capacity {
				value := node.value
				atomic.AddUint64(&node.deSeq, ring.capacity)
				return value
			}
		}
	}
}

// Size 实际大小
func (ring *RingBuffer) Size() uint64 {
	headPos := atomic.LoadUint64(&ring.head) & ring.mask
	tailPos := atomic.LoadUint64(&ring.tail) & ring.mask
	if tailPos >= headPos {
		return tailPos - headPos + 1
	} else {
		// tail已经循环一圈过来了
		return tailPos + ring.capacity - headPos + 1
	}
}

// Capacity 最大容量
func (ring *RingBuffer) Capacity() uint64 {
	return atomic.LoadUint64(&ring.capacity)
}
