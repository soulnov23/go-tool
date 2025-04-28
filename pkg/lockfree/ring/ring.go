// Package ring 实现了基于环形缓冲区的无锁队列
// 该实现通过维护序列号计数器解决ABA问题，提供高性能的并发消息传递机制
package ring

import (
	"errors"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

// 错误定义
var (
	// ErrQueueFull 队列已满错误
	ErrQueueFull = errors.New("queue is full")
	// ErrQueueEmpty 队列为空错误
	ErrQueueEmpty = errors.New("queue is empty")
)

const (
	// 缓存行大小
	cacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})
	// 最小容量
	minCapacity = 2
)

// node 表示队列中的节点
// 环形队列通过维护入队(enSeq)和出队(deSeq)序列号解决ABA问题
// 当一个节点被入队和出队多次时，序列号会保持递增，确保操作安全
// 每个序列号操作增加容量大小的值而不是简单地增加1，可以确保序列号不会在短时间内重复
type node struct {
	/*----------------CacheLine----------------*/
	enSeq *atomic.Uint64             // 入队序列号，每次入队操作+capacity
	_     [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	deSeq *atomic.Uint64             // 出队序列号，每次出队操作+capacity
	_     [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	value any // 节点存储的值
}

// Queue 为了获得高性能，使用缓存行填充在多线程环境下避免伪共享
// 伪共享(False Sharing)指多个CPU核心访问不同变量，但这些变量在同一缓存行，导致缓存失效
// 通过确保关键字段处于不同的缓存行，可以避免缓存一致性流量，提高并发性能
type Queue struct {
	/*----------------CacheLine----------------*/
	capacity uint64                      // 队列容量，必须是2的幂
	size     *atomic.Uint64              // 队列当前大小
	mask     uint64                      // 掩码，用于计算索引(等于capacity-1)
	_        [cacheLinePadSize - 24]byte // 填充至缓存行大小
	/*----------------CacheLine----------------*/
	head *atomic.Uint64             // 队列头部索引
	_    [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	tail *atomic.Uint64             // 队列尾部索引
	_    [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	nodes []*node // 节点数组
}

// New 创建指定容量的环形队列
// 容量会自动调整为大于等于输入值的最小2的幂
// 使用2的幂作为容量可以通过位运算快速计算索引(index & mask)
func New(capacity uint64) *Queue {
	// 确保容量至少为2
	if capacity < minCapacity {
		capacity = minCapacity
	}

	capacity = roundUpToPower2(capacity)
	queue := &Queue{
		capacity: capacity,
		size:     &atomic.Uint64{},
		mask:     capacity - 1,
		head:     &atomic.Uint64{},
		tail:     &atomic.Uint64{},
		nodes:    make([]*node, capacity),
	}
	for index := range queue.nodes {
		node := &node{
			enSeq: &atomic.Uint64{},
			deSeq: &atomic.Uint64{},
		}
		// 初始化时序列号设置为索引位置
		// 这确保了序列号的唯一性并与索引位置对应
		node.enSeq.Store(uint64(index))
		node.deSeq.Store(uint64(index))
		queue.nodes[index] = node
	}
	return queue
}

// roundUpToPower2 将输入值调整为大于等于它的最小2的幂
func roundUpToPower2(v uint64) uint64 {
	if v == 0 {
		return minCapacity
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

// Enqueue 将元素添加到队列尾部
func (queue *Queue) Enqueue(value any) error {
	for {
		if queue.IsFull() {
			return ErrQueueFull
		}

		// 抢占pos
		tail := queue.tail.Load()
		if !queue.tail.CompareAndSwap(tail, tail+1) {
			continue
		}

		// 抢到位置后，就没有数据竞争了
		queue.size.Add(1)
		node := queue.nodes[tail&queue.mask]

		for {
			// 当Dequeue更新ring.head后，还没有更新node.deSeq，这里需要判断是否已经被读取，避免被覆盖
			// 通过比较序列号确保节点状态一致性，防止ABA问题
			if node.enSeq.Load() == node.deSeq.Load() {
				node.value = value
				// 增加一个完整的容量值而不是简单地+1
				// 这确保即使节点被重用多次，序列号也会有显著差异，彻底解决ABA问题
				node.enSeq.Add(queue.capacity)
				return nil
			}
			// 入列失败继续try
		}
	}
}

// Dequeue 从队列头部取出元素
func (queue *Queue) Dequeue() (any, error) {
	for {
		if queue.IsEmpty() {
			return nil, ErrQueueEmpty
		}

		// 抢占pos
		head := queue.head.Load()
		if !queue.head.CompareAndSwap(head, head+1) {
			continue
		}

		// 抢到位置后，就没有数据竞争了
		queue.size.Add(^uint64(0))
		node := queue.nodes[head&queue.mask]

		for {
			// 当Enqueue更新ring.tail后，还没有更新node.enSeq，这里需要判断是否已经被写入，避免取旧值
			// 通过序列号检查确保节点已被写入且未被读取
			// enSeq比deSeq大一个容量值表示节点已写入但未读取
			if node.enSeq.Load() == node.deSeq.Load()+queue.capacity {
				value := node.value
				// 增加一个完整的容量值标记节点已读取
				// 这确保序列号保持较大差异，有效解决ABA问题
				node.deSeq.Add(queue.capacity)
				// 清除引用帮助GC
				node.value = nil
				return value, nil
			}
			// 出列失败继续try
		}
	}
}

// Size 返回队列当前大小
func (queue *Queue) Size() uint64 {
	return queue.size.Load()
}

// Capacity 返回队列最大容量
func (queue *Queue) Capacity() uint64 {
	return queue.capacity
}

// IsEmpty 检查队列是否为空
func (queue *Queue) IsEmpty() bool {
	return queue.Size() == 0
}

// IsFull 检查队列是否已满
func (queue *Queue) IsFull() bool {
	return queue.Size() == queue.capacity
}
